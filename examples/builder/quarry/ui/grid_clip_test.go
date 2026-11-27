package ui

import (
	"image/color"
	"strings"
	"testing"
)

// TestGridRowsNeverPaintPastTheBody proves the virtual grid shows whole rows
// only: no row straddles the bottom or top edge of the body, and the body ends
// above the status bar, in a table tab and in a query-result grid.
func TestGridRowsNeverPaintPastTheBody(t *testing.T) {
	check := func(r *rig, label string) {
		t.Helper()
		body, status := r.box("grid-body"), r.box("statusbar")
		if body.Y+body.H > status.Y+0.5 {
			t.Errorf("%s: grid body (%v..%v) overlaps the status bar at %v", label, body.Y, body.Y+body.H, status.Y)
		}
		if int(body.H)%gridRowHeight != 0 {
			t.Errorf("%s: body height %v is not whole rows", label, body.H)
		}
		for _, b := range r.a.Page().Boxes() {
			if !strings.HasPrefix(b.ID, "grid-row-") {
				continue
			}
			straddlesBottom := b.Y < body.Y+body.H && b.Y+b.H > body.Y+body.H+0.5
			straddlesTop := b.Y < body.Y-0.5 && b.Y+b.H > body.Y+0.5
			if straddlesBottom || straddlesTop {
				t.Errorf("%s: %s (y=%v h=%v) straddles the body edge %v..%v", label, b.ID, b.Y, b.H, body.Y, body.Y+body.H)
			}
		}
	}
	r := newRig(t, 1440, 960, "")
	r.openTableByName("audit_events")
	check(r, "table")
	body := r.box("grid-body")
	for i := 0; i < 5; i++ {
		_, err := r.a.onWheel(r.ctx, body.X+40, body.Y+40, 0, -1)
		r.must(err)
		check(r, "table scrolled")
	}
	r.a.activeTabPtr().inspectorOpen = true
	r.must(r.a.redraw(r.ctx))
	check(r, "table with inspector")
	q := newRig(t, 1280, 800, "")
	q.must(q.a.activeTabPtr().ed.insertAndRun(q, "SELECT * FROM users"))
	check(q, "query")
}

// insertAndRun is a test helper: it sets the editor text and runs it.
func (e *editor) insertAndRun(r *rig, sql string) error {
	r.a.activeTabPtr().ed = newEditor(sql)
	r.chord("enter", false)
	r.settle()
	return nil
}

// isAccentTinted reports whether pixel c carries the Quarry green selection
// accent (G noticeably stronger than R in both dark and light themes).
func isAccentTinted(c color.RGBA) bool {
	return int(c.G)-int(c.R) >= 35
}

// TestRowSelectionBottomBorderVisible proves clicking a row number shows a
// full bottom accent border on both middle rows (above an alternating row) and
// the bottom-most visible row (even when grid-body meets the status bar), in
// table tabs and query grids, in dark and light themes.
func TestRowSelectionBottomBorderVisible(t *testing.T) {
	checkRow := func(r *rig, label, idx string, wantGutter bool) {
		t.Helper()
		row := r.box("grid-row-" + idx)
		gutter := r.box("grid-gutter-" + idx)
		cell := r.box("cell-" + idx + "-0")
		if int(cell.H) != gridRowHeight-1 || int(gutter.H) != gridRowHeight-1 {
			t.Fatalf("%s: gutter H=%v cell H=%v, want %d (row content height)", label, gutter.H, cell.H, gridRowHeight-1)
		}
		// Bottom border sits inside the row at dy = 24..25 (above row.Y+26).
		botY := int(row.Y) + gridRowHeight - 1
		if c := pixelAt(r, int(cell.X+cell.W/2), botY); !isAccentTinted(c) {
			t.Errorf("%s: cell bottom border at y=%d is %v, expected accent tint", label, botY, c)
		}
		gc := pixelAt(r, int(gutter.X+gutter.W/2), botY)
		if wantGutter && !isAccentTinted(gc) {
			t.Errorf("%s: gutter bottom border at y=%d is %v, expected accent tint", label, botY, gc)
		}
		if !wantGutter && isAccentTinted(gc) {
			t.Errorf("%s: gutter bottom border at y=%d is %v, expected no accent on single-cell select", label, botY, gc)
		}
	}

	bottomGutter := func(r *rig) string {
		var lastID string
		var lastY float64 = -1
		for _, b := range r.a.Page().Boxes() {
			if strings.HasPrefix(b.ID, "grid-gutter-") && b.Y > lastY {
				lastY = b.Y
				lastID = b.ID
			}
		}
		if lastID == "" {
			t.Fatal("no visible grid-gutter-* box")
		}
		return lastID
	}

	// 1440x958 makes grid-body end flush with statusbar (Y=934).
	r := newRig(t, 1440, 958, "")
	r.openTableByName("order_items")
	for _, theme := range []string{"dark", "light"} {
		if theme == "light" {
			r.tap("tb-theme")
		}
		r.tap("grid-gutter-4")
		checkRow(r, theme+"-table-middle", "4", true)

		lastID := bottomGutter(r)
		idx := strings.TrimPrefix(lastID, "grid-gutter-")
		r.tap(lastID)
		checkRow(r, theme+"-table-bottom", idx, true)

		// Single-cell selection keeps its bottom border without bordering the gutter.
		r.tap("cell-" + idx + "-0")
		checkRow(r, theme+"-table-single-cell", idx, false)
	}

	// Query result grid uses the same grid renderer and styles.
	q := newRig(t, 1280, 800, "")
	q.must(q.a.activeTabPtr().ed.insertAndRun(q, "SELECT * FROM order_items LIMIT 50"))
	lastID := bottomGutter(q)
	idx := strings.TrimPrefix(lastID, "grid-gutter-")
	q.tap(lastID)
	checkRow(q, "query-bottom", idx, true)
}
