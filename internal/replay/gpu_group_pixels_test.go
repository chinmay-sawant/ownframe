package replay

import (
	"fmt"
	"image"
	"image/color"

	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
)

// compareGroupPixels draws the display with the replay and the bitmap
// painters and requires the sampled pixels to agree within rounding.
func compareGroupPixels(d *render.Display, name string, points [][2]int) error {
	want, err := bitmap.Paint(d)
	if err != nil {
		return err
	}

	img := ebiten.NewImage(200, 140)
	defer img.Dispose()

	DrawVisible(img, d, 0, 0)

	pixels := make([]byte, 200*140*4)
	img.ReadPixels(pixels)

	for _, p := range points {
		if err := closePixel(pixels, want, 200, p[0], p[1]); err != nil {
			return fmt.Errorf("%s pixel (%d,%d): %w", name, p[0], p[1], err)
		}
	}

	return nil
}

// closePixel compares one rendered pixel with the bitmap painter's, within
// three channel steps.
func closePixel(pixels []byte, want image.Image, stride, x, y int) error {
	at := (y*stride + x) * 4
	w := color.NRGBAModel.Convert(want.At(x, y)).(color.NRGBA)
	wantRGBA := []uint8{w.R, w.G, w.B, w.A}

	for i, v := range wantRGBA {
		if delta := int(pixels[at+i]) - int(v); delta < -3 || delta > 3 {
			return fmt.Errorf("got %v, want %v", pixels[at:at+4], wantRGBA)
		}
	}

	return nil
}
