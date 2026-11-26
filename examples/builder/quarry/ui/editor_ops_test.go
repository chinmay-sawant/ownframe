package ui

import (
	"image/color"
	"strings"
	"testing"
	"time"
)

// pressAt presses the editor at (line, col).
func (r *rig) pressAt(line, col int) {
	r.t.Helper()
	wx, wy, _ := r.a.editorWrapOrigin()
	x := wx + edGutterW + float64(col)*edCharW + 2
	y := wy + 4 + float64((line-r.a.activeTabPtr().ed.scroll)*edLineH) + 6
	_, err := r.a.onPress(r.ctx, r.box("ed-line-"+itoa(line)), x, y)
	r.must(err)
}

func (r *rig) setEditor(text string) *editor {
	r.t.Helper()
	t := r.a.activeTabPtr()
	t.ed = newEditor(text)
	t.ed.caret, t.ed.anchor = 0, 0
	r.a.focus = "editor"
	r.must(r.a.redraw(r.ctx))
	return t.ed
}

func TestEditorDoubleClickWordAndTripleClickLine(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	ed := r.setEditor("SELECT customer_id FROM orders\nWHERE x = 1")
	r.pressAt(0, 10) // inside customer_id
	r.now = r.now.Add(100 * time.Millisecond)
	r.pressAt(0, 10)
	if ed.selectionText() != "customer_id" {
		t.Fatalf("double click selected %q", ed.selectionText())
	}
	r.now = r.now.Add(100 * time.Millisecond)
	r.pressAt(0, 10)
	if ed.selectionText() != "SELECT customer_id FROM orders\n" {
		t.Fatalf("triple click selected %q", ed.selectionText())
	}
	// A slow second click starts over as a plain caret placement.
	r.now = r.now.Add(2 * time.Second)
	r.pressAt(1, 3)
	if ed.hasSelection() {
		t.Fatal("a late click must not extend the multi-click")
	}
}

func TestEditorDragSelection(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	ed := r.setEditor("alpha beta gamma\ndelta epsilon")
	r.pressAt(0, 6)
	claimed, err := r.a.onDragStart(r.ctx, r.box("ed-line-0"))
	r.must(err)
	if !claimed {
		t.Fatal("the editor must claim a drag that starts on a line")
	}
	r.must(r.a.onDragMove(r.ctx, 4*edCharW, 0))
	if ed.selectionText() != "beta" {
		t.Fatalf("horizontal drag selected %q", ed.selectionText())
	}
	r.must(r.a.onDragMove(r.ctx, 0, float64(edLineH)))
	if !strings.HasPrefix(ed.selectionText(), "beta gamma\n") {
		t.Fatalf("dragging down selected %q", ed.selectionText())
	}
	r.must(r.a.onDragEnd(r.ctx))
	if r.a.edDrag {
		t.Fatal("drag state must clear")
	}
}

func TestEditorWordJumpsHomeEndAndShiftSelection(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	ed := r.setEditor("select  alpha_beta, gamma")
	ed.caret, ed.anchor = len(ed.text), len(ed.text)
	r.key("ctrl+arrowleft")
	if ed.text[ed.caret:] != "gamma" {
		t.Fatalf("ctrl+left landed at %q", ed.text[ed.caret:])
	}
	r.key("ctrl+arrowleft")
	if ed.text[ed.caret:] != "alpha_beta, gamma" {
		t.Fatalf("second jump landed at %q", ed.text[ed.caret:])
	}
	r.key("ctrl+arrowright")
	if ed.text[:ed.caret] != "select  alpha_beta" {
		t.Fatalf("ctrl+right landed after %q", ed.text[:ed.caret])
	}
	r.key("shift+ctrl+arrowright")
	if ed.selectionText() != ", gamma" {
		t.Fatalf("shift+ctrl+right selected %q", ed.selectionText())
	}
	r.key("home")
	if ed.caret != 0 || ed.hasSelection() {
		t.Fatal("Home goes to the line start and clears the selection")
	}
	r.key("shift+end")
	if ed.selectionText() != ed.text {
		t.Fatalf("shift+End selected %q", ed.selectionText())
	}
}

