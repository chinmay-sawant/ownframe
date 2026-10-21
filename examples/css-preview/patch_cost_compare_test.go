package main

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// addPatchStages adds the stage times of the page's last redraw to the sums
// the benchmark reports.
func addPatchStages(p *ownframe.Page, template, cascade, layout time.Duration) (
	time.Duration, time.Duration, time.Duration,
) {
	s := p.Stats()
	return template + s.LastTemplate, cascade + s.LayoutTime, layout + s.DisplayListTime
}

// comparePatchOps compares two operation slices operation by operation. The
// kind, the geometry, the text, and the color must all agree. The operation
// id is left out: a full redraw assigns new ids.
func comparePatchOps(t *testing.T, label string, got, want []ownframe.DisplayOp) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: %d operations, want %d", label, len(got), len(want))
	}
	for i := range want {
		a, b := got[i], want[i]
		if a.Kind != b.Kind || a.X != b.X || a.Y != b.Y || a.W != b.W || a.H != b.H {
			t.Fatalf("%s operation %d: kind %v box %v %v %v %v, want %v %v %v %v %v",
				label, i, a.Kind, a.X, a.Y, a.W, a.H, b.Kind, b.X, b.Y, b.W, b.H)
		}
		if a.R != b.R || a.G != b.G || a.B != b.B || a.Alpha != b.Alpha {
			t.Fatalf("%s operation %d: rgba %v %v %v %v, want %v %v %v %v",
				label, i, a.R, a.G, a.B, a.Alpha, b.R, b.G, b.B, b.Alpha)
		}
		if a.Text != b.Text {
			t.Fatalf("%s operation %d: text %q, want %q", label, i, a.Text, b.Text)
		}
	}
}

// comparePatchBox compares two element boxes. The parent index is left out: a
// patch carries the subtree's boxes on their own.
func comparePatchBox(t *testing.T, got, want ownframe.Box) {
	t.Helper()
	if got.ID != want.ID || got.Tag != want.Tag || got.X != want.X || got.Y != want.Y ||
		got.W != want.W || got.H != want.H {
		t.Fatalf("box %+v, want %+v", got, want)
	}
}
