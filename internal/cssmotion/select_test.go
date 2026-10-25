package cssmotion

import (
	"slices"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// selectDisplay is one parent with a fill, a label, and a border around a child
// card. The owners are explicit so the test does not depend on geometry.
func selectDisplay(owners []int) *layout.Display {
	return &layout.Display{
		PointsPerPixel: 1,
		Ops: []layout.DisplayOp{
			{ID: 11, Kind: layout.DisplayOpFillRect, X: 1, Y: 1, W: 8, H: 8},
			{ID: 12, Kind: layout.DisplayOpText, X: 2, Y: 4, W: 4, H: 4},
			{ID: 13, Kind: layout.DisplayOpStrokeRect, X: 1, Y: 1, W: 8, H: 8},
			{ID: 14, Kind: layout.DisplayOpLine, X: 1, Y: 9, W: 8, H: 1},
			{ID: 21, Kind: layout.DisplayOpFillRect, X: 3, Y: 3, W: 4, H: 4},
		},
		OpOwners: owners,
		Boxes: []layout.Box{
			{ID: "parent", OperationIDs: []uint64{11, 12, 13, 14, 21}, ParentIndex: -1},
			{ID: "child", OperationIDs: []uint64{21}, ParentIndex: 0},
		},
	}
}

func TestSelectKeepsColourOnTheNearestOwner(t *testing.T) {
	d := selectDisplay([]int{0, 0, 0, 0, 1})
	parent := d.Boxes[0]
	tests := []struct {
		kind Kind
		want []int
	}{
		{KindFill, []int{0}},
		{KindInk, []int{1}},
		{KindStroke, []int{2, 3}},
		{KindSubtree, []int{0, 1, 2, 3, 4}},
	}
	for _, test := range tests {
		got := Select(d, parent, 0, test.kind)
		if !slices.Equal(got, test.want) {
			t.Errorf("Select kind %d = %v, want %v", test.kind, got, test.want)
		}
	}
}

// TestSelectFallsBackWithoutOwners checks the documented fallback: an older
// engine that supplies no OpOwners has the owner kinds filtered from the box's
// subtree, so a parent's colour can still touch a child's fill there.
func TestSelectFallsBackWithoutOwners(t *testing.T) {
	d := selectDisplay(nil)
	parent := d.Boxes[0]
	if got := Select(d, parent, 0, KindFill); !slices.Equal(got, []int{0, 4}) {
		t.Fatalf("fill fallback = %v, want the subtree fills", got)
	}
	if got := Select(d, parent, -1, KindInk); !slices.Equal(got, []int{1}) {
		t.Fatalf("ink fallback = %v, want the subtree text", got)
	}
}

// TestSelectFallsBackForAnAnonymousBox checks that a box without a trusted
// display index uses the subtree even when ownership is available.
func TestSelectFallsBackForAnAnonymousBox(t *testing.T) {
	d := selectDisplay([]int{0, 0, 0, 0, 1})
	parent := d.Boxes[0]
	if got := Select(d, parent, -1, KindFill); !slices.Equal(got, []int{0, 4}) {
		t.Fatalf("anonymous fill = %v, want the subtree fills", got)
	}
}
