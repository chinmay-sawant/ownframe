package insights

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	xdraw "golang.org/x/image/draw"
)

const (
	badgeSize  = 70
	badgeScale = 4
)

// ringText repeats around the badge, one rune per step.
const ringText = "SHARE · SHARE · SHARE · "

// ShareBadgePNG draws the circular SHARE badge: a cream disc with the word
// repeated around a ring. The share arrow is an SVG painted on top.
func ShareBadgePNG() []byte {
	side := badgeSize * badgeScale
	img := image.NewRGBA(image.Rect(0, 0, side, side))
	cream := color.RGBA{R: creamRed, G: creamGreen, B: creamBlue, A: opaqueAlpha}

	drawDisc(img, float64(side)/halfDivisor, float64(side)/halfDivisor, float64(side)/halfDivisor, cream)
	drawRing(img)

	small := image.NewRGBA(image.Rect(0, 0, badgeSize, badgeSize))
	xdraw.CatmullRom.Scale(small, small.Bounds(), img, img.Bounds(), draw.Src, nil)

	var buf bytes.Buffer
	_ = png.Encode(&buf, small)

	return buf.Bytes()
}

// drawDisc fills one circle.
func drawDisc(img *image.RGBA, cx, cy, r float64, c color.RGBA) {
	for y := range img.Bounds().Dy() {
		for x := range img.Bounds().Dx() {
			dx, dy := float64(x)-cx, float64(y)-cy
			if dx*dx+dy*dy <= r*r {
				img.SetRGBA(x, y, c)
			}
		}
	}
}
