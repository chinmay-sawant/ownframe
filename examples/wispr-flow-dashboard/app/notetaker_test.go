package app

import (
	"context"
	"testing"
)

// TestNotetakerPage checks the note rows and the replay path.
func TestNotetakerPage(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	app.show("notetaker")

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	if _, ok := findBox(app.Boxes(), "page-notetaker"); !ok {
		t.Fatal("no notetaker page")
	}

	if _, ok := findBox(app.Boxes(), "note-row-3"); !ok {
		t.Fatal("no note-row-3 box")
	}

	newNote, ok := findBox(app.Boxes(), "note-new")
	if !ok {
		t.Fatal("no note-new box")
	}

	if newNote.Action != "note-new" {
		t.Fatalf("note-new action = %q", newNote.Action)
	}

	if got := len(app.View().Notetaker.Notes); got != 5 {
		t.Fatalf("notes = %d", got)
	}

	if app.Page().Probe().Display() == nil {
		t.Fatal("notetaker fell back to a bitmap")
	}
}
