package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// TestEditDirtyFlag checks a caret-only edit paints no status, a real edit
// says "Changes made" and dots the tab, and save clears both.
func TestEditDirtyFlag(t *testing.T) {
	t.Run("caret", func(t *testing.T) {
		app := newTestApp(t)
		if app.View().Active != "README.md" {
			t.Fatalf("active = %q, want README.md", app.View().Active)
		}

		enterEditAtFirstLine(t, app)
		assertCleanEditFlag(t, app)
	})

	t.Run("save", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "README.md", "# Notes\n\nHello.\n")

		app, err := New(dir)
		if err != nil {
			t.Fatal(err)
		}

		ctx := context.Background()
		if err := app.Page().Redraw(ctx); err != nil {
			t.Fatal(err)
		}

		enterEditAtFirstLine(t, app)
		assertCleanEditFlag(t, app)

		if err := app.Page().Edit().Type(ctx, "x"); err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(app.edit.Text(), "x") {
			t.Fatalf("buffer = %q, want the typed character", app.edit.Text())
		}

		flag, ok := boxByID(app.Page().Boxes(), "edit-flag")
		if !ok {
			t.Fatal("dirty buffer did not render the edit flag")
		}

		if flag.Text != "Changes made" {
			t.Fatalf("edit flag = %q, want %q", flag.Text, "Changes made")
		}

		dotID := tabDotID(t, app)
		dot, ok := boxByID(app.Page().Boxes(), dotID)
		if !ok || dot.W < 1 || dot.H < 1 {
			t.Fatalf("dirty tab marker %s = %+v", dotID, dot)
		}

		if !activeEditorTab(t, app).Dirty {
			t.Fatal("tab row is not marked dirty")
		}

		assertEditButtons(t, app)

		if _, ok := boxByID(app.Page().Boxes(), "edit-save"); !ok {
			t.Fatal("save button missing")
		}

		if err := app.onClick(ctx, ownframe.Box{Action: "edit-save"}); err != nil {
			t.Fatal(err)
		}

		if err := app.Page().Redraw(ctx); err != nil {
			t.Fatal(err)
		}

		assertCleanEditFlag(t, app)
	})
}

func enterEditAtFirstLine(t *testing.T, app *App) {
	t.Helper()

	if err := app.onClick(context.Background(), ownframe.Box{Action: "edit-begin-line:0"}); err != nil {
		t.Fatal(err)
	}

	if !app.View().Editing {
		t.Fatal("click did not enter edit mode")
	}
}

func assertCleanEditFlag(t *testing.T, app *App) {
	t.Helper()

	if flag, ok := boxByID(app.Page().Boxes(), "edit-flag"); ok && strings.TrimSpace(flag.Text) != "" {
		t.Fatalf("clean edit flag = %q, want none", flag.Text)
	}

	dotID := tabDotID(t, app)
	if _, ok := boxByID(app.Page().Boxes(), dotID); ok {
		t.Fatalf("clean tab %s shows a dirty marker", dotID)
	}

	if activeEditorTab(t, app).Dirty {
		t.Fatal("clean tab row is marked dirty")
	}

	assertEditButtons(t, app)
}

func assertEditButtons(t *testing.T, app *App) {
	t.Helper()

	if _, ok := boxByID(app.Page().Boxes(), "edit-save"); !ok {
		t.Fatal("save button missing")
	}

	if _, ok := boxByID(app.Page().Boxes(), "edit-toggle"); !ok {
		t.Fatal("done button missing")
	}
}

func activeEditorTab(t *testing.T, app *App) tabRow {
	t.Helper()

	tabs := app.View().Tabs
	index := app.View().ActiveTab
	if index < 0 || index >= len(tabs) {
		t.Fatalf("active tab %d of %d", index, len(tabs))
	}

	return tabs[index]
}

func tabDotID(t *testing.T, app *App) string {
	t.Helper()

	_ = activeEditorTab(t, app)

	return fmt.Sprintf("tab-dot-%d", app.View().ActiveTab)
}
