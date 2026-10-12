package board_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/departure-board/board"
	"github.com/chinmay-sawant/ownframe/internal/frame"
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

	for i := 0; i < 60; i++ {
		if err := app.Tick(ctx); err != nil {
			t.Fatal(err)
		}
	}

	d := app.Page().Display()
	if d == nil {
		t.Fatal("departure board fell back to a bitmap")
	}

	for id, ch := range want {
		found := false

		for _, b := range app.Boxes() {
			if b.ID != id {
				continue
			}

			found = true

			if op := frame.Text(d, b); op == nil || op.Text != ch {
				t.Fatalf("%s = %+v, want %q", id, op, ch)
			}
		}

		if !found {
			t.Fatalf("no box id=%q", id)
		}
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
