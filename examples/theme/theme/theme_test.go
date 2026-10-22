package theme_test

import (
	"context"
	"math"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/examples/theme/theme"
)

// clickBox clicks the centre of the box with id.
func clickBox(t *testing.T, ctx context.Context, app *theme.App, id string) {
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

// hasFill reports whether the display list fills a rectangle with the color.
func hasFill(t *testing.T, app *theme.App, r, g, b float64) bool {
	t.Helper()

	d := app.Page().Display()
	if d == nil {
		t.Fatal("theme example fell back to a bitmap")
	}

	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind != layout.DisplayOpFillRect {
			continue
		}

		if math.Abs(op.R-r) < 0.01 && math.Abs(op.G-g) < 0.01 && math.Abs(op.B-b) < 0.01 {
			return true
		}
	}

	return false
}

func TestToggleSwitchesTheme(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app, err := theme.New()
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if !hasFill(t, app, 26/255.0, 86/255.0, 219/255.0) {
		t.Fatal("light accent fill not found")
	}

	clickBox(t, ctx, app, "toggle")

	if got := app.View().Status; got != "dark theme" {
		t.Fatalf("status = %q, want dark theme", got)
	}

	if !hasFill(t, app, 122/255.0, 162/255.0, 247/255.0) {
		t.Fatal("dark accent fill not found")
	}

	if hasFill(t, app, 26/255.0, 86/255.0, 219/255.0) {
		t.Fatal("light accent fill still present after the toggle")
	}
}
