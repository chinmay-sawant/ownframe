package window

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestDisplayBackground(t *testing.T) {
	t.Parallel()

	display := &layout.Display{
		Width: 320, Height: 400, PointsPerPixel: 0.75,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, R: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 480, H: 600, R: 0.9, G: 0.8, B: 0.7, Alpha: 1},
		},
		Order: []int{0, 1},
	}

	c, ok := displayBackground(display)
	if !ok {
		t.Fatal("no background")
	}

	r, g, b, _ := c.RGBA()
	if r>>8 != 229 || g>>8 != 204 || b>>8 != 178 {
		t.Fatalf("color = %d,%d,%d", r>>8, g>>8, b>>8)
	}
}

func TestDisplayBackgroundMissing(t *testing.T) {
	t.Parallel()

	display := &layout.Display{
		Width: 320, Height: 400, PointsPerPixel: 0.75,
		Ops:   []layout.DisplayOp{{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, Alpha: 1}},
		Order: []int{0},
	}

	if _, ok := displayBackground(display); ok {
		t.Fatal("small fill treated as background")
	}
}

func TestDisplayBackgroundExactCanvas(t *testing.T) {
	t.Parallel()

	// 320 CSS px at 0.75 px per point is a 240-point canvas. A fill that
	// spans it must be found; dividing instead of multiplying the canvas
	// conversion would reject it.
	display := &layout.Display{
		Width: 320, Height: 320, PointsPerPixel: 0.75,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 240, H: 240, R: 1, G: 1, B: 1, Alpha: 1},
		},
		Order: []int{0},
	}

	if _, ok := displayBackground(display); !ok {
		t.Fatal("exact-canvas fill not found")
	}
}
