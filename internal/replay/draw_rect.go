package replay

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// DrawRect replays the operations of display whose painted box meets rect,
// in the same paint order Draw uses. rect is in CSS pixels in the page's own
// space; dx and dy only move the paint, as they do in Draw. The rect is
// clamped to the content, so one that reaches outside still paints the part
// that lands inside.
//
// The ops draw into a sub-image of dst, so an op larger than the rect (the
// page background fill, for one) is clipped to the rect instead of painting
// over the rest of the buffer. A blend or isolation group composites inside
// that sub-image too, so a partial repaint matches a full one.
func DrawRect(dst *ebiten.Image, display *layout.Display, rect image.Rectangle, dx, dy float64) {
	if display == nil || dst == nil {
		return
	}

	rect = clipRect(display, rect).Intersect(dst.Bounds())
	if rect.Empty() {
		return
	}

	clip, ok := dst.SubImage(rect).(*ebiten.Image)
	if !ok {
		return
	}

	ppt := display.PointsPerPixel
	paintChunked(clip, display, func(op *layout.DisplayOp) (float64, float64, bool) {
		if op.Fixed || !opTouches(op, ppt, rect) {
			return 0, 0, false
		}

		return dx, dy, true
	})
}

// canvasRect is the page canvas in CSS pixels.
func canvasRect(display *layout.Display) image.Rectangle {
	if display.Width <= 0 || display.Height <= 0 {
		return image.Rectangle{}
	}

	return image.Rect(0, 0, display.Width, display.Height)
}

// clipRect clamps rect to the page content.
func clipRect(display *layout.Display, rect image.Rectangle) image.Rectangle {
	return rect.Intersect(contentRect(display))
}

// opTouches reports whether an op's painted box meets rect. An op whose ink
// cannot be bounded, such as a rotated run, always touches.
func opTouches(op *layout.DisplayOp, ppt float64, rect image.Rectangle) bool {
	box, ok := opBounds(op, ppt)
	if !ok {
		return true
	}

	return !box.Intersect(rect).Empty()
}
