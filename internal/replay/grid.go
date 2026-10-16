package replay

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
)

// drawGrid replays one collapsed table grid as its line segments, in
// emission order. Each segment inherits the owner's opacity, transform, and
// fixed placement through masking.GridLine, matching the bitmap painter.
func drawGrid(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	if op.Grid == nil {
		return
	}

	for i := range op.Grid.Segs {
		line := masking.GridLine(op, i)

		drawLine(dst, &line, dx, dy)
	}
}
