package devtools

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// Test image geometry and palette steps for shotPNG.
const (
	shotSide  = 4
	shotRed   = 0x30
	shotGreen = 0x40
	shotBlue  = 0xdb
	shotAlpha = 0xff
)

// shotPNG builds the small image the template shows through SetImage. The
// example ships no binary asset, so the bytes come from an encoder.
func shotPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, shotSide, shotSide))

	for y := range shotSide {
		for x := range shotSide {
			img.Set(x, y, color.RGBA{
				R: uint8(shotRed * x),   //nolint:gosec // reason: x stays 0..3, so shotRed*x tops out at 0x90 and fits a byte.
				G: uint8(shotGreen * y), //nolint:gosec // reason: y stays 0..3, so shotGreen*y tops out at 0xc0 and fits a byte.
				B: shotBlue,
				A: shotAlpha,
			})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil
	}

	return buf.Bytes()
}
