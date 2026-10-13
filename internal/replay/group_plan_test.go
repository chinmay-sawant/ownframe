package replay

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// TestPlanBucketsNestedGroups pins the group tree: a child group's node sits
// inside its parent's items where the child first paints, whatever order the
// boundary markers take.
func TestPlanBucketsNestedGroups(t *testing.T) {
	t.Parallel()

	outer := &layout.DisplayGroup{ID: 1, Mode: "multiply", Isolate: true}
	inner := &layout.DisplayGroup{ID: 2, Mode: "screen", Isolate: true, Parent: outer}

	rootOp := layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 50, H: 20}
	first := layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 10, H: 10}
	first.SetBlendGroup(outer)
	second := layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 20, Y: 0, W: 10, H: 10}
	second.SetBlendGroup(inner)

	d := &layout.Display{PointsPerPixel: 1, Ops: []layout.DisplayOp{rootOp, first, second}, Order: []int{0, 1, 2}}
	plan := buildPlan(d)

	if len(plan.items) != 2 || plan.items[0].op == nil || plan.items[1].child == nil {
		t.Fatalf("root items = %v", plan.items)
	}

	outerNode := plan.items[1].child
	if outerNode.group != outer || len(outerNode.items) != 2 {
		t.Fatalf("outer node = %+v", outerNode)
	}

	innerNode := outerNode.items[1].child
	if innerNode == nil || innerNode.group != inner || len(innerNode.items) != 1 {
		t.Fatalf("inner node = %+v", innerNode)
	}

	box, ok := nodeBox(outerNode, 1)
	if !ok || box != image.Rect(-1, -1, 31, 11) {
		t.Fatalf("group box = %v, %v", box, ok)
	}
}
