package bitmap

import (
	"image/color"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

func channel(v float64) uint8 {
	return uint8(math.Round(max(0, min(1, v)) * 255))
}

func opColor(op *layout.DisplayOp) color.NRGBA {
	return color.NRGBA{R: channel(op.R), G: channel(op.G), B: channel(op.B), A: channel(op.Opacity())}
}

func over(dst, src color.NRGBA) color.NRGBA {
	a, b := float64(src.A)/255, float64(dst.A)/255
	out := a + b*(1-a)
	if out == 0 {
		return color.NRGBA{}
	}
	mix := func(s, d uint8) uint8 {
		return uint8(math.Round((float64(s)*a + float64(d)*b*(1-a)) / out))
	}
	return color.NRGBA{R: mix(src.R, dst.R), G: mix(src.G, dst.G), B: mix(src.B, dst.B), A: channel(out)}
}
