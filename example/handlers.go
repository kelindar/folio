package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Define template meta tags.
		metaTags := pages.MetaTags(
			"gowebly, htmx example page, go with htmx",               // define meta keywords
			"Welcome to example! You're here because it worked out.", // define meta description
		)

		people, err := folio.Search[*docs.Person](db, folio.Query{
			Namespaces: []string{"default"},
		})
		if err != nil {
			slog.Error("fetch", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Define template body content.
		bodyContent := pages.BodyContent(&render.Context{
			Mode: render.ModeView,
			Kind: "person",
		}, people)

		// Define template layout for index page.
		indexTemplate := templates.Layout(
			"Welcome to example!", // define title text
			metaTags,              // define meta tags
			bodyContent,           // define body content
		)

		// Render index page template.
		if err := htmx.NewResponse().RenderTempl(r.Context(), w, indexTemplate); err != nil {
			slog.Error("render template", "method", r.Method, "status", http.StatusInternalServerError, "path", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Send log message.
		slog.Info("render page", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
	})
}

func editObject(mode render.Mode, db folio.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urn, err := folio.ParseURN(r.PathValue("urn"))
		if err != nil {
			slog.Error("parse urn", "error", err)
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		// Get the person from the database
		document, err := db.Fetch(urn)
		if err != nil {
			slog.Error("fetch", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if err := htmx.NewResponse().RenderTempl(r.Context(), w, blocks.ListElementEdit(&render.Context{
			Mode: mode,
		}, document.(*docs.Person))); err != nil {
			slog.Error("render template", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Send log message.
		slog.Info("edit object", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
	})

}

func makeObject(registry folio.Registry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rt, err := registry.Resolve(folio.Kind(r.PathValue("kind")))
		if err != nil {
			http.Error(w, "invalid kind", http.StatusBadRequest)
			return
		}

		// Create a new instance
		instance, err := folio.NewByType(rt, "default")
		if err != nil {
			slog.Error("create", "error", err)
			http.Error(w, "unable to create", http.StatusBadRequest)
			return
		}

		if err := htmx.NewResponse().RenderTempl(r.Context(), w, blocks.ListElementEdit(&render.Context{
			Mode: render.ModeCreate,
		}, instance.(*docs.Person))); err != nil {
			slog.Error("render template", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Send log message.
		slog.Info("create object", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
	})
}

func search(db folio.Storage) http.Handler {
	type request struct {
		Query string `json:"query"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req request
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid Data", http.StatusBadRequest)
			return
		}

		people, err := folio.Search[*docs.Person](db, folio.Query{
			Namespaces: []string{"default"},
			Match:      req.Query,
		})
		if err != nil {
			slog.Error("fetch", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if err := htmx.NewResponse().RenderTempl(r.Context(), w, blocks.ListContent(&render.Context{
			Mode: render.ModeView,
		}, people)); err != nil {
			slog.Error("render template", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Send log message.
		slog.Info("search", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
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

func deleteObject(db folio.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urn, err := folio.ParseURN(r.PathValue("urn"))
		if err != nil {
			slog.Error("parse urn", "error", err)
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		// Get the latest instance from the database
		if _, err := db.Delete(urn, "sys"); err != nil {
			slog.Error("fetch", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if err := htmx.NewResponse().RenderTempl(r.Context(), w, blocks.ListElementDelete(urn)); err != nil {
			slog.Error("render template", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})
}

type Response struct {
	hx htmx.Response
	r  *http.Request
	w  http.ResponseWriter
}

// Render renders the given template
func (r *Response) Render(template templ.Component) error {
	if err := r.hx.RenderTempl(r.r.Context(), r.w, template); err != nil {
		return errors.Internal("unable to render template, %w", err)
	}
	return nil
}

// RenderWith renders the given template with custom htmx response
func (r *Response) RenderWith(template templ.Component, fn func(htmx.Response) htmx.Response) error {
	if err := fn(r.hx).RenderTempl(r.r.Context(), r.w, template); err != nil {
		return errors.Internal("unable to render template, %w", err)
	}
	return nil
}

func handle(fn func(r *http.Request, w *Response) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hx := &Response{
			hx: htmx.NewResponse(),
			r:  r,
			w:  w,
		}

		if err := fn(r, hx); err != nil {
			if httpErr := err.(interface {
				HTTP() int
			}); httpErr != nil {
				http.Error(w, err.Error(), httpErr.HTTP())
				return
			}

			slog.Error("error", "error", err)
			http.Error(w, fmt.Sprintf("Internal server error has occured, due to %v", err),
				http.StatusInternalServerError)
		}
	})
}
