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

	db, err := sqlite.Open("file:data.db?_journal_mode=WAL", reg)
	if err != nil {
		panic(err)
	}

	/*for i := 0; i < 100; i++ {
		folio.Insert(db, docs.NewPerson(), "sys")
	}*/

	if err := runServer(reg, db); err != nil {
		slog.Error("Failed to start server!", "details", err.Error())
		os.Exit(1)
	}
}
