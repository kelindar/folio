package main

import (
	"log/slog"
	"os"

	"github.com/kelindar/folio"
	"github.com/kelindar/folio/example/docs"
	"github.com/kelindar/folio/render"
	"github.com/kelindar/folio/sqlite"
)

func main() {
	reg := folio.NewRegistry()
	folio.Register[*docs.Person](reg, folio.Options{
		Icon:   "person",
		Title:  "Person",
		Plural: "People",
		Sort:   "1",
	})

	folio.Register[*docs.Company](reg, folio.Options{
		Icon:   "storefront",
		Title:  "Company",
		Plural: "Companies",
		Sort:   "2",
	})

	folio.Register[*docs.Vehicle](reg, folio.Options{
		Icon:   "directions_car",
		Title:  "Vehicle",
		Plural: "Vehicles",
		Sort:   "3",
	})

	db, err := sqlite.Open("file:data.db?_journal_mode=WAL", reg)
	if err != nil {
		panic(err)
	}

	/*for i := 0; i < 100; i++ {
		folio.Insert(db, docs.NewPerson(), "sys")
	}*/

	if err := render.ListenAndServe(7000, reg, db); err != nil {
		slog.Error("Failed to start server!", "details", err.Error())
		os.Exit(1)
	}
}
