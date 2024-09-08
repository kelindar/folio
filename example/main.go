package main

import (
	"log/slog"
	"os"

	"github.com/kelindar/folio"
	"github.com/kelindar/folio/example/docs"
	"github.com/kelindar/folio/sqlite"
)

func main() {
	reg := folio.NewRegistry()
	folio.Register[*docs.Person](reg)

	db := sqlite.OpenEphemeral(reg)

	if err := runServer(reg, db); err != nil {
		slog.Error("Failed to start server!", "details", err.Error())
		os.Exit(1)
	}
}
