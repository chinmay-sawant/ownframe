package draw

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

func TestEncodePNGNil(t *testing.T) {
	t.Parallel()

	if EncodePNG(nil) != nil {
		t.Fatal("nil")
	}
}

func TestEncodePNGRoundTrip(t *testing.T) {
	t.Parallel()

	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	b := EncodePNG(img)
	if b == nil {
		t.Fatal("encode")
	}
	if _, err := png.Decode(bytes.NewReader(b)); err != nil {
		t.Fatal(err)
	}
}
