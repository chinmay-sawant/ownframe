package app

import (
	"context"
	"testing"
)

// TestResizeReflowsCards checks the grids step the cards down as the frame
// narrows: words under the top row, then apps under streak, then one column.
func TestResizeReflowsCards(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx := context.Background()

	openInsights(t, app)

	order := func(w, h int, ids ...string) {
		t.Helper()

		app.Page().SetSize(w, h)

		if err := app.Page().RedrawIfChanged(ctx); err != nil {
			t.Fatalf("Redraw %dx%d: %v", w, h, err)
		}

		for i := 1; i < len(ids); i++ {
			above, ok := findBox(app.Boxes(), ids[i-1])
			if !ok {
				t.Fatalf("%dx%d: no %s box", w, h, ids[i-1])
			}

			below, ok := findBox(app.Boxes(), ids[i])
			if !ok {
				t.Fatalf("%dx%d: no %s box", w, h, ids[i])
			}

			if below.Y <= above.Y {
				t.Fatalf("%dx%d: %s y=%.0f is not below %s y=%.0f",
					w, h, ids[i], below.Y, ids[i-1], above.Y)
			}
		}
	}

	order(700, 900, "card-wpm", "card-words", "card-apps", "card-streak")
	order(500, 900, "card-wpm", "card-fixes", "card-words", "card-apps", "card-streak")
}
