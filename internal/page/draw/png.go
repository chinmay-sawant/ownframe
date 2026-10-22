package draw

import (
	"bytes"
	"image"
	"image/png"
)

// EncodePNG writes img as PNG bytes, or nil on failure.
func EncodePNG(img image.Image) []byte {
	if img == nil {
		return nil
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil
	}

	return buf.Bytes()
}
