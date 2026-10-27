package main

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/chinmay-sawant/ownframe"
)

// selDragCode is short enough that the wrap plan maps one visual row to each
// source line, so a drag row offset is also a source line offset.
const selDragCode = "package main\n\nvar Greeting = \"ownframe drag selection\"\n\nfunc main() {\n\tprintln(Greeting)\n}\n"

// The drag starts on the first rune of line 0 and ends on selDragEndCol of
// selDragEndRow.
const (
	selDragEndRow = 2
	selDragEndCol = 10
)

// TestEditSelectionDrag covers the mouse selection path in the custom editor:
// a press inside the code body anchors it, a drag extends the caret, Ctrl+A
// takes the whole buffer, and Ctrl+C returns the selected text.
func TestEditSelectionDrag(t *testing.T) {
	ctx := context.Background()
	app, body := selDragEditor(t)

	anchorX := body.X + editTextLeft
	anchorY := body.Y + float64(rowH)/halfDivisor

	if _, err := app.onPress(ctx, body, anchorX, anchorY); err != nil {
		t.Fatalf("press: %v", err)
	}

	if !app.dragSelect {
		t.Fatal("a press inside the code body did not start a drag selection")
	}

	if app.editAnchor != app.editCaret {
		t.Fatalf("press anchor=%d caret=%d, want a collapsed anchor", app.editAnchor, app.editCaret)
	}

	targetX := body.X + editTextLeft + float64(selDragEndCol)*codeCellW
	targetY := body.Y + float64(selDragEndRow)*rowH + float64(rowH)/halfDivisor

	if err := app.onDragMove(ctx, targetX-anchorX, targetY-anchorY); err != nil {
		t.Fatalf("drag move: %v", err)
	}

	if app.editAnchor == app.editCaret {
		t.Fatalf("drag left anchor=%d caret=%d, want a range", app.editAnchor, app.editCaret)
	}

	if app.editAnchor != 0 {
		t.Fatalf("drag anchor=%d, want the press point at offset 0", app.editAnchor)
	}

	copied, ok, err := app.onCopy(ctx)
	if err != nil {
		t.Fatalf("copy after drag: %v", err)
	}

	if !ok || copied == "" {
		t.Fatalf("copy after drag = %q ok=%v, want selected text", copied, ok)
	}

	if want := selDragPrefix(app.edit.Text()); copied != want {
		t.Fatalf("copy after drag = %q, want %q", copied, want)
	}

	if !hasEditSel(app.View().Rows) {
		t.Fatalf("no edit-sel segment in %d visible rows", len(app.View().Rows))
	}

	t.Logf("drag selection = %d runes %q", utf8.RuneCountInString(copied), copied)

	if err := app.onSelectAll(ctx); err != nil {
		t.Fatalf("select all: %v", err)
	}

	if app.editAnchor != 0 || app.editCaret != utf8.RuneCountInString(app.edit.Text()) {
		t.Fatalf("select all anchor=%d caret=%d, want 0..%d",
			app.editAnchor, app.editCaret, utf8.RuneCountInString(app.edit.Text()))
	}

	all, ok, err := app.onCopy(ctx)
	if err != nil {
		t.Fatalf("copy after select all: %v", err)
	}

	if !ok {
		t.Fatal("copy after select all reported no selection")
	}

	if all != app.edit.Text() {
		t.Fatalf("copy after select all = %d runes, buffer has %d",
			utf8.RuneCountInString(all), utf8.RuneCountInString(app.edit.Text()))
	}

	if !hasEditSel(app.View().Rows) {
		t.Fatal("select all left no edit-sel segment in the visible rows")
	}

	t.Logf("select all = %d runes, drag = %d runes",
		utf8.RuneCountInString(all), utf8.RuneCountInString(copied))
}

// selDragEditor writes the fixture, opens it, and enters edit mode through the
// edit surface the way a pointer click does.
func selDragEditor(t *testing.T) (*App, ownframe.Box) {
	t.Helper()

	dir := t.TempDir()
	writeFile(t, dir, "main.go", selDragCode)

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	app.openFile("main.go")

	ctx := context.Background()

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("redraw: %v", err)
	}

	surface, ok := boxByID(app.Page().Boxes(), "edit-surface")
	if !ok {
		t.Fatal("no edit surface over the code body")
	}

	scrollX, scrollY := app.Page().Scroll().ScrollOffset()

	if err := app.Page().Click(ctx, surface.X+surface.W/2+float64(scrollX), surface.Y+surface.H/2+float64(scrollY)); err != nil {
		t.Fatalf("click edit surface: %v", err)
	}

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("redraw after edit: %v", err)
	}

	if !app.View().Editing || app.edit == nil {
		t.Fatal("clicking the edit surface did not enter edit mode")
	}

	if app.edit.Text() != selDragCode {
		t.Fatalf("edit buffer = %q, want the fixture", app.edit.Text())
	}

	body, ok := boxByID(app.Page().Boxes(), "code-body")
	if !ok {
		t.Fatal("no code body while editing")
	}

	return app, body
}

// selDragPrefix returns the buffer prefix the drag covers. The anchor is the
// first rune of line 0, so the range ends at selDragEndCol on selDragEndRow.
func selDragPrefix(text string) string {
	offset := 0

	for line, part := range strings.Split(text, "\n") {
		length := utf8.RuneCountInString(part)
		if line == selDragEndRow {
			return string([]rune(text)[:offset+min(selDragEndCol, length)])
		}

		offset += length + 1
	}

	return text
}

// hasEditSel reports whether any visible row carries the selection class the
// edit view paints.
func hasEditSel(rows []codeRow) bool {
	for _, row := range rows {
		for _, seg := range row.Segments {
			for _, class := range strings.Fields(seg.Class) {
				if class == "edit-sel" {
					return true
				}
			}
		}
	}

	return false
}
