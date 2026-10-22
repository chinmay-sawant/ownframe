package console_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/ops-console/console"
)

func newApp(t *testing.T, ctx context.Context) *console.App {
	t.Helper()

	app, err := console.New()
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if len(app.Boxes()) == 0 {
		t.Fatal("no boxes")
	}

	return app
}

func click(t *testing.T, ctx context.Context, app *console.App, id, action string) {
	t.Helper()

	x, y := center(t, app, id, action)
	if err := app.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}
}

func center(t *testing.T, app *console.App, id, action string) (float64, float64) {
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

func dumpBoxes(app *console.App) string {
	var b strings.Builder

	for _, box := range app.Boxes() {
		fmt.Fprintf(
			&b,
			"\n%s id=%q action=%q text=%q",
			box.Tag,
			box.ID,
			box.Action,
			box.Text,
		)
	}

	if b.Len() == 0 {
		return "\n(no boxes)"
	}

	return b.String()
}

func boxByID(t *testing.T, app *console.App, id string) ownframe.Box {
	t.Helper()

	for _, box := range app.Boxes() {
		if box.ID == id {
			return box
		}
	}

	t.Fatalf("no box id=%q:%s", id, dumpBoxes(app))

	return ownframe.Box{}
}
