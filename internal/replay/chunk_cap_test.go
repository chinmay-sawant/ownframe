package replay

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// TestChunkPlanSplitsPastOpCap checks a run longer than the op cap splits
// into runs that both fit.
func TestChunkPlanSplitsPastOpCap(t *testing.T) {
	d := &layout.Display{
		Width: 400, Height: 200, PointsPerPixel: 1,
		Boxes: []layout.Box{{X: 10, Y: 10, W: 300, H: 100}},
	}

	for i := range chunkMaxOps + 6 {
		d.Ops = append(d.Ops, layout.DisplayOp{
			Kind: layout.DisplayOpFillRect, X: 12 + float64(i), Y: 12, W: 20, H: 20,
		})
		d.Order = append(d.Order, i)
	}

	p := buildChunks(d, buildPlan(d))
	if len(p.items) != 2 {
		t.Fatalf("items = %d, want 2 runs", len(p.items))
	}

	if got := len(p.items[0].run.ops); got != chunkMaxOps {
		t.Fatalf("first run = %d ops, want %d", got, chunkMaxOps)
	}

	if got := len(p.items[1].run.ops); got != 6 {
		t.Fatalf("second run = %d ops, want 6", got)
	}
}
