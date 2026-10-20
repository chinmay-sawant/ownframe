package app

import (
	"testing"
)

// TestMoreFlyout checks More opens the demo app tiles and a tile click
// leaves a note.
func TestMoreFlyout(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)

	clickID(t, app, "rail-more")

	if app.View().Flyout != "more" {
		t.Fatalf("Flyout = %q", app.View().Flyout)
	}

	if _, ok := findBox(app.Boxes(), "app-word"); !ok {
		t.Fatal("no app-word box")
	}

	clickID(t, app, "app-word")

	if note := app.View().Note; note != "Word is a demo tile" {
		t.Fatalf("note = %q", note)
	}

	clickID(t, app, "rail-more")

	if app.View().Flyout != "" {
		t.Fatalf("Flyout = %q after the second rail-more", app.View().Flyout)
	}
}

// TestScrimClosesFlyout checks a click outside closes the flyout.
func TestScrimClosesFlyout(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)

	clickID(t, app, "rail-more")
	clickID(t, app, "scrim")

	if app.View().Flyout != "" {
		t.Fatalf("Flyout = %q after the scrim click", app.View().Flyout)
	}
}

// TestProfileFlyout checks the profile flyout changes presence and signs out.
func TestProfileFlyout(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)

	clickID(t, app, "me")

	if app.View().Flyout != "profile" {
		t.Fatalf("Flyout = %q", app.View().Flyout)
	}

	clickID(t, app, "status-busy")

	if app.View().Presence != "busy" {
		t.Fatalf("Presence = %q", app.View().Presence)
	}

	clickID(t, app, "signout")

	view := app.View()
	if view.Flyout != "" || view.Note == "" {
		t.Fatalf("after signout: flyout %q note %q", view.Flyout, view.Note)
	}
}
