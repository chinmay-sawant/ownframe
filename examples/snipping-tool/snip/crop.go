package snip

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
)

// Crop returns the pixels of r clamped to src. An empty or
// non-overlapping r returns a copy of src.
func Crop(src image.Image, r Rect) image.Image {
	b := src.Bounds()
	c := Clamp(r, Rect{X: b.Min.X, Y: b.Min.Y, W: b.Dx(), H: b.Dy()})
	if c.W <= 0 || c.H <= 0 {
		c = Rect{X: b.Min.X, Y: b.Min.Y, W: b.Dx(), H: b.Dy()}
	}
	dst := image.NewRGBA(image.Rect(0, 0, c.W, c.H))
	draw.Draw(dst, dst.Bounds(), src, image.Pt(c.X, c.Y), draw.Src)
	return dst
}

// CropPNG decodes a PNG, crops it with Crop, and re-encodes it.
func CropPNG(pngBytes []byte, r Rect) ([]byte, error) {
	src, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, Crop(src, r)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
