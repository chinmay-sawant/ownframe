package replay

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// TestChunkPlanSkipsIneligibleOps pins the run breaks: fixed, z-indexed, and
// translucent ops stay direct, and an opaque pair still folds.
func TestChunkPlanSkipsIneligibleOps(t *testing.T) {
	d := &layout.Display{
		Width: 400, Height: 200, PointsPerPixel: 1,
		Boxes: []layout.Box{{X: 10, Y: 10, W: 100, H: 120}},
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 12, Y: 12, W: 40, H: 20},
			{Kind: layout.DisplayOpFillRect, X: 12, Y: 36, W: 40, H: 20},
			{Kind: layout.DisplayOpFillRect, X: 12, Y: 60, W: 40, H: 20, Alpha: 0.5},
			{Kind: layout.DisplayOpFillRect, X: 12, Y: 84, W: 40, H: 20, Fixed: true},
			{Kind: layout.DisplayOpFillRect, X: 12, Y: 108, W: 40, H: 10, ZIndex: 2, ZIndexSet: true},
		},
		Order: []int{0, 1, 2, 3, 4},
	}

	p := buildChunks(d, buildPlan(d))
	if len(p.items) != 4 {
		t.Fatalf("items = %d, want 4", len(p.items))
	}

	if p.items[0].run == nil || len(p.items[0].run.ops) != 2 {
		t.Fatalf("first item = %+v, want a two-op run", p.items[0])
	}

	for i := 1; i < len(p.items); i++ {
		if p.items[i].op == nil {
			t.Fatalf("item %d is not direct", i)
		}
	}
}

// TestChunkPlanSkipsOversizedOps checks an op wider than the side cap stays
// direct.
func TestChunkPlanSkipsOversizedOps(t *testing.T) {
	d := &layout.Display{
		Width: 900, Height: 200, PointsPerPixel: 1,
		Boxes: []layout.Box{{X: 10, Y: 10, W: 800, H: 100}},
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 12, Y: 12, W: 600, H: 20},
			{Kind: layout.DisplayOpFillRect, X: 12, Y: 40, W: 600, H: 20},
		},
		Order: []int{0, 1},
	}

	p := buildChunks(d, buildPlan(d))
	if len(p.items) != 2 {
		t.Fatalf("items = %d, want 2 direct ops", len(p.items))
	}

	if p.items[0].op == nil || p.items[1].op == nil {
		t.Fatal("oversized ops folded into a run")
	}
}
