package replay

import (
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

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
	if matrix, ok := textProjection(op); ok {
		return matrixBounds(op, ppt, matrix)
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
