package board_test

import (
	"context"
	"math"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/examples/departure-board/board"
)

// clickBox clicks the centre of the box with id.
func clickBox(t *testing.T, ctx context.Context, app *board.App, id string) {
	t.Helper()

	for _, b := range app.Boxes() {
		if b.ID == id && b.W > 0 && b.H > 0 {
			if err := app.Click(ctx, b.X+b.W/2, b.Y+b.H/2); err != nil {
				t.Fatal(err)
			}

			return
		}
	}

	t.Fatalf("no box id=%q", id)
}

// hasFill reports whether the display list fills a rectangle with the colour.
func hasFill(app *board.App, c [3]float64) bool {
	d := app.Page().Probe().Display()
	if d == nil {
		return false
	}

	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind == layout.DisplayOpFillRect && math.Abs(op.R-c[0]) < 0.01 && math.Abs(op.G-c[1]) < 0.01 && math.Abs(op.B-c[2]) < 0.01 {
			return true
		}
	}

	return false
}

// draw creates the app and draws it once.
func draw(t *testing.T, ctx context.Context) *board.App {
	t.Helper()

	app, err := board.New()
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	return app
}

func TestNewDrawsCells(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := draw(t, ctx)

	d := app.Page().Probe().Display()
	if d == nil {
		t.Fatal("departure board fell back to a bitmap")
	}

	rows := app.View().Rows
	if len(rows) != 8 {
		t.Fatalf("rows = %d, want 8", len(rows))
	}

	chars := make(map[string]bool)

	for i, row := range rows {
		if len(row.Cells) == 0 {
			t.Fatalf("row %d has no cells", i)
		}

		for _, c := range row.Cells {
			chars[c.Ch] = true
		}
	}

	for i := range d.Ops {
		if d.Ops[i].Kind == layout.DisplayOpText && chars[d.Ops[i].Text] {
			return
		}
	}

	t.Fatal("no text op equals a cell character")
}
