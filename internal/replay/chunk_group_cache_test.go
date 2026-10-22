package replay

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// groupedDisplay builds a display whose chunkable operations sit inside an
// isolation group. The engine marks a grouped operation as z-indexed, so the
// display is built by hand.
func groupedDisplay() *layout.Display {
	group := &layout.DisplayGroup{ID: 1, Isolate: true}
	ops := []layout.DisplayOp{
		{Kind: layout.DisplayOpFillRect, X: 12, Y: 12, W: 40, H: 20},
		{Kind: layout.DisplayOpFillRect, X: 12, Y: 36, W: 40, H: 20},
	}
	for i := range ops {
		ops[i].SetBlendGroup(group)
	}

	return &layout.Display{
		Width: 200, Height: 100, PointsPerPixel: 1,
		Boxes: []layout.Box{{X: 10, Y: 10, W: 50, H: 80}},
		Ops:   ops, Order: []int{0, 1},
	}
}

// firstRun returns the plan's first chunk run.
func firstRun(p *chunkPlan) *chunkRun {
	for i := range p.items {
		if p.items[i].run != nil {
			return p.items[i].run
		}
	}

	return nil
}

// TestChunkGroupRunSurvivesRedraw pins that a chunk holding an operation
// inside a group is reused after a redraw. The group is a fresh allocation
// on every layout run, so a signature folding its address missed every
// redraw.
func TestChunkGroupRunSurvivesRedraw(t *testing.T) {
	flushChunks()
	defer flushChunks()

	first := groupedDisplay()
	plan := buildChunks(first, buildPlan(first))
	run := firstRun(plan)
	if run == nil {
		t.Fatal("no run holds the grouped operations")
	}

	baked := chunkRaster(first, plan, run)
	if baked == nil {
		t.Fatal("the grouped run did not bake")
	}

	second := groupedDisplay()
	plan2 := buildChunks(second, buildPlan(second))
	run2 := firstRun(plan2)
	if run2 == nil {
		t.Fatal("the redraw has no run holding the grouped operations")
	}

	if run2.ops[0].Group() == run.ops[0].Group() {
		t.Fatal("the redraw reused the group allocation")
	}

	if again := chunkRaster(second, plan2, run2); again != baked {
		t.Fatal("the grouped chunk was baked again after the redraw")
	}
}
