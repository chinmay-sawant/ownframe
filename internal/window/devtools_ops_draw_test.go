package window

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// devBigDisplay is n painted fills in paint order, enough rows to scroll.
func devBigDisplay(n int) *layout.Display {
	ops := make([]layout.DisplayOp, n)
	order := make([]int, n)

	for i := range ops {
		ops[i] = layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: float64(10 + i), Y: 10, W: 40, H: 20}
		order[i] = i
	}

	return &layout.Display{Ops: ops, Order: order, Width: 800, Height: 600, PointsPerPixel: 0.75}
}

// TestDevOpsDrawRebuildsOnNewGeneration draws the Ops tab twice on a still
// generation and once per run on a moving one: the still draw rebuilds no
// rows, so the two runs differ by the whole list build. The rest of the draw
// is Ebiten text rasterization, outside the tab's control.
func TestDevOpsDrawRebuildsOnNewGeneration(t *testing.T) {
	d := newDevScreen()
	d.redraws = 1
	s := newDevShell(d)
	s.display = devBigDisplay(400)
	s.dev.tab = devTabOps

	img := ebiten.NewImage(s.screenW, s.screenH)
	s.drawDevTools(img)

	still := testing.AllocsPerRun(30, func() { s.drawDevTools(img) })
	rebuilt := testing.AllocsPerRun(30, func() {
		d.redraws++
		s.drawDevTools(img)
	})

	if rebuilt-still < 100 {
		t.Fatalf("still = %v allocs, rebuilt = %v, want the build in the gap", still, rebuilt)
	}
}
