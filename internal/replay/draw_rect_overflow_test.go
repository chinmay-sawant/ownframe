package replay

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// overflowDisplay is a 100x100 canvas with one red fill below it, inside a
// 100x170 content. Points use 72/96 per pixel, so 75 points are 100 pixels.
func overflowDisplay() *layout.Display {
	const ppt = 72.0 / 96.0

	return &layout.Display{
		Width: 100, Height: 100, PointsPerPixel: ppt,
		Boxes: []layout.Box{{X: 0, Y: 150, W: 100, H: 20}},
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 75, H: 75, R: 1, G: 1, B: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 112.5, W: 75, H: 15, R: 1, Alpha: 1},
		},
		Order: []int{0, 1},
	}
}

// The buffer covers the content past the canvas, so a dirty rect below the
// canvas must survive clipping and still select its ops for replay.
func TestDrawRectKeepsOverflowRect(t *testing.T) {
	t.Parallel()

	display := overflowDisplay()
	rect := clipRect(display, image.Rect(0, 140, 100, 180))
	if rect != image.Rect(0, 140, 100, 170) {
		t.Fatalf("clipped = %v, want the content overlap", rect)
	}

	deep := &display.Ops[1]
	if !opTouches(deep, display.PointsPerPixel, rect) {
		t.Fatal("the fill below the canvas was skipped")
	}

	shallow := &display.Ops[0]
	if opTouches(shallow, display.PointsPerPixel, rect) {
		t.Fatal("the canvas fill leaks into the overflow rect")
	}
}

func TestClipRectDropsPastContent(t *testing.T) {
	t.Parallel()

	display := overflowDisplay()
	if !clipRect(display, image.Rect(0, 180, 100, 200)).Empty() {
		t.Fatal("a rect past the content should clamp away")
	}
}
