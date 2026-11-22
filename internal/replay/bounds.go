package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"math"
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

// PaintBounds returns the op's painted box in CSS pixels, rounded outward,
// with no slack. opBounds is the same box grown by one pixel per side for the
// replay's dirty-rect filter; a reader that wants the ink itself, such as the
// window's inspector, uses this.
//
// The box follows the paint: a text or bullet run is bounded by its face
// ascent and ink descent around the baseline in Y, a line by its stroke width
// and inward geometry, a stroke rect by half its width outside the box, and a
// grid run by the union of its segments.
//
// ok is false when the box cannot be bounded, such as a transformed or
// rotated run; the caller must treat such an op as unbounded.
func PaintBounds(op *layout.DisplayOp, ppt float64) (image.Rectangle, bool) {
	box, ok := opBounds(op, ppt)
	if !ok || box.Empty() {
		return box, ok
	}

	// opBounds adds boundsPad after rounding outward, so taking the same
	// whole pixels back is exact.
	return box.Inset(boundsPad), ok
}

// boundsPad is one CSS pixel of slack per side. Rounded corners and
// antialiasing paint just outside the integer box, and drawing an op that
// only grazes the rect costs less than a missed pixel.
const boundsPad = 1

const halfDiv = 2

// opBounds returns the op's painted box in CSS pixels, rounded outward. ok is
// false when the box cannot be bounded; the caller must draw such an op
// rather than skip it. An op that paints nothing gets an empty box.
func opBounds(op *layout.DisplayOp, ppt float64) (image.Rectangle, bool) {
	if ppt <= 0 {
		return image.Rectangle{}, false
	}

	// A projected or quarter-turned run paints through its matrix, so its box
	// is the matrix box. Only text and bullets can carry either, and the scan
	// meets every op, so the cheap presence check stays inline: most runs
	// carry neither a projection payload nor a rotation.
	if (op.Kind == layout.DisplayOpText || op.Kind == layout.DisplayOpBullet) && !op.XformSet &&
		(op.RotateDeg != 0 || op.XformValue() != (layout.Matrix2D{})) {
		if m, ok := masking.TextMatrix(op); ok {
			return matrixBounds(op, ppt, m)
		}
	}

	if op.RotateDeg != 0 {
		return image.Rectangle{}, false
	}

	// A corner path bounds itself from its command points, through the rebase
	// onto the op box and the op matrix. The box corners xformBounds maps do
	// not cover a concave shaped outline.
	if op.Kind == layout.DisplayOpPath {
		return pathBounds(op, ppt), true
	}

	if op.XformSet && !op.Transform().IsIdentity() {
		return xformBounds(op, ppt)
	}

	switch op.Kind {
	case layout.DisplayOpText, layout.DisplayOpBullet:
		return textBounds(op, ppt), true
	case layout.DisplayOpLine:
		return lineBounds(op, ppt), true
	case layout.DisplayOpGridRun:
		return gridBounds(op, ppt), true
	case layout.DisplayOpFillRect, layout.DisplayOpImage:
		return boxBounds(op.X, op.Y, op.W, op.H, 0, ppt), true
	case layout.DisplayOpStrokeRect:
		return boxBounds(op.X, op.Y, op.W, op.H, op.Width/halfDiv, ppt), true
	}

	return image.Rectangle{}, true
}

// pathBounds converts a corner path's mapped ink box to a whole-pixel box.
// The command points bound the curves, because a quadratic or cubic stays
// inside the convex hull of its control points, and a stroke grows the box by
// half its width. An op with no command point paints nothing and gets an
// empty box.
func pathBounds(op *layout.DisplayOp, ppt float64) image.Rectangle {
	left, top, right, bottom, grow, ok := pathInkBox(op)
	if !ok {
		return image.Rectangle{}
	}

	return boxBounds(left, top, right-left, bottom-top, grow, ppt)
}

// boxBounds converts a point-space box to a whole-pixel box, grown by pad
// points on every side. display.PointsPerPixel converts CSS pixels to points,
// so the division below is the points-to-pixels direction.
func boxBounds(x, y, w, h, grow, ppt float64) image.Rectangle {
	pad := boundsPad + grow/ppt

	return image.Rect(
		int(math.Floor(x/ppt-pad)), int(math.Floor(y/ppt-pad)),
		int(math.Ceil((x+w)/ppt+pad)), int(math.Ceil((y+h)/ppt+pad)),
	)
}

// strokeHalfDiv halves a stroke to its center-line extent.
const strokeHalfDiv = 2

// lineBounds bounds one stroked segment. The engine centers the stroke on
// the line and extends it half a stroke past each endpoint. A diagonal
// strokes a square-capped segment, so its box unions both endpoints, each
// grown by half the stroke; image.Rectangle.Union keeps the bounding box,
// which covers the middle of the segment too.
func lineBounds(op *layout.DisplayOp, ppt float64) image.Rectangle {
	x, y, w, h, width := op.PaintLineGeometry()

	stroke := width
	if stroke <= 0 {
		stroke = ppt
	}

	half := stroke / strokeHalfDiv
	if w != 0 && h != 0 {
		start := boxBounds(x-half, y-half, stroke, stroke, 0, ppt)
		end := boxBounds(x+w-half, y+h-half, stroke, stroke, 0, ppt)

		return start.Union(end)
	}
	if h <= 0 {
		return boxBounds(x-half, y-half, w+stroke, stroke, 0, ppt)
	}

	return boxBounds(x-half, y-half, stroke, h+stroke, 0, ppt)
}

// gridBounds unions the boxes of a collapsed table grid's segments.
func gridBounds(op *layout.DisplayOp, ppt float64) image.Rectangle {
	if op.Grid == nil {
		return image.Rectangle{}
	}

	box := image.Rectangle{}

	for i := range op.Grid.Segs {
		seg := &op.Grid.Segs[i]
		line := layout.DisplayOp{
			X: seg.X, Y: seg.Y, W: seg.W, H: seg.H,
			Width: seg.Width, LineInset: seg.LineInset,
		}
		box = box.Union(lineBounds(&line, ppt))
	}

	return box
}

// textStrokePad returns the point-space radius added around a text run by
// -webkit-text-stroke, matching drawTextStroke.
func textStrokePad(op *layout.DisplayOp) float64 {
	if width, _, ok := op.TextStroke(); ok && width > 0 {
		return width / strokeHalfWidthDivisor
	}
	return 0
}

// textBounds is the line box around a baseline. A text op carries its
// baseline in Y, so the box starts one font ascent above it and ends one ink
// descent below it, the same rule internal/frame uses to find a run.
func textBounds(op *layout.DisplayOp, ppt float64) image.Rectangle {
	top := op.Y - textAscent(op)
	bottom := op.Y + op.InkDescent
	if bottom < top {
		bottom = top
	}

	return boxBounds(op.X, top, op.W, bottom-top, textStrokePad(op), ppt)
}

// fallbackAscentRatio estimates ascent from size without a face.
const fallbackAscentRatio = 0.8

// textAscent is the font ascent above the baseline. The op's line box height
// is the fallback when the face is absent.
func textAscent(op *layout.DisplayOp) float64 {
	if op.Font != nil && op.Font.UnitsPerEm() > 0 {
		return float64(op.Font.Ascent()) * op.Size / float64(op.Font.UnitsPerEm())
	}

	if op.H > 0 {
		return op.H - op.InkDescent
	}

	return op.Size * fallbackAscentRatio
}
