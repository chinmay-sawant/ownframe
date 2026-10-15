package masking

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestImageRejectsNilAndEmpty(t *testing.T) {
	b := image.Rect(0, 0, 4, 4)
	if _, err := Image(b, nil); err == nil {
		t.Fatal("nil desc must err")
	}
	if _, err := Image(b, &layout.MaskDescriptor{}); err == nil {
		t.Fatal("empty layers must err")
	}
}
