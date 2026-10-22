package app

import (
	"context"
	"testing"
)

// TestHelpPage checks the help root, the search row, and the sample counts.
func TestHelpPage(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	app.show("help")

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	if _, ok := findBox(app.Boxes(), "page-help"); !ok {
		t.Fatal("no page-help box")
	}

	if _, ok := findBox(app.Boxes(), "help-search"); !ok {
		t.Fatal("no help-search box")
	}

	if _, ok := findBox(app.Boxes(), "help-popular"); !ok {
		t.Fatal("no help-popular box")
	}

	if app.Page().Probe().Display() == nil {
		t.Fatal("the help page fell back to the bitmap path")
	}

	help := DefaultView().Help

	if len(help.Topics) != 6 {
		t.Fatalf("topics = %d", len(help.Topics))
	}

	if len(help.Articles) != 5 {
		t.Fatalf("articles = %d", len(help.Articles))
	}
}
