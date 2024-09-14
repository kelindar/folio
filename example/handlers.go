package main

import (
	"encoding/json"
	"net/http"

	"github.com/angelofallars/htmx-go"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/kelindar/folio"
	"github.com/kelindar/folio/errors"
	"github.com/kelindar/folio/example/docs"
	"github.com/kelindar/folio/example/render"
	"github.com/kelindar/folio/example/templates"
	"github.com/kelindar/folio/example/templates/blocks"
	"github.com/kelindar/folio/example/templates/pages"
)

// indexViewHandler handles a view for the index page.
func indexViewHandler(db folio.Storage) http.Handler {
	return handle(func(r *http.Request, w *Response) error {

		// Define template meta tags.
		metaTags := pages.MetaTags(
			"gowebly, htmx example page, go with htmx",               // define meta keywords
			"Welcome to example! You're here because it worked out.", // define meta description
		)

		people, err := folio.Search[*docs.Person](db, folio.Query{
			Namespaces: []string{"default"},
		})
		if err != nil {
			return errors.Internal("Unable to fetch, %v", err)
		}

		// Define template body content.
		bodyContent := pages.BodyContent(&render.Context{
			Mode: render.ModeView,
			Kind: "person",
		}, people)

		return w.Render(templates.Layout(
			"kelindar/folio", // define title text
			metaTags,         // define meta tags
			bodyContent,      // define body content
		))
	})
}

func editObject(mode render.Mode, db folio.Storage) http.Handler {
	return handle(func(r *http.Request, w *Response) error {
		urn, err := folio.ParseURN(r.PathValue("urn"))
		if err != nil {
			return errors.BadRequest("Unable to decode URN, %v", err)
		}

		// Get the person from the database
		document, err := db.Fetch(urn)
		if err != nil {
			return errors.Internal("Unable to fetch object, %v", err)
		}

		return w.Render(blocks.ListElementEdit(&render.Context{
			Mode: mode,
		}, document))
	})

}

func makeObject(registry folio.Registry) http.Handler {
	return handle(func(r *http.Request, w *Response) error {
		rt, err := registry.Resolve(folio.Kind(r.PathValue("kind")))
		if err != nil {
			return errors.BadRequest("invalid kind, %v", err)
		}

		// Create a new instance
		instance, err := folio.NewByType(rt, "default")
		if err != nil {
			return errors.Internal("Unable to create object, %v", err)
		}

		return w.Render(blocks.ListElementEdit(&render.Context{
			Mode: render.ModeCreate,
		}, instance))
	})
}

func search(db folio.Storage) http.Handler {
	return handle(func(r *http.Request, w *Response) error {
		var req struct {
			Query string `json:"query"`
		}

		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return errors.BadRequest("Unable to decode request, %v", err)
		}

		people, err := folio.Search[*docs.Person](db, folio.Query{
			Namespaces: []string{"default"},
			Match:      req.Query,
		})
		if err != nil {
			return errors.Internal("Unable to search, %v", err)
		}

		return w.Render(blocks.ListContent(&render.Context{
			Mode: render.ModeView,
		}, people))
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

		return w.Render(blocks.ListElementDelete(urn))
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
				blocks.ValidationErrors(instance, trans, err.(validator.ValidationErrors)),
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
			return w.Render(blocks.ListElementCreate(&render.Context{
				Mode: render.ModeView,
			}, updated.(*docs.Person)))
		default:
			return w.Render(blocks.ListElementUpdate(&render.Context{
				Mode: render.ModeView,
			}, updated.(*docs.Person)))
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
