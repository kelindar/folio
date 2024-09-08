package main

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/angelofallars/htmx-go"
	"github.com/kelindar/ultima-web/internal/docs"
	"github.com/kelindar/ultima-web/internal/object"
	"github.com/kelindar/ultima-web/internal/object/storage"
	"github.com/kelindar/ultima-web/internal/render"
	"github.com/kelindar/ultima-web/templates"
	"github.com/kelindar/ultima-web/templates/blocks"
	"github.com/kelindar/ultima-web/templates/pages"
)

// indexViewHandler handles a view for the index page.
func indexViewHandler(db storage.Storage) http.Handler {
	for i := 0; i < 100; i++ {
		storage.Insert(db, docs.NewPerson(), "sys")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Define template meta tags.
		metaTags := pages.MetaTags(
			"gowebly, htmx example page, go with htmx",               // define meta keywords
			"Welcome to example! You're here because it worked out.", // define meta description
		)

		people, err := storage.Search[*docs.Person](db, storage.Query{
			Namespaces: []string{"default"},
		})
		if err != nil {
			slog.Error("fetch", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Define template body content.
		bodyContent := pages.BodyContent(
			people,
		)

		// Define template layout for index page.
		indexTemplate := templates.Layout(
			"Welcome to example!", // define title text
			metaTags,              // define meta tags
			bodyContent,           // define body content
		)

		// Render index page template.
		if err := htmx.NewResponse().RenderTempl(r.Context(), w, indexTemplate); err != nil {
			// If not, return HTTP 400 error.
			slog.Error("render template", "method", r.Method, "status", http.StatusInternalServerError, "path", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Send log message.
		slog.Info("render page", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
	})
}

// showContentAPIHandler handles an API endpoint to show content.
func showContentAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Check, if the current request has a 'HX-Request' header.
	// For more information, see https://htmx.org/docs/#request-headers
	if !htmx.IsHTMX(r) {
		// If not, return HTTP 400 error.
		slog.Error("request API", "method", r.Method, "status", http.StatusBadRequest, "path", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Write HTML content.
	w.Write([]byte("<p>🎉 Yes, <strong>htmx</strong> is ready to use! (<code>GET /api/hello-world</code>)</p>"))

	// Send htmx response.
	htmx.NewResponse().Write(w)

	// Send log message.
	slog.Info("request API", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
}

func createForm(mode render.Mode, db storage.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urn, err := object.ParseURN(r.PathValue("urn"))
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

		ctx := &render.Context{
			Mode: mode,
		}

		if err := htmx.NewResponse().RenderTempl(r.Context(), w, blocks.PersonEdit(ctx, document.(*docs.Person))); err != nil {
			slog.Error("render template", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Send log message.
		slog.Info("render person", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
	})

}

func saveForm(registry object.Registry, db storage.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urn, err := object.ParseURN(r.PathValue("urn"))
		if err != nil {
			slog.Error("parse urn", "error", err)
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		// Get the latest instance from the database
		instance, err := db.Fetch(urn)
		if err != nil {
			slog.Error("fetch", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Hydrate the instance with the new data we've received
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(instance); err != nil {
			http.Error(w, "Invalid Data", http.StatusBadRequest)
			return
		}

		// Save the instance back to the database
		output, err := storage.Update(db, instance, "sys")
		if err != nil {
			slog.Error("update", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Render the updated instance
		if err := htmx.NewResponse().RenderTempl(r.Context(), w, blocks.PersonEdit(&render.Context{
			Mode: render.ModeView,
		}, output.(*docs.Person))); err != nil {
			slog.Error("render template", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Send log message.
		slog.Info("render person", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
	})
}
