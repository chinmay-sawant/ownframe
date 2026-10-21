package replay

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

// textBounds is the line box around a baseline. A text op carries its
// baseline in Y, so the box starts one font ascent above it and ends one ink
// descent below it, the same rule internal/frame uses to find a run.
func textBounds(op *layout.DisplayOp, ppt float64) image.Rectangle {
	top := op.Y - textAscent(op)
	bottom := op.Y + op.InkDescent
	if bottom < top {
		bottom = top
	}

	return boxBounds(op.X, top, op.W, bottom-top, 0, ppt)
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
