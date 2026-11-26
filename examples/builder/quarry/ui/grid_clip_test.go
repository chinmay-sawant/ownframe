package ui

import (
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
