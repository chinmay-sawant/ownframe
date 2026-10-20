package app

import (
	"testing"
)

// TestThemeDefaultsToDark checks the app opens in the dark theme.
func TestThemeDefaultsToDark(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)

	if !app.View().Dark {
		t.Fatal("Dark = false on open")
	}

	if _, ok := findBox(app.Boxes(), "theme-dark"); ok {
		t.Fatal("theme chips drawn before the profile flyout opens")
	}
}

// TestThemeToggle checks the profile flyout switches the theme.
func TestThemeToggle(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)

	clickID(t, app, "me")
	clickID(t, app, "theme-light")

	if app.View().Dark {
		t.Fatal("Dark = true after the light click")
	}

	if _, ok := findBox(app.Boxes(), "theme-light"); !ok {
		t.Fatal("no theme-light box after the light click")
	}

	clickID(t, app, "theme-dark")

	if !app.View().Dark {
		t.Fatal("Dark = false after the dark click")
	}
}
