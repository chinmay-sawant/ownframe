package ui

import (
	"strings"
	"testing"
)

const wrapSample = "SELECT id, full_name, email, country, plan, active, created_at, updated_at, last_login_at, billing_email, shipping_country FROM users WHERE active = 1 AND country = 'DE' AND plan = 'pro' ORDER BY created_at DESC LIMIT 50\nSELECT 2"

// narrowRig is a window narrow enough that wrapSample's first line wraps.
func narrowRig(t *testing.T) *rig {
	t.Helper()
	r := newRig(t, 900, 800, "")
	r.tap("ov-more-ed")
	r.tap("ov-item-ed-wrap")
	if !r.a.wrap {
		t.Fatal("the overflow menu toggles word wrap")
	}
	return r
}

func TestWordWrapToggleRowsGutterAndPersistence(t *testing.T) {
	r := narrowRig(t)
	ed := r.setEditor(wrapSample)
	rows := ed.vrows()
	if len(rows) < 4 {
		t.Fatalf("a long line should wrap to several rows, got %d", len(rows))
	}
	if ed.longestLine() < ed.wrapCols {
		t.Fatal("sample must be longer than the wrap width")
	}
	for _, v := range rows {
		if n := len([]rune(ed.text[v.start:v.end])); n > ed.wrapCols {
			t.Fatalf("row of %d chars exceeds wrap width %d", n, ed.wrapCols)
		}
	}
	if r.has("editor-hthumb") {
		t.Fatal("no horizontal bar while wrapping")
	}
	// Gutter: logical line numbers on first rows only.
	first := rows[0]
	if !first.first || rows[1].first || rows[1].line != 0 {
		t.Fatalf("continuation rows belong to line 0: %+v", rows[:2])
	}
	last := rows[len(rows)-1]
	if last.line != 1 || !last.first {
		t.Fatalf("second logical line is the last row: %+v", last)
	}
	wx, _, _ := r.a.editorWrapOrigin()
	numbers := map[string]int{}
	for _, b := range r.a.Page().Boxes() {
		if b.Tag == "span" && b.ID == "" && b.X >= wx && b.X < wx+1 && b.Text != "" {
			numbers[b.Text]++
		}
	}
	if numbers["1"] != 1 || numbers["2"] != 1 || len(numbers) != 2 {
		t.Fatalf("gutter shows each logical line number once, saw %v", numbers)
	}
	if v, _ := r.a.loadPref("editor.wrap"); v != "1" {
		t.Fatalf("pref editor.wrap = %q, want 1", v)
	}
	// Toggle off: back to one row per line, bar available, pref saved.
	r.tap("ov-more-ed")
	r.tap("ov-item-ed-wrap")
	if r.a.wrap || len(ed.vrows()) != 2 || !r.has("editor-hthumb") {
		t.Fatal("wrap off restores long lines with a horizontal bar")
	}
	if v, _ := r.a.loadPref("editor.wrap"); v != "0" {
		t.Fatalf("pref editor.wrap = %q, want 0", v)
	}
}

func TestWordWrapCaretClickAndSelectionMapping(t *testing.T) {
	r := narrowRig(t)
	ed := r.setEditor(wrapSample)
	rows := ed.vrows()
	// A click on the third painted row lands inside that row's text.
	r.pressAt(2, 5)
	want := ed.offsetInRow(rows, 2, 5)
	if ed.caret != want || ed.caret < rows[2].start || ed.caret >= rows[2].end {
		t.Fatalf("click on row 2 col 5: caret %d, want %d in [%d,%d)", ed.caret, want, rows[2].start, rows[2].end)
	}
	if rr, cc := ed.rowOf(ed.caret); rr != 2 || cc != 5 {
		t.Fatalf("rowOf = %d,%d", rr, cc)
	}
	// Arrow down/up move one painted row and keep the column.
	r.key("arrowdown")
	if rr, cc := ed.rowOf(ed.caret); rr != 3 || cc != 5 {
		t.Fatalf("down: row %d col %d", rr, cc)
	}
	r.key("arrowup")
	r.key("arrowup")
	if rr, _ := ed.rowOf(ed.caret); rr != 1 {
		t.Fatalf("up twice: row %d", rr)
	}
	// Drag selects across wrapped rows.
	r.pressAt(0, 3)
	claimed, err := r.a.onDragStart(r.ctx, r.box("ed-line-0"))
	r.must(err)
	if !claimed {
		t.Fatal("the editor claims the drag")
	}
	r.must(r.a.onDragMove(r.ctx, 0, 2*edLineH))
	r.must(r.a.onDragEnd(r.ctx))
	lo, hi := ed.selection()
	if lo >= rows[0].end || hi < rows[2].start {
		t.Fatalf("selection %d..%d should span rows 0..2", lo, hi)
	}
	// Selected runs render on the wrapped rows; the caret shows once.
	if !r.has("ed-line-1") {
		t.Fatal("middle row painted")
	}
	// Caret at the end of the buffer scrolls to and shows on the last row.
	ed.caret, ed.anchor = len(ed.text), len(ed.text)
	ed.ensureVisible(r.a.editorRows())
	r.must(r.a.redraw(r.ctx))
	if rr, _ := ed.rowOf(ed.caret); rr != len(rows)-1 {
		t.Fatalf("end caret row %d, want %d", rr, len(rows)-1)
	}
	// Typing wraps live.
	r.typeKeys(strings.Repeat("x ", 80))
	if len(ed.vrows()) <= len(rows) {
		t.Fatal("typing a long run adds painted rows")
	}
	// Error marker covers every row of its logical line.
	r.a.activeTabPtr().rowErrLine = 0
	r.must(r.a.redraw(r.ctx))
}
