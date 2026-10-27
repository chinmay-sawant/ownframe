package main

import (
	"context"
	"strings"
	"testing"
)

// TestEditSurfaceClickBeginsEditing checks the code body is the edit target:
// one click enters the highlighted edit view, typing reaches the buffer, and
// the rendered rows keep their syntax spans.
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

	if _, ok := boxByID(app.Page().Boxes(), "edit-surface"); ok {
		t.Fatal("edit surface is still painted while editing")
	}

	// The highlighted rows stay live while editing. A keyword span must
	// still be present, not a plain textarea.
	foundKeyword := false

	for _, row := range app.View().Rows {
		for _, seg := range row.Segments {
			if strings.Contains(seg.Class, "kw") {
				foundKeyword = true
			}
		}
	}

	if !foundKeyword {
		for _, row := range app.View().Rows {
			for _, seg := range row.Segments {
				t.Logf("line %d class=%q text=%q", row.Line, seg.Class, seg.Text)
			}
		}

		t.Fatal("edit mode lost the syntax spans")
	}

	if err := app.onKeyDown(ctx, "x"); err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Edit().Type(ctx, "x"); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(app.edit.Text(), "x") {
		t.Fatalf("buffer = %q, want it to contain x", app.edit.Text())
	}

	if !strings.Contains(app.View().EditText, "x") {
		t.Fatalf("EditText = %q, want it to contain x", app.View().EditText)
	}

	if !app.View().EditDirty {
		t.Fatal("typing did not mark the buffer dirty")
	}
}
