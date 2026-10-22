package app

import (
	"context"
	"testing"
)

// TestSnippetsPage checks the page root, the sample rows, and the
// display-list path.
func TestSnippetsPage(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	app.show("snippets")

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	if _, ok := findBox(app.Boxes(), "page-snippets"); !ok {
		t.Fatal("no page-snippets box")
	}

	if _, ok := findBox(app.Boxes(), "snip-add"); !ok {
		t.Fatal("no snip-add box")
	}

	rows := app.View().Snippets.Rows
	if len(rows) != 4 {
		t.Fatalf("snippet rows = %d", len(rows))
	}

	if rows[0].Trigger != ";addr" || rows[0].Uses != "18 times" {
		t.Fatalf("first snippet = %+v", rows[0])
	}

	if app.Page().Testing().Display() == nil {
		t.Fatal("snippets fell back to the bitmap path")
	}
}

// TestSnippetAddLeavesNote checks the New snippet button reports.
func TestSnippetAddLeavesNote(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	app.show("snippets")

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	add, ok := findBox(app.Boxes(), "snip-add")
	if !ok {
		t.Fatal("no snip-add box")
	}

	if add.Action != "snip-add" {
		t.Fatalf("snip-add action = %q", add.Action)
	}

	clickBox(t, app, add)

	if app.View().Note != "Snippet added" {
		t.Fatalf("note = %q", app.View().Note)
	}
}