func TestEditorSelectAllCopyCutPaste(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	ed := r.setEditor("SELECT 1")
	r.must(r.a.onSelectAll(r.ctx))
	got, ok, err := r.a.onCopy(r.ctx)
	r.must(err)
	if !ok || got != "SELECT 1" {
		t.Fatalf("copy = %q %v", got, ok)
	}
	cut, ok, err := r.a.onCut(r.ctx)
	r.must(err)
	if !ok || cut != "SELECT 1" || ed.text != "" {
		t.Fatalf("cut = %q %v, buffer %q", cut, ok, ed.text)
	}
	r.must(r.a.onPaste(r.ctx, cut))
	if ed.text != "SELECT 1" {
		t.Fatalf("paste = %q", ed.text)
	}
	r.must(r.a.onUndo(r.ctx))
	r.must(r.a.onUndo(r.ctx))
	if ed.text != "SELECT 1" {
		t.Fatalf("undo of paste and cut = %q", ed.text)
	}
}

func TestEditorTabIndentAndOutdent(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	ed := r.setEditor("a\nb\nc")
	r.key("tab")
	if ed.text != "  a\nb\nc" {
		t.Fatalf("tab at the caret = %q", ed.text)
	}
	ed.anchor, ed.caret = 0, len(ed.text)
	r.key("tab")
	if ed.text != "    a\n  b\n  c" {
		t.Fatalf("tab over a selection = %q", ed.text)
	}
	r.a.SetModifiers(func() (bool, bool) { return true, false })
	r.key("tab")
	r.a.SetModifiers(func() (bool, bool) { return false, false })
	if ed.text != "  a\nb\nc" {
		t.Fatalf("shift+tab = %q", ed.text)
	}
}

func TestEditorCommentToggle(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	ed := r.setEditor("SELECT 1\n\nFROM t")
	ed.anchor, ed.caret = 0, len(ed.text)
	r.chord("slash", false)
	if ed.text != "-- SELECT 1\n\n-- FROM t" {
		t.Fatalf("comment = %q", ed.text)
	}
	ed.anchor, ed.caret = 0, len(ed.text)
	r.chord("slash", false)
	if ed.text != "SELECT 1\n\nFROM t" {
		t.Fatalf("uncomment = %q", ed.text)
	}
	// One uncommented line makes the toggle comment everything.
	ed.text = "-- a\nb"
	ed.anchor, ed.caret = 0, len(ed.text)
	r.chord("slash", false)
	if ed.text != "-- -- a\n-- b" {
		t.Fatalf("mixed = %q", ed.text)
	}
	// A bare caret comments its own line only.
	ed.text, ed.anchor, ed.caret = "x\ny", 3, 3
	r.chord("slash", false)
	if ed.text != "x\n-- y" {
		t.Fatalf("single line = %q", ed.text)
	}
}

func TestEditorAutoPairBracketsAndQuotes(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	ed := r.setEditor("")
	r.typeKeys("count(")
	if ed.text != "count()" || ed.caret != len("count(") {
		t.Fatalf("paren pair = %q caret %d", ed.text, ed.caret)
	}
	r.typeKeys(")")
	if ed.text != "count()" || ed.caret != len(ed.text) {
		t.Fatalf("typing the closer should step over it: %q caret %d", ed.text, ed.caret)
	}
	r.typeKeys(" = 'a")
	if ed.text != "count() = 'a'" {
		t.Fatalf("quote pair = %q", ed.text)
	}
	r.typeKeys("'")
	if ed.text != "count() = 'a'" || ed.caret != len(ed.text) {
		t.Fatalf("closing quote = %q", ed.text)
	}
	ed.text, ed.caret, ed.anchor = "x", 1, 1
	r.typeKeys("'")
	if ed.text != "x'" {
		t.Fatalf("an apostrophe after a word must not pair: %q", ed.text)
	}
	ed.text, ed.anchor, ed.caret = "word", 0, 4
	r.typeKeys("[")
	if ed.text != "[word]" {
		t.Fatalf("selection wrap = %q", ed.text)
	}
}

