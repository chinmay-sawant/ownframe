package replay

import (
	"image/color"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
)

// pxPerPt is the CSS pixel length of one layout point at zoom 1.
const pxPerPt = 96.0 / 72.0

// rgba converts an op's color and effective alpha to a color.RGBA. Opacity
// folds element opacity and the op alpha exactly as the engine's painters do.
func rgba(op *layout.DisplayOp) color.RGBA {
	return masking.PaintRGBA(op)
}
