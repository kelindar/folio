package render

import (
	"compress/gzip"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/angelofallars/htmx-go"
	"github.com/kelindar/folio"
	"github.com/kelindar/folio/errors"
)

//go:embed all:assets
var assets embed.FS

// ListenAndServe starts the server on the given port.
func ListenAndServe(port int, registry folio.Registry, db folio.Storage) error {
	vd := errors.NewValidator()

	// Handle static files from the embed FS (with a custom handler).
	http.Handle("GET /assets/", serveStatic(http.FS(assets)))

	// Handle page view
	http.Handle("GET /{kind}", page(registry, db))
	http.Handle("POST /content/{kind}", content(registry, db))
	http.Handle("GET /content/{kind}", content(registry, db))

	// Handle API endpoints
	http.Handle("GET /view/{urn}", editObject(ModeView, registry, db))
	http.Handle("GET /edit/{urn}", editObject(ModeEdit, registry, db))
	http.Handle("GET /make/{kind}", makeObject(registry, db))

	// Object CRUD endpoints
	http.Handle("PUT /obj/{urn}", saveObject(registry, db, vd))
	http.Handle("DELETE /obj/{urn}", deleteObject(db))

	// Search and listing endpoints
	http.Handle("GET /search/{kind}", search(registry, db))
	http.Handle("POST /search/{kind}", search(registry, db))

	// Array-based endpoints
	http.Handle("PUT /array/{kind}", addArray(registry, db, vd))

	// Create a new server instance with options from environment variables.
	// For more information, see https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Send log message.
	slog.Info("Starting server...", "port", port)

	return server.ListenAndServe()
}

// serveStatic handles a custom handler for serve embed static folder.
func serveStatic(fs http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fs.Open(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			slog.Error(err.Error(), "method", r.Method, "status", http.StatusNotFound, "path", r.URL.Path)
			return
		}

		// File is found, serve it using the standard http.FileServer.
		http.FileServer(fs).ServeHTTP(w, r)
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
	return withGzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hx := &Response{
			hx: htmx.NewResponse(),
			r:  r,
			w:  w,
		}

		w.Header().Set("Content-Type", "text/html")

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
	}))
}

func withGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.Header.Get("Accept-Encoding"), "gzip"):
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Add("Vary", "Accept-Encoding")
			w.Header().Del("Content-Length")

			gz := gzip.NewWriter(w)
			defer gz.Close()
			next.ServeHTTP(&gzipWriter{Writer: gz, ResponseWriter: w}, r)
		default:
			next.ServeHTTP(w, r)
		}
	})
}

type gzipWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
