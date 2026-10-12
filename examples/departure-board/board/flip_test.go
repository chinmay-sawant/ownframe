package board_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/departure-board/board"
)

func TestLiverySwitchFillsAccent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := draw(t, ctx)
	clickBox(t, ctx, app, "livery-lcd")

	if got := app.View().Livery; got != "lcd" {
		t.Fatalf("livery = %q, want lcd", got)
	}

	accent, ok := board.LiveryAccent("lcd")
	if !ok {
		t.Fatal("no lcd accent")
	}

	if !hasFill(app, accent) {
		t.Fatal("lcd accent fill not found")
	}
}

func TestFlapsSettle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := draw(t, ctx)

	// Bind the flaps first. Advancing before any tick has no previous glyph
	// to flip from, and every flap would be born settled.
	for i := 0; i < 3; i++ {
		if err := app.Tick(ctx); err != nil {
			t.Fatal(err)
		}
	}

	if err := app.Advance(ctx); err != nil {
		t.Fatal(err)
	}

	// A blank tile paints no text op, so only glyph tiles need to settle.
	want := make(map[string]string)

	for _, row := range app.View().Rows {
		for _, c := range row.Cells {
			if c.Ch != " " {
				want[c.ID] = c.Ch
			}
		}
	}

	done := false
	for i := 0; i < 200 && !done; i++ {
		if err := app.Tick(ctx); err != nil {
			t.Fatal(err)
		}

		done = settled(app, want)
	}

	if !done {
		t.Fatal("flaps did not settle within 200 ticks")
	}
}

func TestNightAndPause(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := draw(t, ctx)
	night := app.View().Night
	clickBox(t, ctx, app, "btn-night")

	if app.View().Night == night {
		t.Fatal("night did not flip")
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if app.Page().Display() == nil {
		t.Fatal("no display after night redraw")
	}

	clickBox(t, ctx, app, "btn-pause")

	if !app.View().Paused {
		t.Fatal("pause did not set")
	}
}
