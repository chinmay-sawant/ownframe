package ui

import (
	"context"
	"testing"
)

func TestThemeTogglePersists(t *testing.T) {
	b := newFake()
	app := newTestApp(t, b)
	settle(t, app)

	light := pixel(t, app, HeadW+6, ChromeH+6)

	app.toggleTheme()
	settle(t, app)
	if !app.dark {
		t.Fatal("dark flag not set")
	}

	if v, _ := b.Pref("theme"); v != "dark" {
		t.Fatalf("stored theme = %q", v)
	}

	dark := pixel(t, app, HeadW+6, ChromeH+6)
	if near(light, dark) {
		t.Fatalf("background did not change: %v", light)
	}

	// A fresh screen over the same backend restores the theme.
	again, err := New(Options{Backend: b, Width: 1000, Height: 700})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = again.Close() }()
	if !again.dark || again.View().Dark != true {
		t.Fatal("theme was not restored")
	}
}

func TestThemeToggleFromToolbarClick(t *testing.T) {
	app := newTestApp(t, newFake())
	settle(t, app)

	var box Box
	for _, b := range app.View().Toolbar {
		if string(b.Action) == "theme" {
			box = b
		}
	}

	if string(box.Action) != "theme" {
		t.Fatal("theme button missing")
	}

	ctx := context.Background()
	if err := app.page.Testing().Click(ctx, float64(box.X+4), float64(box.Y+4)); err != nil {
		t.Fatal(err)
	}

	flush(t, app)
	if !app.dark {
		t.Fatal("click did not toggle the theme")
	}
}
