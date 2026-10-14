package replay

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// drawOp replays one operation. Kinds that carry no paint are ignored.
// A viewport-fixed op ignores the scroll translation: the engine placed it
// against the viewport, so it stays on screen while the page moves. An op
// with a non-identity transform draws through its baked matrix; an image
// already places itself through its own transform.
func drawOp(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	if op.Fixed {
		dx, dy = 0, 0
	}

	if op.XformSet && op.Kind != layout.DisplayOpImage &&
		!op.Transform().IsIdentity() {
		drawXform(dst, op, dx, dy)

		return
	}

	switch op.Kind {
	case layout.DisplayOpFillRect:
		fillRect(dst, op, dx, dy)
	case layout.DisplayOpStrokeRect:
		drawStrokeRect(dst, op, dx, dy)
	case layout.DisplayOpImage:
		drawImage(dst, op, dx, dy)
	case layout.DisplayOpLine:
		drawLine(dst, op, dx, dy)
	case layout.DisplayOpGridRun:
		drawGrid(dst, op, dx, dy)
	case layout.DisplayOpText, layout.DisplayOpBullet:
		drawText(dst, op, dx, dy)
	}
}
