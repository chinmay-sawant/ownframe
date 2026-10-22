package replay

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

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
