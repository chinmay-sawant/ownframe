package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// orderPinScreen is a fakeScreen that reports a viewport pin z.
type orderPinScreen struct {
	fakeScreen
	pinZ int
}

func (p *orderPinScreen) ViewportPinZ() int { return p.pinZ }

func orderDisplay() *layout.Display {
	return &layout.Display{
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, ZIndexSet: true, ZIndex: 5},
			{Kind: layout.DisplayOpFillRect},
		},
		Order: []int{0, 1},
	}
}

func TestScanOrderRescansWhenPinZChanges(t *testing.T) {
	app := &orderPinScreen{}
	s := newShell(context.Background(), app)
	s.display = orderDisplay()

	if s.hasViewportOps() || !s.bufferOrderSafe() {
		t.Fatal("unpinned display reported viewport ops")
	}

	app.pinZ = 5
	if !s.hasViewportOps() {
		t.Fatal("order cache ignored the new viewport pin z")
	}
	if s.bufferOrderSafe() {
		t.Fatal("pinned layer before an unpinned one must make the buffer order unsafe")
	}
}

func TestScanOrderRescansWhenGenerationChanges(t *testing.T) {
	app := &orderPinScreen{pinZ: 5}
	s := newShell(context.Background(), app)
	s.display = orderDisplay()
	s.seq = 1
	if !s.hasViewportOps() {
		t.Fatal("pinned op not seen")
	}

	// Same pointer, new generation, ops rewritten in place.
	s.display.Ops[0].ZIndex = 0
	s.display.Ops[0].ZIndexSet = false
	s.seq = 2
	if s.hasViewportOps() {
		t.Fatal("order cache served a stale scan for a recycled display")
	}
}

func TestSetDisplayEvictsRecycledPointer(t *testing.T) {
	app := &orderPinScreen{}
	s := newShell(context.Background(), app)
	d := orderDisplay()
	s.setDisplay(d)
	s.renderer.CachePlan(d)
	if !s.renderer.PlanCached(d) {
		t.Fatal("plan not cached")
	}

	s.setDisplay(d)
	if s.renderer.PlanCached(d) {
		t.Fatal("same pointer, new generation kept the old plan")
	}
}
