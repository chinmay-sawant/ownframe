package replay

import (
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

// chunkCoreBox is the op's layout box in CSS pixels, before stroke growth:
// the geometry the element box should contain. The painted box can reach
// half a border width outside it, so assignment tests this box and the bake
// still covers opBounds.
func chunkCoreBox(op *layout.DisplayOp, ppt float64) (image.Rectangle, bool) {
	if op.XformSet && !op.Transform().IsIdentity() {
		return xformBounds(op, ppt)
	}

	switch op.Kind {
	case layout.DisplayOpText, layout.DisplayOpBullet:
		return textBounds(op, ppt), true
	case layout.DisplayOpLine:
		x, y, w, h, _ := op.PaintLineGeometry()

		return boxBounds(min(x, x+w), min(y, y+h), math.Abs(w), math.Abs(h), 0, ppt), true
	case layout.DisplayOpGridRun:
		return gridBounds(op, ppt), true
	case layout.DisplayOpFillRect, layout.DisplayOpImage, layout.DisplayOpStrokeRect:
		return boxBounds(op.X, op.Y, op.W, op.H, 0, ppt), true
	}

	return image.Rectangle{}, false
}

// chunkBoxContains reports whether box holds inner within the tolerance.
func chunkBoxContains(box, inner image.Rectangle) bool {
	return inner.Min.X >= box.Min.X-chunkTol && inner.Min.Y >= box.Min.Y-chunkTol &&
		inner.Max.X <= box.Max.X+chunkTol && inner.Max.Y <= box.Max.Y+chunkTol
}
