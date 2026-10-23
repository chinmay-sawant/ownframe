package draw

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

// Fronter is the front-half surface every sink needs: the display list or
// the picture, its size, and its generation. A screen satisfies it, so
// sinks pass their screen straight in.
type Fronter interface {
	Display() *layout.Display
	Image() image.Image
	Size() (int, int)
	Generation() uint64
}

// Front is one display-else-image snapshot. Display is set for a
// display-list page, Image for a bitmap page, and neither before the first
// draw, when Width and Height fall back to the frame size. Fallback reports
// a bitmap page.
type Front struct {
	Display    *layout.Display
	Image      image.Image
	Width      int
	Height     int
	Fallback   bool
	Generation uint64
}

// SelectFront prefers the display list, then the picture bounds, then the
// frame size. It calls only Display, Image, Size, and Generation, so the
// window and the browser share one preference and keep only sink delivery.
func SelectFront(s Fronter) Front {
	gen := s.Generation()
	if display := s.Display(); display != nil {
		return Front{
			Display:    display,
			Width:      display.Width,
			Height:     display.Height,
			Generation: gen,
		}
	}
	if img := s.Image(); img != nil {
		b := img.Bounds()

		return Front{
			Image:      img,
			Width:      b.Dx(),
			Height:     b.Dy(),
			Fallback:   true,
			Generation: gen,
		}
	}
	w, h := s.Size()

	return Front{Width: w, Height: h, Generation: gen}
}
