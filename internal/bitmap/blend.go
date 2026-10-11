package bitmap

import (
	"image"
	"image/color"
)

func blendSupported(mode string) bool {
	switch mode {
	case "", "normal", "multiply", "screen", "overlay", "darken", "lighten",
		"color-dodge", "color-burn", "hard-light", "soft-light", "difference", "exclusion",
		"hue", "saturation", "color", "luminosity":
		return true
	}
	return false
}

func composite(dst, src *image.NRGBA, mode string) {
	for y := dst.Bounds().Min.Y; y < dst.Bounds().Max.Y; y++ {
		for x := dst.Bounds().Min.X; x < dst.Bounds().Max.X; x++ {
			s, b := src.NRGBAAt(x, y), dst.NRGBAAt(x, y)
			if s.A == 0 {
				continue
			}
			dst.SetNRGBA(x, y, blendPixel(b, s, mode))
		}
	}
}

func blendPixel(b, s color.NRGBA, mode string) color.NRGBA {
	if mode == "" || mode == "normal" {
		return over(b, s)
	}
	cb, cs := rgbColor(b), rgbColor(s)
	blended := blendRGB(cb, cs, mode)
	ab, as := float64(b.A)/255, float64(s.A)/255
	alpha := as + ab*(1-as)
	out := [3]uint8{}
	for i := range out {
		out[i] = channel((as*(1-ab)*cs[i] + as*ab*blended[i] + (1-as)*ab*cb[i]) / alpha)
	}
	return color.NRGBA{R: out[0], G: out[1], B: out[2], A: channel(alpha)}
}

func rgbColor(c color.NRGBA) [3]float64 {
	return [3]float64{float64(c.R) / 255, float64(c.G) / 255, float64(c.B) / 255}
}
