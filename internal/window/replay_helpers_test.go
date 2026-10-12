package window

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// partialDisplay is a small replayable page with one box.
func partialDisplay(w, h int) *layout.Display {
	return &layout.Display{
		Width: w, Height: h, PointsPerPixel: 1,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: float64(w), H: float64(h), R: 1, G: 1, B: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, R: 1, Alpha: 1},
		},
		Order: []int{0, 1},
	}
}

func partialBuf(w, h int) *ebiten.Image {
	return ebiten.NewImage(w, h)
}
