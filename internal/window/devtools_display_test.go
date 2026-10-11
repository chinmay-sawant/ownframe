package window

import "github.com/chinmay-sawant/blinkless/layout"

// devTestDisplay is a small display list with one fill and one text run.
func devTestDisplay() *layout.Display {
	return &layout.Display{
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 100, H: 50},
			{Kind: layout.DisplayOpText, X: 10, Y: 30, W: 80, H: 16, Text: "hi"},
			{Kind: layout.DisplayOpNoop, X: 0, Y: 0, W: 1, H: 1},
		},
		Order:          []int{0, 1, 2},
		Width:          800,
		Height:         600,
		PointsPerPixel: 0.75,
	}
}
