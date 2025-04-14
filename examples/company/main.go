package main

import (
	"log/slog"
	"os"

	"github.com/kelindar/folio"
	"github.com/kelindar/folio/render"
	"github.com/kelindar/folio/sqlite"
)

func main() {
	reg := folio.NewRegistry()
	folio.Register[*Person](reg, folio.Options{
		Icon:   "user-round",
		Title:  "Person",
		Plural: "People",
		Sort:   "1",
	})

	folio.Register[*Company](reg, folio.Options{
		Icon:   "store",
		Title:  "Company",
		Plural: "Companies",
		Sort:   "2",
	})

	folio.Register[*Vehicle](reg, folio.Options{
		Icon:   "car",
		Title:  "Vehicle",
		Plural: "Vehicles",
		Sort:   "3",
	})

	db, err := sqlite.Open("file:data.db?_journal_mode=WAL", reg)
	if err != nil {
		panic(err)
	}

	if err := render.ListenAndServe(7000, reg, db); err != nil {
		slog.Error("Failed to start server!", "details", err.Error())
		os.Exit(1)
	}
}