func TestEditorHorizontalScrollAndBar(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	long := "SELECT " + strings.Repeat("column_name_that_is_long, ", 20) + "1"
	ed := r.setEditor("")
	r.typ(long)
	cols := r.a.editorViewCols()
	if ed.hcol == 0 || len(long) <= cols {
		t.Fatalf("typing a %d-char line must scroll right (hcol %d, cols %d)", len(long), ed.hcol, cols)
	}
	if !r.has("editor-hthumb") {
		t.Fatal("a horizontal scrollbar must appear for long lines")
	}
	// The gutter stays; only the visible window of the line is painted.
	if !strings.Contains(r.texts(), "1\n") || strings.Contains(r.texts(), "SELECT column_name") {
		t.Fatal("the line start should have scrolled out of the window")
	}
	// Dragging the thumb to the start restores the line head.
	thumb := r.box("editor-hthumb")
	claimed, err := r.a.onDragStart(r.ctx, thumb)
	r.must(err)
	if !claimed {
		t.Fatal("the horizontal thumb must claim the drag")
	}
	r.must(r.a.onDragMove(r.ctx, -5000, 0))
	r.must(r.a.onDragEnd(r.ctx))
	if ed.hcol != 0 {
		t.Fatalf("dragging the thumb to the start left hcol %d", ed.hcol)
	}
	if !strings.Contains(r.texts(), "SELECT") {
		t.Fatal("the line head should be painted again")
	}
	// Wheel with dx scrolls the editor sideways too.
	wx, wy, _ := r.a.editorWrapOrigin()
	_, err = r.a.onWheel(r.ctx, wx+100, wy+20, -3, 0)
	r.must(err)
	if ed.hcol == 0 {
		t.Fatal("horizontal wheel should scroll the editor")
	}
	// Short text hides the bar.
	r.setEditor("SELECT 1")
	if r.has("editor-hthumb") {
		t.Fatal("the bar should hide when the line fits")
	}
}

func TestEditorErrorLineMarker(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	tab := r.a.activeTabPtr()
	tab.ed = newEditor("SELECT 1;\nSELEKT * FROM users")
	r.chord("enter", true)
	r.settle()
	if tab.err == nil || tab.rowErrLine != 1 {
		t.Fatalf("expected an error on line 2, got %+v (marker %d)", tab.err, tab.rowErrLine)
	}
	line := r.box("ed-line-1")
	want := color.RGBA{0x3a, 0x20, 0x23, 0xff}
	if got := pixelAt(r, int(line.X+line.W-30), int(line.Y+line.H/2)); got != want {
		t.Fatalf("error line background %v, want %v", got, want)
	}
	ok := r.box("ed-line-0")
	if got := pixelAt(r, int(ok.X+ok.W-30), int(ok.Y+ok.H/2)); got == want {
		t.Fatal("only the failing line is marked")
	}
}

func TestEditorFind(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	ed := r.setEditor("SELECT id FROM users u JOIN orders o ON o.user_id = u.id\nWHERE u.id > 1")
	r.chord("f", false)
	if !r.has("find-bar") || r.a.page.Form().FocusID() != "find-input" {
		t.Fatal("Ctrl+F should open the find bar and focus its field")
	}
	r.a.page.Form().SetFormValue("find-input", "id")
	r.must(r.a.onChange(r.ctx, r.box("find-input")))
	if len(r.a.find.matches) != 4 {
		t.Fatalf("matches = %v", r.a.find.matches)
	}
	if ed.selectionText() != "id" || ed.caret != r.a.find.matches[0]+2 {
		t.Fatalf("first match not selected: %q at %d", ed.selectionText(), ed.caret)
	}
	if !strings.Contains(r.texts(), "1 of 4") {
		t.Fatalf("count label: %q", r.texts())
	}
	r.key("enter")
	if r.a.find.idx != 1 {
		t.Fatalf("Enter should step to the next match, idx %d", r.a.find.idx)
	}
	r.tap("find-prev")
	if r.a.find.idx != 0 {
		t.Fatalf("Prev -> %d", r.a.find.idx)
	}
	r.tap("find-prev")
	if r.a.find.idx != 3 {
		t.Fatalf("Prev wraps to the last match, got %d", r.a.find.idx)
	}
	r.a.page.Form().SetFormValue("find-input", "zzz")
	r.must(r.a.onChange(r.ctx, r.box("find-input")))
	if !strings.Contains(r.texts(), "No matches") {
		t.Fatal("no-match label")
	}
	r.key("escape")
	if r.has("find-bar") || r.a.focus != "editor" {
		t.Fatal("Escape closes the find bar and returns to the editor")
	}
}

// typeKeys types one character at a time, as the keyboard does.
func (r *rig) typeKeys(s string) {
	r.t.Helper()
	for _, c := range s {
		r.typ(string(c))
	}
}
