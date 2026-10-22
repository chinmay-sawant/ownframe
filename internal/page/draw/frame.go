package draw

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

// Frame holds the last committed picture. Page holds one, like dirty.Tracker.
type Frame struct {
	Img             image.Image
	Display         *layout.Display
	FallbackDisplay *layout.Display
	PNG             []byte
}

// DropPNG clears the encoded cache so the next PNG encodes again.
func (f *Frame) DropPNG() { f.PNG = nil }

// StorePNG keeps encoded bytes until the next commit.
func (f *Frame) StorePNG(b []byte) { f.PNG = b }

// CopyPNG returns a copy of the cached bytes, or nil.
func (f *Frame) CopyPNG() []byte {
	if f.PNG == nil {
		return nil
	}

	return append([]byte(nil), f.PNG...)
}
