package board

import (
	"context"
	"testing"
	"time"
)

// runTick runs one frame tick and consumes what it marked, the way the window
// does.
func runTick(t *testing.T, ctx context.Context, app *App) {
	t.Helper()

	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	app.Page().FrameDirty()
	app.Page().TakeDirtyRects()
}

// forceRotation makes the next tick rotate the feed.
func forceRotation(app *App) { app.lastRotate = time.Now().Add(-rotateEvery) }

// flapsHold reports whether every bound flap shows its target glyph.
func flapsHold(app *App) bool {
	for _, fl := range app.flip.cells {
		if fl.op.Text != fl.want {
			return false
		}
	}

	return true
}

// TestRotationRetargets checks a due rotation moves the view and the bound
// flaps without another Redraw.
func TestRotationRetargets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		runTick(t, ctx, app)
	}

	gen := app.Page().Generation()
	forceRotation(app)

	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if got := app.Page().Generation(); got != gen {
		t.Fatalf("rotation redrew the page: generation %d -> %d", gen, got)
	}

	if app.feed != 1 {
		t.Fatalf("feed = %d, want 1", app.feed)
	}

	view := make(map[string]string)
	for _, row := range app.view.Rows {
		for _, c := range row.Cells {
			view[c.ID] = c.Ch
		}
	}

	changed := 0
	for id, fl := range app.flip.cells {
		ch, ok := view[id]
		if !ok {
			continue
		}

		if fl.want != ch {
			t.Fatalf("flap %s want %q, view holds %q", id, fl.want, ch)
		}

		if fl.op.Text != ch {
			changed++
		}
	}

	if changed == 0 {
		t.Fatal("rotation changed no tile")
	}

	for i := 0; i < 200 && !flapsHold(app); i++ {
		if err := app.Tick(ctx); err != nil {
			t.Fatal(err)
		}
	}

	if !flapsHold(app) {
		t.Fatal("retargeted flaps did not settle within 200 ticks")
	}
}
