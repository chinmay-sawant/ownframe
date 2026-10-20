package app

import (
	"testing"
)

// openCallsMenu opens the Calls menu on a fresh app.
func openCallsMenu(t *testing.T) *App {
	t.Helper()

	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	openSection(t, app, "calls")

	if app.View().Section != "calls" {
		t.Fatalf("Section = %q", app.View().Section)
	}

	return app
}

// clickCallsBox finds one box by id and clicks it.
func clickCallsBox(t *testing.T, app *App, id string) {
	t.Helper()

	box, ok := findBox(app.Boxes(), id)
	if !ok {
		t.Fatalf("no %s box", id)
	}

	clickBox(t, app, box)
}

// TestCallsKeypadNumber checks keys build the dialed number and Clear empties it.
func TestCallsKeypadNumber(t *testing.T) {
	t.Parallel()
	app := openCallsMenu(t)

	if got := app.View().Calls.Tab; got != "history" {
		t.Fatalf("Tab = %q on open", got)
	}

	clickCallsBox(t, app, "calls-key-1")
	clickCallsBox(t, app, "calls-key-2")

	if got := app.View().Calls.Number; got != "12" {
		t.Fatalf("Number = %q after key 1 and key 2", got)
	}

	clickCallsBox(t, app, "calls-clear")

	if got := app.View().Calls.Number; got != "" {
		t.Fatalf("Number = %q after clear", got)
	}
}

// TestCallsBackspace checks Backspace drops the last dialed character.
func TestCallsBackspace(t *testing.T) {
	t.Parallel()
	app := openCallsMenu(t)

	clickCallsBox(t, app, "calls-key-5")
	clickCallsBox(t, app, "calls-key-star")
	clickCallsBox(t, app, "calls-backspace")

	if got := app.View().Calls.Number; got != "5" {
		t.Fatalf("Number = %q after backspace", got)
	}
}
