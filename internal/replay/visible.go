package replay

import (
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// DrawVisible skips offscreen operations before text shaping or GPU submission.
// Unbounded ink is retained. Translation maps the target back into page space.
func DrawVisible(dst *ebiten.Image, display *layout.Display, dx, dy float64) {
	DrawVisiblePinned(dst, display, dx, dy, 0, 0, 0)
}

// DrawVisiblePinned is DrawVisible with a viewport-pinned layer: operations
// whose z-index is at least pinZ are drawn at pinDx, pinDy instead of the
// scroll translation, so they stay at the viewport without a redraw.
func DrawVisiblePinned(dst *ebiten.Image, display *layout.Display, dx, dy float64, pinZ int, pinDx, pinDy float64) {
	if dst == nil || display == nil {
		return
	}

	rect := visibleRect(dst.Bounds(), dx, dy)
	pinRect := visibleRect(dst.Bounds(), pinDx, pinDy)

	for _, index := range display.Order {
		if index < 0 || index >= len(display.Ops) {
			continue
		}

		op := &display.Ops[index]
		odx, ody, r := dx, dy, rect
		if pinZ > 0 && op.ZIndexSet && op.ZIndex >= pinZ {
			odx, ody, r = pinDx, pinDy, pinRect
		}

		if op.Fixed {
			r = dst.Bounds()
		}
		if opTouches(op, display.PointsPerPixel, r) {
			drawOp(dst, op, odx, ody)
		}
	}
}

func visibleRect(bounds image.Rectangle, dx, dy float64) image.Rectangle {
	return image.Rect(
		int(math.Floor(float64(bounds.Min.X)-dx)),
		int(math.Floor(float64(bounds.Min.Y)-dy)),
		int(math.Ceil(float64(bounds.Max.X)-dx)),
		int(math.Ceil(float64(bounds.Max.Y)-dy)),
	)
}

func visitVisible(display *layout.Display, rect image.Rectangle, draw func(int)) {
	for _, index := range display.Order {
		if index < 0 || index >= len(display.Ops) {
			continue
		}
		if opTouches(&display.Ops[index], display.PointsPerPixel, rect) {
			draw(index)
		}
	}
}
