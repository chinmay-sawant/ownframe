package replay

import (
	"image/color"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

// pxPerPt is the CSS pixel length of one layout point at zoom 1.
const pxPerPt = 96.0 / 72.0

// rgba converts an op's color and effective alpha to a color.RGBA. Opacity
// folds element opacity and the op alpha exactly as the engine's painters do.
func rgba(op *layout.DisplayOp) color.RGBA {
	return withAlpha(op, op.Opacity())
}

// withAlpha keeps the op color and replaces the alpha, clamped to 0..1.
func withAlpha(op *layout.DisplayOp, alpha float64) color.RGBA {
	return color.RGBA{
		R: channel(op.R * clamp(alpha)),
		G: channel(op.G * clamp(alpha)),
		B: channel(op.B * clamp(alpha)),
		A: uint8(math.Round(clamp(alpha) * 255)),
	}
}

func channel(value float64) uint8 {
	if value >= 1 {
		return 255
	}

	if value <= 0 {
		return 0
	}

	return uint8(math.Round(value * 255))
}

func clamp(value float64) float64 {
	if value < 0 {
		return 0
	}

	if value > 1 {
		return 1
	}

	return value
}
