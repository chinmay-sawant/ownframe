package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// Settings Appearance offers Dark and Light choices that swap the chadcn
// base sheet; light reuses the .of-scope--light shell overrides.
func TestAppearanceLightDarkSwitch(t *testing.T) {
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "settings:open"}); err != nil {
		t.Fatal(err)
	}
	a.settingsTab = "Appearance"
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	html := a.page.HTML()
	if !strings.Contains(html, "dc-appearance-light") || !strings.Contains(html, "dc-appearance-dark") {
		t.Fatal("appearance options missing")
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "theme:light"}); err != nil {
		t.Fatal(err)
	}
	if a.appearance != dcLight {
		t.Fatalf("appearance = %q, want light", a.appearance)
	}
	if !strings.Contains(a.page.HTML(), "of-scope--light") {
		t.Fatal("light scope class missing")
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "theme:dark"}); err != nil {
		t.Fatal(err)
	}
	if a.appearance != dcDark {
		t.Fatal("dark not restored")
	}
	if strings.Contains(a.page.HTML(), "of-scope--light") {
		t.Fatal("light scope leaked into dark")
	}
}
