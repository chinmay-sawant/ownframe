package player

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func mustApp(t testing.TB) *App {
	t.Helper()

	app, _ := newFakeApp(t)

	return app
}

func redraw(t *testing.T, app *App) {
	t.Helper()

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	if app.Page().Probe().Display() == nil {
		t.Fatal("fell back to the bitmap path")
	}
}

func findBox(boxes []ownframe.Box, id string) (ownframe.Box, bool) {
	for _, box := range boxes {
		if box.ID == id {
			return box, true
		}
	}

	return ownframe.Box{}, false
}

func click(t *testing.T, app *App, id string) {
	t.Helper()

	box, ok := findBox(app.Boxes(), id)
	if !ok {
		t.Fatalf("no box %s", id)
	}

	if err := app.Click(context.Background(), box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatalf("click %s: %v", id, err)
	}
}
