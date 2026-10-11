package window

import (
	"image/color"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestDisplayBackgroundPaintOrder(t *testing.T) {
	t.Parallel()

	display := &layout.Display{
		Width: 320, Height: 320, PointsPerPixel: 0.75,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 240, H: 240, B: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 240, H: 240, R: 1, Alpha: 1},
		},
		Order: []int{1, 0},
	}

	c, ok := displayBackground(display)
	if !ok {
		t.Fatal("no background")
	}

	r, _, _, _ := c.RGBA()
	if r>>8 != 255 {
		t.Fatalf("picked the wrong op: red = %d", r>>8)
	}
}

func TestPageBackgroundFallback(t *testing.T) {
	t.Parallel()

	s := &shell{app: &fakeScreen{width: 100, height: 100}}
	if c := s.pageBackground(); c != color.White {
		t.Fatalf("fallback = %v", c)
	}
}
