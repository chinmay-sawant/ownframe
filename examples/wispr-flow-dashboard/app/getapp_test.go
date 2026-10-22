package app

import (
	"context"
	"testing"
)

// TestGetAppWorkflow checks the sidebar button opens the mobile panel, the
// panel keeps the display list, and the close button hides it.
func TestGetAppWorkflow(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	if _, ok := findBox(app.Boxes(), "get-app-modal"); ok {
		t.Fatal("the get-app panel is visible before the click")
	}

	button, ok := findBox(app.Boxes(), "get-app")
	if !ok {
		t.Fatal("no get-app box")
	}

	if button.Action != "get-app" {
		t.Fatalf("get-app action = %q", button.Action)
	}

	clickBox(t, app, button)

	if !app.View().GetApp {
		t.Fatal("GetApp did not open")
	}

	if _, ok := findBox(app.Boxes(), "get-app-modal"); !ok {
		t.Fatal("no get-app-modal box after the click")
	}

	modal, _ := findBox(app.Boxes(), "get-app-modal")
	if modal.X != 0 || modal.W < 1000 || modal.H < 900 {
		t.Fatalf("overlay covers x=%v w=%v h=%v, want the whole frame", modal.X, modal.W, modal.H)
	}

	if app.Page().Display() == nil {
		t.Fatal("the panel fell back to the bitmap path")
	}

	closeBox, ok := findBox(app.Boxes(), "get-app-close")
	if !ok {
		t.Fatal("no get-app-close box")
	}

	clickBox(t, app, closeBox)

	if app.View().GetApp {
		t.Fatal("GetApp did not close")
	}

	button, _ = findBox(app.Boxes(), "get-app")
	clickBox(t, app, button)

	store, ok := findBox(app.Boxes(), "get-app-appstore")
	if !ok {
		t.Fatal("no get-app-appstore box")
	}

	if store.Action != "get-app-store" {
		t.Fatalf("store action = %q", store.Action)
	}

	clickBox(t, app, store)

	if app.View().GetApp {
		t.Fatal("the store click left the panel open")
	}

	if app.View().Note != "Download link copied" {
		t.Fatalf("note after the store click = %q", app.View().Note)
	}
}
