package app

import (
	"context"
	"strings"
	"testing"
)

// TestTransformsPage checks the transforms page lists all five rows, keeps
// the display list, and runs a transform from its button.
func TestTransformsPage(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	app.show("transforms")

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	if _, ok := findBox(app.Boxes(), "page-transforms"); !ok {
		t.Fatal("no page-transforms box")
	}

	if got := len(app.View().Transforms.Rows); got != 5 {
		t.Fatalf("transform rows = %d", got)
	}

	run, ok := findBox(app.Boxes(), "tf-run-1")
	if !ok {
		t.Fatal("no tf-run-1 box")
	}

	if run.Action != "tf-run" {
		t.Fatalf("tf-run-1 action = %q", run.Action)
	}

	if app.Page().Display() == nil {
		t.Fatal("transforms fell back to the bitmap path")
	}

	if !tfPaintsText(app, "Turn into bullets") || !tfPaintsText(app, "Rewrite the text below") {
		t.Fatal("transform text missing from the display list")
	}

	clickBox(t, app, run)

	if app.View().Note != "Transform applied" {
		t.Fatalf("note after run = %q", app.View().Note)
	}
}

// tfPaintsText reports whether any display operation paints text that
// contains want.
func tfPaintsText(app *App, want string) bool {
	for _, op := range app.Page().Display().Ops {
		if strings.Contains(op.Text, want) {
			return true
		}
	}

	return false
}
