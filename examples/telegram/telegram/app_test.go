package telegram_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/telegram/telegram"
)

func newApp(t *testing.T, ctx context.Context) *telegram.App {
	t.Helper()

	app, err := telegram.New()
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	return app
}

// center uses the last matching box so a nested element wins over its parent.
func center(t *testing.T, app *telegram.App, id, action string) (float64, float64) {
	t.Helper()

	var x, y float64
	found := false

	for _, b := range app.Boxes() {
		if id != "" && b.ID != id {
			continue
		}

		if action != "" && b.Action != action {
			continue
		}

		if b.W <= 0 || b.H <= 0 {
			continue
		}

		x = b.X + b.W/2
		y = b.Y + b.H/2
		found = true
	}

	if !found {
		t.Fatalf("no box id=%q action=%q:%s", id, action, dumpBoxes(app))
	}

	return x, y
}

func click(t *testing.T, ctx context.Context, app *telegram.App, id, action string) {
	t.Helper()

	x, y := center(t, app, id, action)
	if err := app.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}
}
