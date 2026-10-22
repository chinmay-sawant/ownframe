package snip

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// DemoImage builds a deterministic gradient plus grid PNG fallback.
// It never touches the screen, so the example works headless.
func DemoImage(w, h int) []byte {
	if w < demoMinSize {
		w = demoWidth
	}
	if h < demoMinSize {
		h = demoHeight
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r := uint8(x * 255 / (w - 1))                   //nolint:gosec // reason: x bounded by w, result 0..255.
			g := uint8(y * 255 / (h - 1))                   //nolint:gosec // reason: y bounded by h, result 0..255.
			b := uint8(blueBase + 127*(x+y)/(w+h-demoEdge)) //nolint:gosec // reason: x,y bounded by w,h, result 0..255.
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	line := color.RGBA{255, 255, 255, 255}
	for x := 0; x < w; x += 32 {
		for y := 0; y < h; y++ {
			img.Set(x, y, line)
		}
	}
	for y := 0; y < h; y += 32 {
		for x := 0; x < w; x++ {
			img.Set(x, y, line)
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
