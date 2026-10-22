package insights

import (
	"image"
	"image/color"
	"math"

	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
)

// drawRing places each ring rune upright against the rim.
func drawRing(img *image.RGBA) {
	ft, err := opentype.Parse(gobold.TTF)
	if err != nil {
		return
	}

	face, err := opentype.NewFace(ft, &opentype.FaceOptions{
		Size: badgeFontSize * badgeScale,
		DPI:  badgeDPI,
	})
	if err != nil {
		return
	}
	defer func() { _ = face.Close() }()

	ink := color.RGBA{R: inkRed, G: inkGreen, B: inkBlue, A: opaqueAlpha}
	side := float64(img.Bounds().Dx())
	radius := side * ringFactor
	step := piMultiplier * math.Pi / float64(len(ringText))

	for i, r := range ringText {
		if r == ' ' {
			continue
		}

		angle := float64(i) * step
		x := side/halfDivisor + radius*math.Sin(angle)
		y := side/halfDivisor - radius*math.Cos(angle)

		rotateInto(img, drawGlyph(face, string(r), ink), x, y, angle)
	}
}
