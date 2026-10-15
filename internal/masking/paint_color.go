package masking

import (
	"image/color"

	"github.com/chinmay-sawant/blinkless/layout"
)

func Clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func Channel(v float64) uint8 {
	return channel(Clamp01(v))
}

func PaintColor(op *layout.DisplayOp) (r, g, b, a float64) {
	if op == nil {
		return 0, 0, 0, 0
	}
	return Clamp01(op.R), Clamp01(op.G), Clamp01(op.B), Clamp01(op.Opacity())
}

func PaintNRGBA(op *layout.DisplayOp) color.NRGBA {
	r, g, b, a := PaintColor(op)
	return color.NRGBA{R: Channel(r), G: Channel(g), B: Channel(b), A: Channel(a)}
}

func PaintRGBA(op *layout.DisplayOp) color.RGBA {
	r, g, b, a := PaintColor(op)
	return color.RGBA{R: Channel(r * a), G: Channel(g * a), B: Channel(b * a), A: Channel(a)}
}
