package app

import (
	"context"
	"testing"
)

// TestSettingsPage checks the settings root, the General controls, and the
// display-list replay path.
func TestSettingsPage(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	app.show("settings")

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	if _, ok := findBox(app.Boxes(), "page-settings"); !ok {
		t.Fatal("no settings page")
	}

	if app.Page().Display() == nil {
		t.Fatal("settings fell back to the bitmap path")
	}

	view := app.View().Settings

	if len(view.Tabs) != 8 || !view.Tabs[0].Active {
		t.Fatalf("tabs = %+v", view.Tabs)
	}

	if len(view.Languages) != 4 || !view.Languages[0].Selected {
		t.Fatalf("languages = %+v", view.Languages)
	}

	if len(view.Toggles) != 4 || !view.Toggles[0].Checked || view.Toggles[3].Checked {
		t.Fatalf("toggles = %+v", view.Toggles)
	}

	for _, id := range []string{"set-language", "set-login", "set-space", "set-save"} {
		if _, ok := findBox(app.Boxes(), id); !ok {
			t.Fatalf("no %s box", id)
		}
	}

	save, _ := findBox(app.Boxes(), "set-save")

	clickBox(t, app, save)

	if app.View().Note != "Settings saved" {
		t.Fatalf("Note = %q", app.View().Note)
	}
}
