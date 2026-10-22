package states_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/states/states"
)

func newApp(t *testing.T, ctx context.Context) *states.App {
	t.Helper()

	app, err := states.New()
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	return app
}

// assertReplayed fails when the last redraw kept a bitmap instead of a
// display list, which is what shows the bitmap fallback badge in the window.
func assertReplayed(t *testing.T, app *states.App) {
	t.Helper()

	if app.Page().Display() == nil {
		t.Fatal("page fell back to the bitmap")
	}
}

func clickBox(t *testing.T, ctx context.Context, app *states.App, id string) {
	t.Helper()

	x, y := boxCenter(t, app, id)
	if err := app.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}
}

func boxCenter(t *testing.T, app *states.App, id string) (float64, float64) {
	t.Helper()

	var x, y float64

	found := false

	for _, b := range app.Boxes() {
		if b.ID != id || b.W <= 0 || b.H <= 0 {
			continue
		}

		x = b.X + b.W/2
		y = b.Y + b.H/2
		found = true
	}

	if !found {
		t.Fatalf("no box id=%q", id)
	}

	return x, y
}
