package main

import (
	"bytes"
	"context"
	"image/png"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestSettingsThemesAndSelection(t *testing.T) {
	a := newTestApp(t)
	ctx := context.Background()
	a.page.SetSize(minWidth, minHeight)
	clickID(t, a, "act-settings")
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	panel, ok := boxByID(a.page.Boxes(), "settings-panel")
	if !ok || panel.X < 0 || panel.Y < 0 || panel.X+panel.W > minWidth || panel.Y+panel.H > minHeight {
		t.Fatalf("settings does not fit: %+v", panel)
	}
	for _, theme := range themeCatalog() {
		clickID(t, a, "theme-"+theme.ID)
		if err := a.page.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		if a.theme != theme.ID || !a.View().SettingsOpen {
			t.Fatalf("theme %s did not apply", theme.ID)
		}
		if _, err := png.Decode(bytes.NewReader(a.page.Probe().PNG())); err != nil {
			t.Fatal(err)
		}
	}
	if err := a.onKeyDown(ctx, "escape"); err != nil {
		t.Fatal(err)
	}
	if a.settingsOpen {
		t.Fatal("escape did not close settings")
	}
}

func TestThemeSessionRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n")
	a, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	a.EnableSession()
	if err := a.onClick(context.Background(), ownframe.Box{Action: "theme:dracula"}); err != nil {
		t.Fatal(err)
	}
	b, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	b.EnableSession()
	if b.theme != "dracula" {
		t.Fatalf("restored theme = %q", b.theme)
	}
	if validTheme("unknown") != "midnight" {
		t.Fatal("unknown theme did not fall back")
	}
}

func TestThemeSelectionPaint(t *testing.T) {
	a := newTestApp(t)
	ctx := context.Background()
	brightness := func(id string) uint32 {
		t.Helper()
		if err := a.onClick(ctx, ownframe.Box{Action: "theme:" + id}); err != nil {
			t.Fatal(err)
		}
		if err := a.page.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		img, err := png.Decode(bytes.NewReader(themePNG(a)))
		if err != nil {
			t.Fatal(err)
		}
		r, g, b, _ := img.At(800, 500).RGBA()
		return r + g + b
	}
	dark, light := brightness("dracula"), brightness("light")
	if light <= dark {
		t.Fatalf("editor background did not change from dark to light: %d, %d", dark, light)
	}
	a.toggleEdit()
	if err := a.onSelectAll(ctx); err != nil {
		t.Fatal(err)
	}
	if a.edit == nil || a.editAnchor != 0 || a.editCaret == 0 {
		t.Fatal("select all did not select the editor text")
	}
}

func themePNG(a *App) []byte { return a.page.Probe().PNG() }
