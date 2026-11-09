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
//
// The plan comes from the generation-keyed tile cache: candidates come from
// the tile buckets built once at plan build instead of scanning the whole
// plan per tile, and the rect paints once across the tiles it covers.
func (renderer *Renderer) DrawRect(dst *ebiten.Image, display *layout.Display, rect image.Rectangle, dx, dy float64) {
	renderer.DrawRects(dst, display, []image.Rectangle{rect}, dx, dy)
}

// DrawRects replays the operations meeting any of rects, painting each dirty
// rect once across the tiles it covers. Rects are in CSS pixels in the
// page's space; dx and dy move the paint. Each rect paints into its own
// sub-image, so a large op is clipped per rect and a partial repaint matches
// a full one. Candidates come from the tile broadphase built once at plan
// build; the precise ink check still uses opTouches per rect.
func (renderer *Renderer) DrawRects(dst *ebiten.Image, display *layout.Display, rects []image.Rectangle, dx, dy float64) {
	if display == nil || dst == nil || len(rects) == 0 {
		return
	}

	p := renderer.tilePlanFor(display)
	if p == nil {
		// The plan cache evicted between the group build and the chunk
		// fold; keep the direct walk so damage still paints.
		for _, rect := range rects {
			rect = clipRect(display, rect).Intersect(dst.Bounds())
			if rect.Empty() {
				continue
			}

			clip, ok := dst.SubImage(rect).(*ebiten.Image)
			if !ok {
				continue
			}

			ppt := display.PointsPerPixel
			renderer.paintChunked(clip, display, func(op *layout.DisplayOp) (float64, float64, bool) {
				if op.Fixed || !opTouches(op, ppt, rect) {
					return 0, 0, false
				}

				return dx, dy, true
			})
		}

		return
	}

	renderer.paintTileRects(dst, display, p, rects, dx, dy, false)
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
	if op.IsGroupBegin() && op.Group() != nil && op.Group().BackdropFilter != "" {
		box := boxBounds(op.X, op.Y, op.W, op.H, 0, ppt)
		return !box.Intersect(rect).Empty()
	}

	box, ok := opBounds(op, ppt)
	if !ok {
		return true
	}

	return !box.Intersect(rect).Empty()
}
