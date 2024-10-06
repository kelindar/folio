package render

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/angelofallars/htmx-go"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/kelindar/folio"
	"github.com/kelindar/folio/convert"
	"github.com/kelindar/folio/errors"
)

// page handles a page request for a given kind, inferred from path.
func page(registry folio.Registry, db folio.Storage) http.Handler {
	return handle(func(r *http.Request, w *Response) error {
		ctx, err := newContext(ModeView, r, registry, db)
		if err != nil {
			return err
		}

		ns := namespaces(db)
		list, err := renderList(ctx, r, folio.Query{})
		if err != nil {
			return err
		}

		return w.Render(hxLayout(
			fmt.Sprintf("Folio - %s", ctx.Type.Plural),
			contentList(ctx, list, ns),
		))
	})
}

// ---------------------------------- Search and Listing ----------------------------------

func namespace(registry folio.Registry, db folio.Storage) http.Handler {
	return handle(func(r *http.Request, w *Response) error {
		var req struct {
			Namespace string `json:"search_namespace"`
		}

		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return errors.BadRequest("unable to decode request, %v", err)
		}

		rx, err := newContext(ModeView, r, registry, db)
		if err != nil {
			return err
		}

		rx.Namespace = req.Namespace
		list, err := renderList(rx, r, folio.Query{
			Namespace: req.Namespace,
		})
		if err != nil {
			return err
		}
		return w.Render(hxList(rx, list))
	})
}

func search(registry folio.Registry, db folio.Storage) http.Handler {
	return handle(func(r *http.Request, w *Response) error {
		ctx, err := newContext(ModeView, r, registry, db)
		if err != nil {
			return err
		}

		var query folio.Query
		switch r.Method {
		case http.MethodPost:
			var req struct {
				Match     string `json:"search_match"`
				Namespace string `json:"search_namespace"`
			}

			defer r.Body.Close()
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				return errors.BadRequest("unable to decode request, %v", err)
			}

			query.Match = req.Match
			if req.Namespace != "" && req.Namespace != "*" {
				query.Namespace = req.Namespace
			}

			fallthrough
		default:
			list, err := renderList(ctx, r, query)
			if err != nil {
				return err
			}
			return w.Render(list)
		}
	})
}

// pageOf returns the URL for the given page.
func pageOf(kind folio.Kind, query folio.Query, page, size int) string {
	var sb strings.Builder
	sb.WriteString("/search/")
	sb.WriteString(string(kind))
	sb.WriteString("?page=")
	sb.WriteString(strconv.Itoa(page))
	sb.WriteString("&size=")
	sb.WriteString(strconv.Itoa(size))
	if query.Namespace != "" {
		sb.WriteString("&ns=")
		sb.WriteString(query.Namespace)
	}
	if filter := convert.Base64(query.String()); filter != "" {
		sb.WriteString("&filter=")
		sb.WriteString(filter)
	}
	return sb.String()
}

func renderList(rx *Context, r *http.Request, defaultQuery folio.Query) (templ.Component, error) {
	typ, err := rx.Registry.Resolve(folio.Kind(r.PathValue("kind")))
	if err != nil {
		return nil, errors.BadRequest("invalid kind, %v", err)
	}

	page := convert.Int(r.URL.Query().Get("page"), 0)
	size := convert.Int(r.URL.Query().Get("size"), 20)
	text, err := base64.URLEncoding.DecodeString(r.URL.Query().Get("filter"))
	if err != nil {
		return nil, errors.BadRequest("unable to decode query, %v", err)
	}

	query, err := folio.ParseQuery(string(text), nil, defaultQuery)
	if err != nil {
		return nil, errors.BadRequest("unable to parse query, %v", err)
	}

	// Count the number of objects
	count, err := rx.Store.Count(typ.Kind, query)
	if err != nil {
		return nil, errors.Internal("unable to count, %v", err)
	}

	// Update the context query
	query.Limit = size
	query.Offset = page * size
	rx.Query = query

	// Search for the objects
	found, err := rx.Store.Search(typ.Kind, query)
	if err != nil {
		return nil, errors.Internal("unable to search, %v", err)
	}

	return hxListContent(rx, found, page, size, count), nil
}

// ---------------------------------- Object CRUD ----------------------------------

