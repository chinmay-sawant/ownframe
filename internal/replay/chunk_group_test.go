package replay

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// TestChunkPlanKeepsGroupsWhole pins that a blend group is one plan item and
// the runs on either side do not reach into it.
func TestChunkPlanKeepsGroupsWhole(t *testing.T) {
	group := &layout.DisplayGroup{ID: 1, Mode: "multiply", Isolate: true}
	ops := []layout.DisplayOp{
		{Kind: layout.DisplayOpFillRect, X: 12, Y: 12, W: 40, H: 20},
		{Kind: layout.DisplayOpFillRect, X: 12, Y: 36, W: 40, H: 20},
		{Kind: layout.DisplayOpFillRect, X: 60, Y: 12, W: 20, H: 20},
		{Kind: layout.DisplayOpFillRect, X: 12, Y: 60, W: 40, H: 20},
	}
	ops[2].SetBlendGroup(group)

	d := &layout.Display{
		Width: 200, Height: 100, PointsPerPixel: 1,
		Boxes: []layout.Box{{X: 10, Y: 10, W: 50, H: 80}},
		Ops:   ops, Order: []int{0, 1, 2, 3},
	}

	p := buildChunks(d, buildPlan(d))
	if len(p.items) != 3 {
		t.Fatalf("items = %d, want run, group, op", len(p.items))
	}

	if p.items[0].run == nil || len(p.items[0].run.ops) != 2 {
		t.Fatalf("first item = %+v, want a two-op run", p.items[0])
	}

	if p.items[1].group == nil || p.items[1].group.group != group {
		t.Fatalf("second item = %+v, want the group node", p.items[1])
	}

	if p.items[2].op == nil {
		t.Fatal("the op after the group is not direct")
	}
}
