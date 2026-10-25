package cssmotion

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestOperationIndicesUsesBlinklessOwnership(t *testing.T) {
	d := &layout.Display{Ops: []layout.DisplayOp{
		{ID: 11, X: 10, Y: 10, W: 8, H: 8},
		{ID: 22, X: 10, Y: 10, W: 8, H: 8},
		{ID: 33, X: 10, Y: 10, W: 8, H: 8},
	}}
	child := layout.Box{X: 0, Y: 0, W: 30, H: 30, OperationIDs: []uint64{22}}
	got := OperationIndices(d, child)
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("child indices = %v, want [1] despite overlapping geometry", got)
	}
	parent := child
	parent.OperationIDs = []uint64{11, 22}
	got = OperationIndices(d, parent)
	if len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Fatalf("parent indices = %v, want [0 1]", got)
	}
}

func TestOperationIndicesFallsBackForLegacyBox(t *testing.T) {
	d := &layout.Display{PointsPerPixel: 1, Ops: []layout.DisplayOp{
		{ID: 11, X: 5, Y: 5, W: 2, H: 2},
		{ID: 22, X: 40, Y: 40, W: 2, H: 2},
	}}
	got := OperationIndices(d, layout.Box{X: 0, Y: 0, W: 20, H: 20})
	if len(got) != 1 || got[0] != 0 {
		t.Fatalf("legacy indices = %v, want [0] from geometry", got)
	}
}

func TestOperationIndicesRespectsExplicitEmptyOwnership(t *testing.T) {
	d := &layout.Display{PointsPerPixel: 1, Ops: []layout.DisplayOp{
		{ID: 11, X: 5, Y: 5, W: 2, H: 2},
	}}
	box := layout.Box{X: 0, Y: 0, W: 20, H: 20, OperationIDs: []uint64{}}
	if got := OperationIndices(d, box); len(got) != 0 {
		t.Fatalf("explicit empty ownership indices = %v, want none", got)
	}
}
