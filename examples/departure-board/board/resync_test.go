package board

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// TestRedrawResyncsAfterRetargets rotates the feed several times through the
// tick and then asks for a real Redraw: the list must rebuild from the view
// with a text op for every tile.
func TestRedrawResyncsAfterRetargets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	app.Page().Testing().TakeDirtyRects()

	for feed := 1; feed <= 5; feed++ {
		forceRotation(app)

		if err := app.Tick(ctx); err != nil {
			t.Fatal(err)
		}

		for i := 0; i < 200 && !flapsHold(app); i++ {
			if err := app.Tick(ctx); err != nil {
				t.Fatal(err)
			}
		}
	}

	if app.feed != 5 {
		t.Fatalf("feed = %d, want 5", app.feed)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	d := app.Page().Display()
	boxes := app.Page().Boxes()

	for _, row := range app.view.Rows {
		for _, c := range row.Cells {
			b, ok := boxFor(boxes, c.ID)
			if !ok {
				t.Fatalf("no box for %s after the resync redraw", c.ID)
			}

			if op := cellOp(d, b); op == nil || op.Text != c.Ch {
				t.Fatalf("cell %s has no text op %q after the resync redraw", c.ID, c.Ch)
			}
		}
	}
}

// boxFor returns the first hit-test box with id.
func boxFor(boxes []ownframe.Box, id string) (ownframe.Box, bool) {
	for _, b := range boxes {
		if b.ID == id {
			return b, true
		}
	}

	return ownframe.Box{}, false
}
