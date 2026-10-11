package replay

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestClipRectClampsNotDrops(t *testing.T) {
	t.Parallel()

	display := &layout.Display{Width: 100, Height: 80, PointsPerPixel: 1}

	got := clipRect(display, image.Rect(-20, -20, 30, 30))
	if got != image.Rect(0, 0, 30, 30) {
		t.Fatalf("clamped = %v", got)
	}

	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 10, H: 10}
	if !opTouches(op, 1, got) {
		t.Fatal("op at the canvas corner was dropped")
	}

	if !clipRect(display, image.Rect(-30, -30, -20, -20)).Empty() {
		t.Fatal("a rect fully outside should clamp away")
	}
}

func TestOpTouchesSkipsOutsideOps(t *testing.T) {
	t.Parallel()

	inside := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20}
	outside := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 80, Y: 80, W: 10, H: 10}
	rect := image.Rect(0, 0, 40, 40)

	if !opTouches(inside, 1, rect) {
		t.Fatal("inside op skipped")
	}

	if opTouches(outside, 1, rect) {
		t.Fatal("outside op drawn")
	}
}

func TestOpBoundsInertIsEmpty(t *testing.T) {
	t.Parallel()

	op := &layout.DisplayOp{Kind: layout.DisplayOpLinkURI, X: 1, Y: 1, W: 9, H: 9}

	box, ok := opBounds(op, 1)
	if !ok || !box.Empty() {
		t.Fatalf("link op bounds = %v, %v", box, ok)
	}
}
