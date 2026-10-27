package main

import (
	"context"
	"testing"
	"unicode/utf8"
)

// editCaretCode keeps every line short enough for one visual row, so a source
// line and a visual row are the same offset from the top of the code body.
const editCaretCode = "package main\n\nfunc main() {\n\tgreeting := \"hi\"\n\tprintln(greeting)\n}\n"

// editCaretLine is the zero-based source line the press targets: the third
// line, "func main() {". editCaretCol sits inside "main", past column zero.
const (
	editCaretLine = 2
	editCaretCol  = 6
)

// TestEditCaretClickTypesAndMoves is the regression test for the custom
// editor's caret: a press inside the code body places the caret on the clicked
// glyph, the painted row carries the caret at that column, typing lands at the
// caret, and the arrow keys move it one rune while the caret row follows.
func TestEditCaretClickTypesAndMoves(t *testing.T) {
	ctx := context.Background()
	app := editCaretEditor(t, ctx)

	body, ok := boxByID(app.Page().Boxes(), "code-body")
	if !ok {
		t.Fatal("no code body while editing")
	}

	// onPress takes a viewport point while Boxes reports laid-out page points,
	// so undo the page scroll for the y coordinate. The fixture is narrower
	// than the viewport, so the horizontal offset is zero.
	scrollX, scrollY := app.Page().Scroll().ScrollOffset()
	if scrollX != 0 {
		t.Fatalf("horizontal scroll = %d, want 0 for the short fixture", scrollX)
	}

	x := body.X + editTextLeft + float64(editCaretCol)*codeCellW
	y := body.Y + float64(editCaretLine)*rowH + float64(rowH)/2 - float64(scrollY)

	if _, err := app.onPress(ctx, body, x, y); err != nil {
		t.Fatalf("press at (%.1f,%.1f): %v", x, y, err)
	}

	caretLine, caretCol := offsetLineCol(app.edit.Text(), app.editCaret)
	if caretLine != editCaretLine || caretCol != editCaretCol {
		t.Fatalf("press mapped to line %d col %d, want line %d col %d",
			caretLine, caretCol, editCaretLine, editCaretCol)
	}

	if app.editAnchor != app.editCaret {
		t.Fatalf("press anchor=%d caret=%d, want a collapsed caret", app.editAnchor, app.editCaret)
	}

	t.Logf("press lands on line %d col %d (offset %d)", caretLine, caretCol, app.editCaret)

	// onPress publishes state only; the drag path refreshes on movement. The
	// rows below are what the next frame paints from that state.
	editCaretRedraw(t, app, ctx)

	row := editCaretRow(t, app.View().Rows, editCaretLine, editCaretCol)
	if !row.Caret {
		t.Fatalf("row on line %d has no caret", editCaretLine)
	}

	if row.CaretLeft <= 0 {
		t.Fatalf("caret left = %d on line %d, want it past column zero", row.CaretLeft, editCaretLine)
	}

	if want := int(float64(caretCol) * codeCellW); row.CaretLeft != want {
		t.Fatalf("caret left = %d, want %d for col %d", row.CaretLeft, want, editCaretCol)
	}

	// Type at the caret and check the buffer grew there, not at the end.
	before := app.edit.Text()
	beforeRunes := utf8.RuneCountInString(before)
	start := app.editCaret

	const typed = "abcdefgh"

	typedRunes := utf8.RuneCountInString(typed)

	if err := app.onType(ctx, typed); err != nil {
		t.Fatalf("type: %v", err)
	}

	runes := []rune(app.edit.Text())
	if len(runes) != beforeRunes+typedRunes {
		t.Fatalf("buffer has %d runes, want %d", len(runes), beforeRunes+typedRunes)
	}

	if got := string(runes[start : start+typedRunes]); got != typed {
		t.Fatalf("buffer at offset %d = %q, want %q", start, got, typed)
	}

	if app.editCaret != start+typedRunes {
		t.Fatalf("caret after typing = %d, want %d", app.editCaret, start+typedRunes)
	}

	if app.editAnchor != app.editCaret {
		t.Fatalf("typing left anchor=%d caret=%d, want a collapsed caret", app.editAnchor, app.editCaret)
	}

	caretLine, caretCol = offsetLineCol(app.edit.Text(), app.editCaret)
	if caretLine != editCaretLine || caretCol != editCaretCol+typedRunes {
		t.Fatalf("typing caret at line %d col %d, want line %d col %d",
			caretLine, caretCol, editCaretLine, editCaretCol+typedRunes)
	}

	row = editCaretRow(t, app.View().Rows, editCaretLine, caretCol)
	if want := int(float64(caretCol) * codeCellW); row.CaretLeft != want {
		t.Fatalf("caret left after typing = %d, want %d for col %d", row.CaretLeft, want, caretCol)
	}

	if !app.View().EditDirty {
		t.Fatal("typing did not mark the buffer dirty")
	}

	t.Logf("typing lands on line %d col %d", caretLine, caretCol)

	// One press of arrowleft moves the caret one rune left and keeps the row.
	typedCaret := app.editCaret
	leftCol := caretCol - 1

	if err := app.editKey(ctx, "arrowleft"); err != nil {
		t.Fatalf("arrowleft: %v", err)
	}

	if app.editCaret != typedCaret-1 {
		t.Fatalf("arrowleft caret = %d, want %d", app.editCaret, typedCaret-1)
	}

	caretLine, caretCol = offsetLineCol(app.edit.Text(), app.editCaret)
	if caretLine != editCaretLine || caretCol != leftCol {
		t.Fatalf("arrowleft caret at line %d col %d, want line %d col %d",
			caretLine, caretCol, editCaretLine, leftCol)
	}

	row = editCaretRow(t, app.View().Rows, editCaretLine, leftCol)
	if want := int(float64(leftCol) * codeCellW); row.CaretLeft != want {
		t.Fatalf("caret left after arrowleft = %d, want %d for col %d", row.CaretLeft, want, leftCol)
	}

	// arrowright comes back to the typed caret on the same row.
	if err := app.editKey(ctx, "arrowright"); err != nil {
		t.Fatalf("arrowright: %v", err)
	}

	if app.editCaret != typedCaret {
		t.Fatalf("arrowright caret = %d, want %d", app.editCaret, typedCaret)
	}

	caretLine, caretCol = offsetLineCol(app.edit.Text(), app.editCaret)
	if caretLine != editCaretLine || caretCol != editCaretCol+typedRunes {
		t.Fatalf("arrowright caret at line %d col %d, want line %d col %d",
			caretLine, caretCol, editCaretLine, editCaretCol+typedRunes)
	}

	row = editCaretRow(t, app.View().Rows, editCaretLine, caretCol)
	if want := int(float64(caretCol) * codeCellW); row.CaretLeft != want {
		t.Fatalf("caret left after arrowright = %d, want %d for col %d", row.CaretLeft, want, caretCol)
	}

	t.Logf("arrow keys end on line %d col %d", caretLine, caretCol)
}

