package draw

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

// CommitList installs a replayable list and drops the bitmap.
func (f *Frame) CommitList(d *layout.Display) {
	f.Img = nil
	f.Display = d
	f.FallbackDisplay = nil
	f.PNG = nil
}

// CommitBitmap installs a fallback picture and drops the list.
func (f *Frame) CommitBitmap(d *layout.Display, img image.Image) {
	f.Img = img
	f.Display = nil
	f.FallbackDisplay = d
	f.PNG = nil
}

// SwapList replaces the list in place, drops the PNG cache, and returns
// the previous list so a patch can still diff against it.
func (f *Frame) SwapList(d *layout.Display) *layout.Display {
	prev := f.Display
	f.Display = d
	f.PNG = nil

	return prev
}