func editObject(mode Mode, registry folio.Registry, db folio.Storage) http.Handler {
	return handle(func(r *http.Request, w *Response) error {
		urn, err := folio.ParseURN(r.PathValue("urn"))
		if err != nil {
			return errors.BadRequest("Unable to decode URN, %v", err)
		}

		typ, err := registry.Resolve(urn.Kind)
		if err != nil {
			return errors.BadRequest("invalid kind, %v", err)
		}

		// Get the person from the database
		document, err := db.Fetch(urn)
		if err != nil {
			return errors.Internal("Unable to fetch object, %v", err)
		}

		return w.Render(hxFormContent(&Context{
			Mode:     mode,
			Kind:     typ.Kind,
			Type:     typ,
			Store:    db,
			Registry: registry,
		}, document))
	})

}

func makeObject(registry folio.Registry, db folio.Storage) http.Handler {
	return handle(func(r *http.Request, w *Response) error {
		rx, err := newContext(ModeCreate, r, registry, db)
		switch {
		case err != nil:
			return errors.BadRequest("invalid request, %v", err)
		case len(rx.Namespace) <= 1:
			return errors.BadRequest("invalid namespace")
		}

		// Create a new instance
		instance, err := folio.NewByType(rx.Type.Type, rx.Namespace)
		if err != nil {
			return errors.Internal("Unable to create object, %v", err)
		}

		return w.Render(hxFormContent(rx, instance))
	})
}

func deleteObject(db folio.Storage) http.Handler {
	return handle(func(r *http.Request, w *Response) error {
		urn, err := folio.ParseURN(r.PathValue("urn"))
		if err != nil {
			return errors.BadRequest("Unable to decode URN, %v", err)
		}

		// Get the latest instance from the database
		if _, err := db.Delete(urn, "sys"); err != nil {
			return errors.Internal("Unable to delete object, %v", err)
		}

		return w.Render(hxListElementDelete(urn))
	})
}

func saveObject(registry folio.Registry, db folio.Storage) http.Handler {
	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")
	validate := validator.New(validator.WithRequiredStructEnabled())
	en_translations.RegisterDefaultTranslations(validate, trans)

	return handle(func(r *http.Request, w *Response) error {
		urn, err := folio.ParseURN(r.PathValue("urn"))
		if err != nil {
			return errors.BadRequest("Unable to decode URN, %v", err)
		}

		// Make sure this kind exists
		typ, err := registry.Resolve(urn.Kind)
		if err != nil {
			return errors.BadRequest("invalid kind, %v", err)
		}

		// Get the latest instance from the database
		instance, err := fetchOrCreate(registry, db, urn)
		if err != nil {
			return errors.Internal("Unable to fetch or create object, %v", err)
		}

		// Hydrate the instance with the new data we've received
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(instance); err != nil {
			return errors.BadRequest("Unable to decode request, %v", err)
		}

		// Validate the input data, and if it's invalid, return the validation errors. We also
		// need to swap the response strategy to none, so that the client doesn't replace the
		// entire form with the validation errors.
		if err := validate.Struct(instance); err != nil {
			return w.RenderWith(
				hxValidationErrors(instance, trans, err.(validator.ValidationErrors)),
				func(r htmx.Response) htmx.Response {
					return r.Reswap(htmx.SwapNone)
				})
		}

		// Save the instance back to the database
		updated, err := folio.Upsert(db, instance, "sys")
		if err != nil {
			return errors.Internal("Unable to save %T, %v", instance, err)
		}

		switch {
		case isCreated(updated):
			return w.Render(hxListElementCreate(&Context{
				Mode:     ModeView,
				Kind:     typ.Kind,
				Type:     typ,
				Store:    db,
				Registry: registry,
			}, updated))
		default:
			return w.Render(hxListElementUpdate(&Context{
				Mode:     ModeView,
				Kind:     typ.Kind,
				Type:     typ,
				Store:    db,
				Registry: registry,
			}, updated))
		}
	})
}

// fetchOrCreate fetches or creates an object from the database.
func fetchOrCreate(registry folio.Registry, db folio.Storage, urn folio.URN) (folio.Object, error) {
	instance, err := db.Fetch(urn)
	if err != nil {
		instance, err = folio.NewByURN(registry, urn)
	}

	return instance, err
}

func isCreated(obj folio.Object) bool {
	_, createdAt := obj.Created()
	_, updatedAt := obj.Updated()
	return createdAt == updatedAt
}

func newContext(mode Mode, r *http.Request, reg folio.Registry, db folio.Storage) (*Context, error) {
	typ, err := reg.Resolve(folio.Kind(r.PathValue("kind")))
	if err != nil {
		return nil, errors.BadRequest("invalid kind, %v", err)
	}

	return &Context{
		Mode:      mode,
		Kind:      typ.Kind,
		Type:      typ,
		Store:     db,
		Registry:  reg,
		Namespace: r.URL.Query().Get("ns"),
	}, nil
}
