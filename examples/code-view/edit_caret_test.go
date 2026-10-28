package main

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestEditCaretBlinks(t *testing.T) {
	ctx := context.Background()
	app := editCaretEditor(t, ctx)
	var id uint64
	for _, box := range app.Page().Boxes() {
		for _, animation := range box.Animations {
			if animation.Name == "edit-blink" && len(box.OperationIDs) > 0 {
				id = box.OperationIDs[0]
			}
		}
	}
	if id == 0 {
		t.Fatal("missing animated caret")
	}
	visible, hidden := false, false
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) && (!visible || !hidden) {
		if err := app.Page().Testing().Tick(ctx); err != nil {
			t.Fatal(err)
		}
		for _, op := range app.Page().Probe().Display().Ops {
			if op.ID == id {
				visible = visible || op.Opacity() > 0.8
				hidden = hidden || op.Opacity() < 0.2
			}
		}
		if !visible || !hidden {
			time.Sleep(50 * time.Millisecond)
		}
	}
	if !visible || !hidden {
		t.Fatalf("caret blink visible=%v hidden=%v", visible, hidden)
	}
}

// editCaretCode keeps every line short enough for one visual row, so a source
// line and a visual row are the same offset from the top of the code body.
const editCaretCode = "package main\n\nfunc main() {\n\tgreeting := \"hi\"\n\tprintln(greeting)\n}\n"

// editCaretLine is the zero-based source line the press targets: the third
// line, "func main() {". editCaretCol sits inside "main", past column zero.
const (
	editCaretLine = 2
	editCaretCol  = 6
)

func TestEditCaretFirstClickAndVerticalKeys(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# title\nshort\nabcdefghij\nlast line\n")
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	app.openFile("README.md")
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	body, ok := boxByID(app.Page().Boxes(), "code-body")
	if !ok {
		t.Fatal("missing code body")
	}
	x := body.X + editTextLeft + 4*codeCellW
	y := body.Y + 2*rowH + rowH/2
	if err := app.Page().Edit().Press(ctx, x, y); err != nil {
		t.Fatal(err)
	}
	if err := app.Page().Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}
	if err := app.Page().Edit().Release(ctx); err != nil {
		t.Fatal(err)
	}
	if !app.editing {
		t.Fatal("first click did not enter edit mode")
	}
	if line, col := offsetLineCol(app.edit.Text(), app.editCaret); line != 2 || col != 4 {
		t.Fatalf("click caret = %d:%d, want 2:4", line, col)
	}
	row := editCaretRow(t, app.View().Rows, 2, 4)
	img, err := png.Decode(bytes.NewReader(app.Page().Probe().PNG()))
	if err != nil {
		t.Fatal(err)
	}
	cx := int(body.X) + editTextLeft + row.CaretLeft
	cy := int(body.Y) + 2*rowH + 5
	r, g, b, _ := img.At(cx, cy).RGBA()
	if r>>8 != 122 || g>>8 != 162 || b>>8 != 247 {
		t.Fatalf("caret pixel at %d,%d = %d,%d,%d, want accent blue", cx, cy, r>>8, g>>8, b>>8)
	}
	top := int(body.Y) + 2*rowH + 2
	r, g, b, _ = img.At(cx, top).RGBA()
	if r>>8 != 122 || g>>8 != 162 || b>>8 != 247 {
		t.Fatalf("caret top pixel at %d,%d = %d,%d,%d, want accent blue", cx, top, r>>8, g>>8, b>>8)
	}
	for _, step := range []struct {
		key  string
		line int
		col  int
	}{{"arrowup", 1, 4}, {"arrowdown", 2, 4}, {"arrowdown", 3, 4}} {
		if err := app.Page().Edit().KeyDown(ctx, step.key); err != nil {
			t.Fatal(err)
		}
		if line, col := offsetLineCol(app.edit.Text(), app.editCaret); line != step.line || col != step.col {
			t.Fatalf("%s caret = %d:%d, want %d:%d", step.key, line, col, step.line, step.col)
		}
		editCaretRow(t, app.View().Rows, step.line, step.col)
	}
	if err := app.Page().Edit().Type(ctx, "X"); err != nil {
		t.Fatal(err)
	}
	if got := lineTextAt(app.edit.Text(), 3); got != "lastX line" {
		t.Fatalf("typed line = %q, want lastX line", got)
	}
}

func TestEditCaretClickAfterScroll(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	var text strings.Builder
	for i := 0; i < 100; i++ {
		fmt.Fprintf(&text, "line %03d text\n", i)
	}
	writeFile(t, dir, "README.md", text.String())
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	app.openFile("README.md")
	app.wrap = false
	app.Page().Scroll().SetScrollOffset(0, 200)
	app.onScroll(200, app.viewportH)
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	body, ok := boxByID(app.Page().Boxes(), "code-body")
	if !ok {
		t.Fatal("missing code body")
	}
	if err := app.Page().Edit().Press(ctx, body.X+editTextLeft+3*codeCellW, body.Y+200+2*rowH+rowH/2); err != nil {
		t.Fatal(err)
	}
	if line, col := offsetLineCol(app.edit.Text(), app.editCaret); line != 12 || col != 3 {
		t.Fatalf("scrolled click caret = %d:%d, want 12:3", line, col)
	}
}

func TestEditCaretWrappedClickAndVerticalKeys(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	writeFile(t, dir, "README.md", strings.Repeat("x", 240)+"\nend\n")
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	app.Page().SetSize(700, 500)
	app.openFile("README.md")
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	body, ok := boxByID(app.Page().Boxes(), "code-body")
	if !ok {
		t.Fatal("missing code body")
	}
	rows := app.activeDoc().wrapped(app.codeColsForGroup(0)).Rows
	if len(rows) < 3 || rows[1].Line != 0 || rows[2].Line != 0 {
		t.Fatalf("expected wrapped rows: %+v", rows)
	}
	if err := app.Page().Edit().Press(ctx, body.X+editTextLeft+4*codeCellW, body.Y+rowH+rowH/2); err != nil {
		t.Fatal(err)
	}
	if line, col := offsetLineCol(app.edit.Text(), app.editCaret); line != 0 || col != rows[1].Start+4 {
		t.Fatalf("wrapped click caret = %d:%d, want 0:%d", line, col, rows[1].Start+4)
	}
	if err := app.Page().Edit().KeyDown(ctx, "arrowdown"); err != nil {
		t.Fatal(err)
	}
	if line, col := offsetLineCol(app.edit.Text(), app.editCaret); line != 0 || col != rows[2].Start+4 {
		t.Fatalf("wrapped down caret = %d:%d, want 0:%d", line, col, rows[2].Start+4)
	}
}

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
