package main

import (
	"log/slog"
	"os"

	"github.com/kelindar/ultima-web/internal/docs"
	"github.com/kelindar/ultima-web/internal/object"
	"github.com/kelindar/ultima-web/internal/object/storage"
)

func main() {
	reg := object.NewRegistry()
	object.Register[*docs.Person](reg)

	db := storage.OpenEphemeral(reg)

	if err := runServer(reg, db); err != nil {
		slog.Error("Failed to start server!", "details", err.Error())
		os.Exit(1)
	}
}
