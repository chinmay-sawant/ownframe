package window

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// An edit below the canvas must partially repaint the content-sized buffer:
// the buffer covers the overflow, and the dirty rect must reach DrawRect.
func TestOverflowEditRepaintsOneRect(t *testing.T) {
	t.Parallel()

	const ppt = 72.0 / 96.0
	boxes := []layout.Box{{X: 0, Y: 0, W: 100, H: 400}}
	display := &layout.Display{
		Width: 100, Height: 100, PointsPerPixel: ppt,
		Boxes: boxes,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 75, H: 75, R: 1, G: 1, B: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 112.5, W: 75, H: 15, R: 1, Alpha: 1},
		},
		Order: []int{0, 1},
	}

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = display
	screen.boxes = boxes
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(100, 100)

	s.drawReplayPartial(dst)

	if s.partial.buf == nil || s.partial.buf.Bounds() != image.Rect(0, 0, 100, 400) {
		t.Fatalf("buffer covers %v, want the 100x400 content", s.partial.buf.Bounds())
	}

	overflow := image.Rect(0, 140, 100, 180)
	screen.click(overflow)
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintRect || s.partial.last != overflow {
		t.Fatalf("mode = %v, rect = %v, want a repaint of %v", s.partial.mode, s.partial.last, overflow)
	}
}
