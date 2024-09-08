package main

import (
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	gowebly "github.com/gowebly/helpers"
	"github.com/kelindar/ultima-web/internal/object"
	"github.com/kelindar/ultima-web/internal/object/storage"
	"github.com/kelindar/ultima-web/internal/render"
)

//go:embed all:assets
var assets embed.FS

// runServer runs a new HTTP server with the loaded environment variables.
func runServer(registry object.Registry, db storage.Storage) error {
	// Validate environment variables.
	port, err := strconv.Atoi(gowebly.Getenv("BACKEND_PORT", "7000"))
	if err != nil {
		return err
	}

	// Handle static files from the embed FS (with a custom handler).
	http.Handle("GET /assets/", gowebly.StaticFileServerHandler(http.FS(assets)))

	// Handle index page view.
	http.Handle("GET /", indexViewHandler(db))

	// Handle API endpoints.
	http.HandleFunc("GET /api/hello-world", showContentAPIHandler)
	http.Handle("GET /view/{urn}", createForm(render.ModeView, db))
	http.Handle("GET /edit/{urn}", createForm(render.ModeEdit, db))
	http.Handle("POST /save/{urn}", saveForm(registry, db))

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
