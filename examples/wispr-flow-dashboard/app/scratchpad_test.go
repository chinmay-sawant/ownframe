package app

import (
	"context"
	"testing"
)

// TestScratchpadPage checks the note area, the recent list, the replay path,
// and the new-scratchpad control.
func TestScratchpadPage(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	app.show("scratchpad")

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	if _, ok := findBox(app.Boxes(), "page-scratchpad"); !ok {
		t.Fatal("no page-scratchpad box")
	}

	if _, ok := findBox(app.Boxes(), "scratch-note"); !ok {
		t.Fatal("no scratch-note box")
	}

	if got := len(app.View().Scratchpad.Recent); got != 4 {
		t.Fatalf("recent rows = %d", got)
	}

	if app.Page().Display() == nil {
		t.Fatal("scratchpad fell back to the bitmap path")
	}

	if !tfPaintsText(app, "Quick note before standup") {
		t.Fatal("prefilled scratchpad note missing from the display list")
	}

	next, ok := findBox(app.Boxes(), "scratch-new")
	if !ok {
		t.Fatal("no scratch-new box")
	}

	clickBox(t, app, next)

	if app.View().Note != "New scratchpad opened" {
		t.Fatalf("note after new scratchpad = %q", app.View().Note)
	}
}
