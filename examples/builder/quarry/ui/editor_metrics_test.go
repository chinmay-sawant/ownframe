package ui

import (
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// TestEditorMetricsMatchPaint checks that the editor's character width and
// line height constants equal what the engine paints, so clicks map to the
// right column.
func TestEditorMetricsMatchPaint(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.a.activeTabPtr().ed = newEditor("SELECT abcdefghij FROM users")
	r.must(r.a.redraw(r.ctx))
	line := r.box("ed-line-0")
	if int(line.H) != edLineH {
		t.Fatalf("line height %v, want %d", line.H, edLineH)
	}
	var first, last ownframe.DisplayOp
	found := false
	for _, op := range r.a.Page().Probe().Display().Ops {
		if op.Kind != ownframe.DisplayOpText {
			continue
		}
		if op.Text == "SELECT" {
			first, found = op, true
		}
		if op.Text == "users" {
			last = op
		}
	}
	if !found {
		t.Fatal("no SELECT run")
	}
	// Display ops are in points; boxes are CSS pixels (1 px = 0.75 pt).
	const ptToPx = 4.0 / 3.0
	cw := first.W / 6 * ptToPx
	_ = last
	t.Logf("char width %.3f px, code offset %.1f px", cw, first.X*ptToPx-line.X)
	if cw < edCharW-0.3 || cw > edCharW+0.3 {
		t.Errorf("edCharW=%v but engine paints %.3f", edCharW, cw)
	}
	off := first.X*ptToPx - line.X
	if off < float64(edGutterW+edPadLeft)-3 || off > float64(edGutterW+edPadLeft)+3 {
		t.Errorf("code starts %.1f px into the line, constant says %d", off, edGutterW+edPadLeft)
	}
}
