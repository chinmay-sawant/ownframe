package bitmap

import (
	"image"
	"image/color"

	"github.com/chinmay-sawant/ownframe/internal/masking"
)

const nrgbaBytesPerPixel = 4

func blendSupported(mode string) bool {
	return masking.BlendSupported(mode)
}

func composite(dst, src *image.NRGBA, mode string) {
	bounds := dst.Bounds().Intersect(src.Bounds())
	if bounds.Empty() {
		return
	}
	normal := masking.BlendNormal(mode)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		sOff := src.PixOffset(bounds.Min.X, y)
		dOff := dst.PixOffset(bounds.Min.X, y)
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			sa := src.Pix[sOff+3]
			if sa != 0 {
				if normal && sa == channelMax {
					dst.Pix[dOff] = src.Pix[sOff]
					dst.Pix[dOff+1] = src.Pix[sOff+1]
					dst.Pix[dOff+2] = src.Pix[sOff+2]
					dst.Pix[dOff+3] = sa
				} else {
					s := color.NRGBA{R: src.Pix[sOff], G: src.Pix[sOff+1], B: src.Pix[sOff+2], A: sa}
					b := color.NRGBA{R: dst.Pix[dOff], G: dst.Pix[dOff+1], B: dst.Pix[dOff+2], A: dst.Pix[dOff+3]}
					out := blendPixel(b, s, mode)
					dst.Pix[dOff] = out.R
					dst.Pix[dOff+1] = out.G
					dst.Pix[dOff+2] = out.B
					dst.Pix[dOff+3] = out.A
				}
			}
			sOff += nrgbaBytesPerPixel
			dOff += nrgbaBytesPerPixel
		}
	}
}

func blendPixel(b, s color.NRGBA, mode string) color.NRGBA {
	if masking.BlendNormal(mode) {
		return over(b, s)
	}
	cb, cs := rgbColor(b), rgbColor(s)
	blended := blendRGB(cb, cs, mode)
	ab, as := float64(b.A)/channelMax, float64(s.A)/channelMax
	alpha := as + ab*(1-as)
	out := [3]uint8{}
	for i := range out {
		out[i] = channel((as*(1-ab)*cs[i] + as*ab*blended[i] + (1-as)*ab*cb[i]) / alpha)
	}
	return color.NRGBA{R: out[0], G: out[1], B: out[2], A: channel(alpha)}
}

func rgbColor(c color.NRGBA) [3]float64 {
	return [3]float64{float64(c.R) / channelMax, float64(c.G) / channelMax, float64(c.B) / channelMax}
}
