package replay

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// drawDirectRect is the pre-chunk damage path: filter the order and draw
// every op that touches the rect.
func drawDirectRect(dst *ebiten.Image, d *layout.Display, rect image.Rectangle) {
	ppt := d.PointsPerPixel

	for _, i := range d.Order {
		op := &d.Ops[i]
		if op.Fixed || !opTouches(op, ppt, rect) {
			continue
		}

		countOp()
		drawOp(dst, op, 0, 0)
	}
}
