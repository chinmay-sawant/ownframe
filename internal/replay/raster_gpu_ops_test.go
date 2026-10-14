package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
)

// rasterParityOps are the shapes the parity test draws both ways.
func rasterParityOps() []layout.DisplayOp {
	return []layout.DisplayOp{
		{Kind: layout.DisplayOpFillRect, X: 10.5, Y: 20.25, W: 42.5, H: 30, Radius: 3, R: 1, G: 0.5},
		{Kind: layout.DisplayOpStrokeRect, X: 10.5, Y: 20.25, W: 42.5, H: 30, Width: 2, Radius: 3, B: 1},
		{Kind: layout.DisplayOpFillRect, X: 70.5, Y: 20.25, W: 42.5, H: 30, Radius: 3, R: 1, G: 0.5},
		{Kind: layout.DisplayOpStrokeRect, X: 70, Y: 20.25, W: 42.5, H: 30, Width: 2, StrokeMask: 5, B: 1},
		{Kind: layout.DisplayOpFillRect, X: 130.75, Y: 60.5, W: 20, H: 20, Radius: 6, Alpha: 0.5},
	}
}

// drawDirect draws one op through the uncached path only.
func drawDirect(op *layout.DisplayOp, dst *ebiten.Image) {
	switch op.Kind {
	case layout.DisplayOpFillRect:
		radii, ok := render.FillRadii(op)
		if !ok {
			return
		}
		fillRounded(dst, op, radii, 0, 0)
	case layout.DisplayOpStrokeRect:
		strokeDirect(dst, op, 0, 0)
	}
}
