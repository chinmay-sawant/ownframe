package window

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// devCullDisplay is one fill inside the window, one under the dock, one below
// the window, one off the left edge, and one noop without paint.
func devCullDisplay() *layout.Display {
	return &layout.Display{
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 40, H: 20},
			{Kind: layout.DisplayOpFillRect, X: 5000, Y: 10, W: 40, H: 20},
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 5000, W: 40, H: 20},
			{Kind: layout.DisplayOpFillRect, X: -5000, Y: 10, W: 40, H: 20},
			{Kind: layout.DisplayOpNoop, W: 4, H: 4},
		},
		Order:          []int{0, 1, 2, 3, 4},
		Width:          800,
		Height:         600,
		PointsPerPixel: 0.75,
	}
}

// TestDevOpsOutlineCull checks the outlines are stroked only for ops that
// meet the visible page area, while the row list keeps every op so a culled
// row still picks its index.
func TestDevOpsOutlineCull(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.display = devCullDisplay()
	s.dev.tab = devTabOps
	s.dev.ops = true
	s.devLayout()

	var got []devRect
	s.dev.stroke = func(_ *ebiten.Image, r devRect, _ color.RGBA) {
		got = append(got, r)
	}

	s.drawDevOps(ebiten.NewImage(s.screenW, s.screenH))

	op := &s.display.Ops[0]
	want := s.devScreen(devOpRect(op, s.display.PointsPerPixel))
	if len(got) != 1 || got[0] != want {
		t.Fatalf("stroked outlines = %v, want only %v", got, want)
	}

	picks := 0

	for _, row := range s.devOpsRows() {
		if row.act == devActOp {
			picks++
		}
	}

	if picks != 4 {
		t.Fatalf("pickable rows = %d, want all 4 painted ops", picks)
	}
}
