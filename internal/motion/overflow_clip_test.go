package motion

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// TestStampTransformClipsRecordsAncestorBox pins the overflow-clip handoff:
// an operation a transform player writes carries its owner box's ancestor
// clip, so the replay can intersect transformed paint with it. Entries that
// write no matrix leave their ops alone.
func TestStampTransformClipsRecordsAncestorBox(t *testing.T) {
	d := &layout.Display{
		Width: 200, Height: 100, PointsPerPixel: 0.75, PixelPerPoint: 4.0 / 3.0,
		Ops: []layout.DisplayOp{
			{ID: 1, Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20},
			{ID: 2, Kind: layout.DisplayOpFillRect, X: 50, Y: 50, W: 10, H: 10},
		},
		OpOwners: []int{0, 1},
		Boxes: []layout.Box{
			{ID: "clip", Tag: "div", ParentIndex: -1,
				OverflowClip:    [4]float64{8, 8, 160, 80},
				HasOverflowClip: true,
				OperationIDs:    []uint64{1}},
			{ID: "free", Tag: "div", ParentIndex: -1, OperationIDs: []uint64{2}},
		},
	}

	entries := []entry{
		{mask: fieldMatrix, indices: []int{0}, boxIndex: 0},
		{mask: fieldOpacity, indices: []int{1}, boxIndex: 1},
	}

	stampTransformClips(d, entries)

	if x, y, w, h, ok := d.Ops[0].OverflowClip(); !ok {
		t.Fatal("matrix-driven op has no overflow clip")
	} else if x != 8*0.75 || y != 8*0.75 || w != 160*0.75 || h != 80*0.75 {
		t.Fatalf("clip = %.2f,%.2f %.2fx%.2f, want 6,6 120x60", x, y, w, h)
	}

	if _, _, _, _, ok := d.Ops[1].OverflowClip(); ok {
		t.Fatal("opacity-only op gained an overflow clip")
	}
}
