package app

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// findBox returns the first box with the id.
func findBox(boxes []ownframe.Box, id string) (ownframe.Box, bool) {
	for _, box := range boxes {
		if box.ID == id {
			return box, true
		}
	}

	return ownframe.Box{}, false
}

// clickBox clicks the middle of one box.
func clickBox(t *testing.T, app *App, box ownframe.Box) {
	t.Helper()

	if err := app.Click(context.Background(), box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatalf("Click: %v", err)
	}
}

// draw draws a first frame only when the app has no boxes yet. Every
// click already redraws the page, so a second redraw here would lay out the
// whole app again for the same state.
func draw(t *testing.T, app *App) {
	t.Helper()

	if len(app.Boxes()) != 0 {
		return
	}

	if err := app.Redraw(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}
}

// clickID draws if needed, then clicks the box with the id.
func clickID(t *testing.T, app *App, id string) {
	t.Helper()

	draw(t, app)

	box, ok := findBox(app.Boxes(), id)
	if !ok {
		t.Fatalf("no %s box", id)
	}

	clickBox(t, app, box)
}

// openSection draws if needed and clicks one app rail link.
func openSection(t *testing.T, app *App, name string) {
	t.Helper()

	draw(t, app)

	box, ok := findBox(app.Boxes(), "rail-"+name)
	if !ok {
		t.Fatalf("no rail-%s box", name)
	}

	clickBox(t, app, box)
}

// newApp opens the app without drawing a frame.
func newApp(t *testing.T) *App {
	t.Helper()

	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return app
}

// newTestApp opens the app and draws the first frame.
func newTestApp(t *testing.T) *App {
	t.Helper()

	app := newApp(t)

	if err := app.Redraw(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	return app
}
