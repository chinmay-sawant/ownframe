package frame

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe"
)

// testDisplay builds a display whose ops live in points: the box below is
// 100x20 CSS pixels at 0.75 px per point, so its op rect is 7.5..82.5 x
// 7.5..22.5.
func testDisplay() *ownframe.Display {
	return &ownframe.Display{
		Width: 400, Height: 200, PointsPerPixel: 0.75,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 1, Y: 8, W: 4, H: 4, R: 0.2, G: 0.2, B: 0.2, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 12, Y: 10, W: 8, H: 4, R: 0.42, G: 0.15, B: 0.85, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 30, Y: 10, W: 8, H: 4, R: 0.42, G: 0.15, B: 0.85, Alpha: 1},
			{Kind: layout.DisplayOpText, X: 12, Y: 16, Text: "1:08", Alpha: 1},
		},
		Order: []int{0, 1, 2, 3},
	}
}

func TestFillFindsColorInsideBox(t *testing.T) {
	t.Parallel()

	display := testDisplay()
	box := ownframe.Box{X: 10, Y: 10, W: 100, H: 20}

	if got := Fill(display, box, [3]float64{0.42, 0.15, 0.85}); got == nil {
		t.Fatal("no purple fill found")
	} else if got.W != 8 {
		t.Fatalf("fill W = %v", got.W)
	}

	if got := Fill(display, box, [3]float64{0.9, 0.9, 0.9}); got != nil {
		t.Fatalf("unexpected fill %+v", got)
	}
}

func TestFillsAreLeftToRight(t *testing.T) {
	t.Parallel()

	fills := Fills(testDisplay(), ownframe.Box{X: 10, Y: 10, W: 100, H: 20}, [3]float64{0.42, 0.15, 0.85})
	if len(fills) != 2 {
		t.Fatalf("fills = %d", len(fills))
	}

	if fills[0].X >= fills[1].X {
		t.Fatal("fills are not left to right")
	}
}

func TestFillsOutsideTheBoxAreSkipped(t *testing.T) {
	t.Parallel()

	fills := Fills(testDisplay(), ownframe.Box{X: 200, Y: 100, W: 50, H: 20}, [3]float64{0.42, 0.15, 0.85})
	if len(fills) != 0 {
		t.Fatalf("fills = %d", len(fills))
	}
}
