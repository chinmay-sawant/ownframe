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

// withAlpha keeps the op color and replaces the alpha, clamped to 0..1.
func withAlpha(op *layout.DisplayOp, alpha float64) color.RGBA {
	r, g, b, _ := masking.PaintColor(op)
	a := masking.Clamp01(alpha)
	return color.RGBA{
		R: masking.Channel(r * a),
		G: masking.Channel(g * a),
		B: masking.Channel(b * a),
		A: masking.Channel(a),
	}
}

func channel(value float64) uint8 {
	return masking.Channel(value)
}

func clamp(value float64) float64 {
	return masking.Clamp01(value)
}
