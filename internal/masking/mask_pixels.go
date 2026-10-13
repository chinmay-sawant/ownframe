package masking

import (
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
)

func paintMaskPixels(dst []uint8, bw, bh int, src *image.NRGBA,
	px, py float64, layer layout.MaskLayer) []uint8 {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	rx, ry := maskRepeat(layer.Repeat)
	for y := 0; y < bh; y++ {
		sy := float64(y) - py
		if ry {
			sy = positiveMod(sy, float64(h))
		} else if sy < 0 || sy >= float64(h) {
			continue
		}
		for x := 0; x < bw; x++ {
			sx := float64(x) - px
			if rx {
				sx = positiveMod(sx, float64(w))
			} else if sx < 0 || sx >= float64(w) {
				continue
			}
			p := src.NRGBAAt(int(sx), int(sy))
			if !layer.Preprocessed && strings.TrimSpace(layer.Mode) == "luminance" {
				dst[y*bw+x] = maskLuminance(p)
			} else {
				dst[y*bw+x] = p.A
			}
		}
	}
	return dst
}

func maskRepeat(value string) (bool, bool) {
	switch strings.TrimSpace(value) {
	case "repeat":
		return true, true
	case "repeat-x":
		return true, false
	case "repeat-y":
		return false, true
	default:
		return false, false
	}
}

func positiveMod(value, period float64) float64 {
	value = math.Mod(value, period)
	if value < 0 {
		value += period
	}
	return value
}

func maskLuminance(p color.NRGBA) uint8 {
	linear := func(c uint8) float64 {
		v := float64(c) / 255
		if v <= .04045 {
			return v / 12.92
		}
		return math.Pow((v+.055)/1.055, 2.4)
	}
	y := .2126*linear(p.R) + .7152*linear(p.G) + .0722*linear(p.B)
	return channel(y * float64(p.A) / 255)
}

func channel(value float64) uint8 {
	return uint8(math.Round(max(0, min(1, value)) * 255))
}
