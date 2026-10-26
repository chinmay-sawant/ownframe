package main

import (
	"context"
	"strings"
	"testing"
)

// TestEditSurfaceClickBeginsEditing checks the code body is the edit target:
// one click enters edit mode, typing reaches the buffer, and a second click
// inside the textarea keeps focus so the caret is not lost.
func TestEditSurfaceClickBeginsEditing(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.openFile("main.go")

	ctx := context.Background()
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	surface, ok := boxByID(app.Page().Boxes(), "edit-surface")
	if !ok {
		t.Fatal("no edit surface over the code body")
	}

	scrollX, scrollY := app.Page().Scroll().ScrollOffset()

	if err := app.Page().Click(ctx, surface.X+surface.W/2+float64(scrollX), surface.Y+surface.H/2+float64(scrollY)); err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if !app.View().Editing {
		t.Fatal("clicking the code body did not enter edit mode")
	}

	if got := app.Page().Form().FocusID(); got != "editor" {
		t.Fatalf("focus after edit-begin = %q, want editor", got)
	}

	if _, ok := boxByID(app.Page().Boxes(), "edit-surface"); ok {
		t.Fatal("edit surface is still painted while editing")
	}

	if err := app.Page().Edit().Type(ctx, "x"); err != nil {
		t.Fatal(err)
	}

	if got := app.Page().Form().FormValue("editor"); !strings.Contains(got, "x") {
		t.Fatalf("typed value = %q, want it to contain x", got)
	}

	if !strings.Contains(app.View().EditText, "x") {
		t.Fatalf("EditText = %q, want it to contain x", app.View().EditText)
	}

	editor, ok := boxByID(app.Page().Boxes(), "editor")
	if !ok {
		t.Fatal("no editor textarea")
	}

	scrollX, scrollY = app.Page().Scroll().ScrollOffset()

	if err := app.Page().Click(ctx, editor.X+editor.W/2+float64(scrollX), editor.Y+editor.H/2+float64(scrollY)); err != nil {
		t.Fatal(err)
	}

	if got := app.Page().Form().FocusID(); got != "editor" {
		t.Fatalf("focus after clicking the textarea = %q, want editor", got)
	}
}
