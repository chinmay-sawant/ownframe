package replay

import (
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// DrawVisible skips offscreen operations before text shaping or GPU
// submission. Unbounded ink is retained. Translation maps the target back
// into page space. Blend and isolation groups composite as a unit.
func (renderer *Renderer) DrawVisible(dst *ebiten.Image, display *layout.Display, dx, dy float64) {
	renderer.DrawVisiblePinned(dst, display, dx, dy, 0, 0, 0)
}

// DrawVisiblePinned is DrawVisible with a viewport-pinned layer: operations
// whose z-index is at least pinZ are drawn at pinDx, pinDy instead of the
// scroll translation, so they stay at the viewport without a redraw.
func (renderer *Renderer) DrawVisiblePinned(dst *ebiten.Image, display *layout.Display, dx, dy float64, pinZ int, pinDx, pinDy float64) {
	if dst == nil || display == nil {
		return
	}

	rect := visibleRect(dst.Bounds(), dx, dy)
	pinRect := visibleRect(dst.Bounds(), pinDx, pinDy)
	ppt := display.PointsPerPixel

	renderer.paintChunked(dst, display, func(op *layout.DisplayOp) (float64, float64, bool) {
		odx, ody, r := dx, dy, rect
		if pinZ > 0 && op.ZIndexSet && op.ZIndex >= pinZ {
			odx, ody, r = pinDx, pinDy, pinRect
		}

		if op.Fixed {
			// A fixed op ignores the page scroll, including when a chunk
			// run batches several fixed ops into one blit. The per-op
			// drawOp path resets the translation itself, but a run blits
			// at the translation the picker returned, so return zero here
			// as well or the chrome jumps by the scroll offset.
			r = dst.Bounds()
			odx, ody = 0, 0
		}

		if !opTouches(op, ppt, r) {
			return 0, 0, false
		}

		return odx, ody, true
	})
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
