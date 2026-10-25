package replay

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// drawOp replays one operation. Kinds that carry no paint are ignored.
// A viewport-fixed op ignores the scroll translation: the engine placed it
// against the viewport, so it stays on screen while the page moves. An op
// with a non-identity transform draws through its baked matrix; an image
// already places itself through its own transform, and a corner path maps
// each command point through the matrix as it builds the vector path.
func (renderer *Renderer) drawOp(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	if op.Fixed {
		dx, dy = 0, 0
	}
	if _, ok := textProjection(op); ok {
		renderer.drawXformText(dst, op, dx, dy)
		return
	}

	if op.XformSet && op.Kind != layout.DisplayOpImage && op.Kind != layout.DisplayOpPath &&
		!op.Transform().IsIdentity() {
		renderer.drawXform(dst, op, dx, dy)

		return
	}

	switch op.Kind {
	case layout.DisplayOpFillRect:
		renderer.fillRect(dst, op, dx, dy)
	case layout.DisplayOpStrokeRect:
		renderer.drawStrokeRect(dst, op, dx, dy)
	case layout.DisplayOpPath:
		drawPath(dst, op, dx, dy)
	case layout.DisplayOpImage:
		renderer.drawImage(dst, op, dx, dy)
	case layout.DisplayOpLine:
		drawLine(dst, op, dx, dy)
	case layout.DisplayOpGridRun:
		drawGrid(dst, op, dx, dy)
	case layout.DisplayOpText, layout.DisplayOpBullet:
		renderer.drawText(dst, op, dx, dy)
	}
}
