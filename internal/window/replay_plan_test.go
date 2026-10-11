package window

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestPlanRepaintPicksTheWork(t *testing.T) {
	t.Parallel()

	state := partialState{buf: partialBuf(100, 100), gen: 1, width: 100, height: 100}
	badSize := state
	badSize.width = 80
	full := image.Rect(0, 0, 100, 100)

	cases := []struct {
		name     string
		st       partialState
		gen      uint64
		rect     image.Rectangle
		ok       bool
		want     repaintMode
		wantRect image.Rectangle
	}{
		{"first frame", partialState{}, 1, image.Rect(10, 10, 20, 20), true, repaintBuffer, full},
		{"dirty rect", state, 2, image.Rect(10, 10, 20, 20), true, repaintRect, image.Rect(10, 10, 20, 20)},
		{"full rect", state, 2, full, true, repaintBuffer, full},
		{"false rect", state, 2, image.Rectangle{}, false, repaintBuffer, full},
		{"same gen", state, 1, image.Rectangle{}, false, repaintBlit, full},
		{"size change", badSize, 2, image.Rect(0, 0, 10, 10), true, repaintBuffer, full},
		{"outside clamped", state, 2, image.Rect(-20, -20, 30, 30), true, repaintRect, image.Rect(0, 0, 30, 30)},
		{"outside away", state, 2, image.Rect(-30, -30, -20, -20), true, repaintBuffer, full},
	}
	for _, c := range cases {
		got := planRepaint(100, 100, c.st, c.gen, c.rect, c.ok)
		if got.mode != c.want || got.rect != c.wantRect {
			t.Fatalf("%s: plan = %v %v, want %v %v", c.name, got.mode, got.rect, c.want, c.wantRect)
		}
	}
}

// partialDisplay is a small replayable page with one box.
func partialDisplay(w, h int) *layout.Display {
	return &layout.Display{
		Width: w, Height: h, PointsPerPixel: 1,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: float64(w), H: float64(h), R: 1, G: 1, B: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, R: 1, Alpha: 1},
		},
		Order: []int{0, 1},
	}
}

func partialBuf(w, h int) *ebiten.Image {
	return ebiten.NewImage(w, h)
}
