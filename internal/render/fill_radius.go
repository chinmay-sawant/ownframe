package render

import (
	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/masking"
)

// RadiiXY resolves a display fill or stroke's corner radii on both axes.
// The geometry lives in masking so the bitmap painter needs no render
// import; this wrapper keeps existing replay callers compiling.
func RadiiXY(op *layout.DisplayOp) (rx, ry [4]float64) {
	return masking.RadiiXY(op)
}

// FillRadii resolves a display fill's circular corner radii. ok is false when
// a corner is elliptical, because the fill replay draws circular arcs only.
func FillRadii(op *layout.DisplayOp) (radii [4]float64, ok bool) {
	rx, ry := masking.RadiiXY(op)

	for i := range rx {
		if rx[i] != ry[i] {
			return [4]float64{}, false
		}
	}

	return rx, true
}
