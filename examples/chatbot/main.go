package main

import (
	"log/slog"
	"os"

	"github.com/kelindar/folio"
	"github.com/kelindar/folio/render"
	"github.com/kelindar/folio/sqlite"
)

func main() {
	// https://lucide.dev/icons/
	reg := folio.NewRegistry()
	folio.Register[*Intent](reg, folio.Options{
		Icon:   "messages-square",
		Title:  "Intent",
		Plural: "Intents",
	})

	db, err := sqlite.Open("file:chatbot.db?_journal_mode=WAL", reg)
	if err != nil {
		panic(err)
	}

	if err := render.ListenAndServe(7000, reg, db); err != nil {
		slog.Error("Failed to start server!", "details", err.Error())
		os.Exit(1)
	}
}
