package app

import (
	"context"
	"testing"
)

// TestFreeMonthPage checks the free-month root, the progress data, the step
// rows, and the display-list replay path.
func TestFreeMonthPage(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	app.show("free-month")

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	if _, ok := findBox(app.Boxes(), "page-free-month"); !ok {
		t.Fatal("no free-month page")
	}

	if app.Page().Probe().Display() == nil {
		t.Fatal("free-month fell back to the bitmap path")
	}

	view := app.View().FreeMonth

	if view.Joined != 2 || view.Goal != 3 || view.BarWidth != "66%" {
		t.Fatalf("progress = %+v", view)
	}

	if view.Unlocked {
		t.Fatal("the reward should start locked")
	}

	if len(view.Steps) != 3 || !view.Steps[0].Done || !view.Steps[1].Done || view.Steps[2].Done {
		t.Fatalf("steps = %+v", view.Steps)
	}

	ids := []string{"free-step-invite", "free-step-joined", "free-step-reward", "free-claim"}

	for _, id := range ids {
		if _, ok := findBox(app.Boxes(), id); !ok {
			t.Fatalf("no %s box", id)
		}
	}
}
