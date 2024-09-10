package main

import (
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/kelindar/folio"
	"github.com/kelindar/folio/example/render"
)

//go:embed all:assets
var assets embed.FS

// runServer runs a new HTTP server with the loaded environment variables.
func runServer(registry folio.Registry, db folio.Storage) error {
	const port = 7000

	// Handle static files from the embed FS (with a custom handler).
	http.Handle("GET /assets/", serveStatic(http.FS(assets)))

	// Handle index page view.
	http.Handle("GET /", indexViewHandler(db))

	// Handle API endpoints.
	http.HandleFunc("GET /api/hello-world", showContentAPIHandler)
	http.Handle("GET /view/{urn}", createForm(render.ModeView, db))
	http.Handle("GET /edit/{urn}", createForm(render.ModeEdit, db))
	http.Handle("POST /save/{urn}", saveForm(db))
	http.Handle("POST /search", search(db))

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
