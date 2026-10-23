package window

import (
	"github.com/chinmay-sawant/blinkless/layout"
)

const pinLayer = 2

type pinScreen struct {
	*dirtyScreen
}

func (p *pinScreen) ViewportPinZ() int { return pinLayer }

func pinOp(x, y, w, h, r, g, b float64, z int) layout.DisplayOp {
	op := layout.DisplayOp{
		Kind: layout.DisplayOpFillRect, X: x, Y: y, W: w, H: h,
		R: r, G: g, B: b, Alpha: 1,
	}
	if z > 0 {
		op.ZIndex, op.ZIndexSet = z, true
	}

	return op
}

func pinDisplay() *layout.Display {
	return &layout.Display{
		Width: 100, Height: 200, PointsPerPixel: 1,
		Ops: []layout.DisplayOp{
			pinOp(0, 0, 100, 400, 1, 1, 1, 0),
			pinOp(10, 80, 80, 40, 0, 1, 0, 0),
			pinOp(0, 0, 100, 40, 1, 0, 0, pinLayer),
			pinOp(0, 160, 100, 40, 0, 0, 1, pinLayer),
		},
		Order: []int{0, 1, 2, 3},
		Boxes: []layout.Box{{X: 0, Y: 0, W: 100, H: 400}},
	}
}

func newPinShell() (*pinScreen, *shell) {
	d := pinDisplay()
	screen := &pinScreen{dirtyScreen: &dirtyScreen{
		fakeScreen: &fakeScreen{width: 100, height: 200},
		display:    d, gen: 1,
	}}
	s := newDirtyShell(screen.dirtyScreen)
	s.app, s.display = screen, d
	s.screenW, s.screenH = 100, 200

	return screen, s
}
