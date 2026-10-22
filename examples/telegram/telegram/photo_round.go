package telegram

import (
	"image"
	"image/color"

	"golang.org/x/image/draw"
)

// roundRadius is the corner radius baked into a photo bubble, because the
// page paints an image as a square.
const roundRadius = 14

// roundPhoto returns src with transparent pixels outside a corner radius.
func roundPhoto(src image.Image) *image.RGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)

	r := roundRadius
	if r*2 > w {
		r = w / halfDivisor
	}

	if r*2 > h {
		r = h / halfDivisor
	}

	for y := 0; y < r; y++ {
		for x := 0; x < r; x++ {
			dx, dy := r-1-x, r-1-y
			if dx*dx+dy*dy <= r*r {
				continue
			}

			dst.SetRGBA(x, y, color.RGBA{})
			dst.SetRGBA(w-1-x, y, color.RGBA{})
			dst.SetRGBA(x, h-1-y, color.RGBA{})
			dst.SetRGBA(w-1-x, h-1-y, color.RGBA{})
		}
	}

	return dst
}