// editCaretEditor writes the fixture, opens it, and enters edit mode through
// the edit surface the way a pointer click does.
func editCaretEditor(t *testing.T, ctx context.Context) *App {
	t.Helper()

	dir := t.TempDir()
	writeFile(t, dir, "main.go", editCaretCode)

	app, err := New(dir) //nolint:contextcheck // New owns its bounded scan and git budget
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	app.openFile("main.go")

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

	if app.edit.Text() != editCaretCode {
		t.Fatalf("edit buffer = %q, want the fixture", app.edit.Text())
	}

	return app
}

// editCaretRedraw publishes the current caret state and paints it.
func editCaretRedraw(t *testing.T, app *App, ctx context.Context) {
	t.Helper()

	app.refresh()

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("redraw caret: %v", err)
	}
}

// editCaretRow returns the visible caret row for a source line and checks it
// sits at the expected display column.
func editCaretRow(t *testing.T, rows []codeRow, line, col int) codeRow {
	t.Helper()

	for _, row := range rows {
		if row.Line != line || !row.Caret {
			continue
		}

		length := segmentRuneLen(row.Segments)
		if col < row.Start || col > row.Start+length {
			t.Fatalf("caret col %d outside row %d start %d length %d", col, line, row.Start, length)
		}

		return row
	}

	for _, row := range rows {
		t.Logf("row line=%d start=%d caret=%v left=%d", row.Line, row.Start, row.Caret, row.CaretLeft)
	}

	t.Fatalf("no caret row on line %d", line)

	return codeRow{}
}
