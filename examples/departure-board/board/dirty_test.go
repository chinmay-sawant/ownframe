package board_test

import (
	"context"
	"testing"
)

func TestTickOptsIntoFrameDirty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := draw(t, ctx)
	app.Page().Probe().TakeDirty()

	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if !app.Page().Probe().FrameDirty() {
		t.Fatal("tick did not opt into the frame dirty path")
	}

	if app.Page().Probe().FrameDirty() {
		t.Fatal("FrameDirty did not clear")
	}
}

func TestCascadeDirtyRectsStaySmall(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := draw(t, ctx)
	app.Page().Probe().TakeDirty()

	// Bind the flaps first: an Advance before any tick has no previous
	// glyphs to flip from, so the cascade would never run.
	for i := 0; i < 3; i++ {
		if err := app.Tick(ctx); err != nil {
			t.Fatal(err)
		}

		app.Page().Probe().FrameDirty()
		app.Page().Probe().TakeDirty()
	}

	if err := app.Advance(ctx); err != nil {
		t.Fatal(err)
	}

	app.Page().Probe().TakeDirty()

	d := app.Page().Probe().Display()
	if d == nil {
		t.Fatal("departure board fell back to a bitmap")
	}

	quarter := d.Width * d.Height / 4
	reported, empty := 0, 0
	gen := app.Page().Probe().Generation()

	for i := 0; i < 120; i++ {
		if err := app.Tick(ctx); err != nil {
			t.Fatal(err)
		}

		if !app.Page().Probe().FrameDirty() {
			t.Fatalf("frame %d did not opt into the frame dirty path", i)
		}

		// A generation change is a full Redraw (a feed rotation or the night
		// sheet), which legitimately repaints the board. The area bound is
		// about the flap cascade between those events, so a frame that
		// committed a new display starts a fresh baseline.
		redrawn := app.Page().Probe().Generation() != gen
		gen = app.Page().Probe().Generation()

		rect, ok := app.Page().Probe().TakeDirty()
		if !ok {
			empty++
			continue
		}

		reported++

		if redrawn {
			continue
		}

		if i >= 2 && rect.Dx()*rect.Dy() >= quarter {
			t.Fatalf("frame %d rect %v area %d, want under %d", i, rect, rect.Dx()*rect.Dy(), quarter)
		}
	}

	// The bar repaints at about 10 Hz and the cascade bands last a few frames,
	// so a settled window is mostly blit frames with no rect at all.
	if empty < 20 {
		t.Fatalf("only %d of 120 frames reported no rect, want at least 20", empty)
	}

	if reported < 5 {
		t.Fatalf("only %d frames reported a rect, want at least 5", reported)
	}
}
