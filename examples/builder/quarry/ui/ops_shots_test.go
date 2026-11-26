package ui

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestShotsEditorAndGridOps captures the editor and grid features (selection,
// find, error marker, horizontal scroll, frozen column, sort, filter, context
// menu data) in both themes. It runs only with QUARRY_SHOTS=1.
func TestShotsEditorAndGridOps(t *testing.T) {
	if os.Getenv("QUARRY_SHOTS") == "" {
		t.Skip("set QUARRY_SHOTS=1 to capture the screenshots")
	}
	for _, dark := range []bool{true, false} {
		name := map[bool]string{true: "dark", false: "light"}[dark]
		r := newRig(t, 1440, 960, "")
		if !dark {
			r.tap("tb-theme")
		}
		ed := r.setEditor("SELECT o.id, o.status, 'it''s' AS note\nFROM orders o\n-- a long comment line " + strings.Repeat("with many many words ", 14) + "\nWHERE o.total_cents > 1000\nORDER BY o.id")
		ed.anchor, ed.caret = ed.offsetAt(0, 7), ed.offsetAt(0, 11)
		r.png("30-editor-selection-" + name)
		ed.hcol = 40
		r.must(r.a.redraw(r.ctx))
		r.png("31-editor-hscroll-" + name)
		ed.hcol = 0
		r.chord("f", false)
		r.a.page.Form().SetFormValue("find-input", "o.")
		r.must(r.a.onChange(r.ctx, r.box("find-input")))
		r.png("32-editor-find-" + name)
		r.key("escape")
		r.a.activeTabPtr().ed = newEditor("SELECT 1;\nSELEKT * FROM users")
		r.chord("enter", true)
		r.settle()
		r.png("33-editor-error-line-" + name)
		r.must(r.a.RunSQL(r.ctx, wideQuery(30)))
		r.tap("rs-freeze")
		for i := 0; i < 6; i++ {
			_, _ = r.a.onWheel(r.ctx, 700, 500, -3, 0)
		}
		r.tap("cell-1-9")
		r.png("34-grid-hscroll-frozen-" + name)
		r.tap("rs-freeze")
		r.must(r.a.RunSQL(r.ctx, "SELECT id, full_name, country, plan FROM users WHERE id <= 60"))
		r.tap("grid-h-2")
		r.tap("grid-h-2")
		r.a.page.Form().SetFormValue("grid-filter", "pro")
		r.must(r.a.onChange(r.ctx, r.box("grid-filter")))
		r.tap("cell-2-1")
		r.png("35-grid-sort-filter-" + name)
		_ = fmt.Sprint
	}
}
