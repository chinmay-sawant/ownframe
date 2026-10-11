package window

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// TestDevOpRectTextLineBox checks a text op's outline starts above its
// baseline, one line box tall.
func TestDevOpRectTextLineBox(t *testing.T) {
	t.Parallel()

	op := layout.DisplayOp{Kind: layout.DisplayOpText, X: 10, Y: 30, W: 80, H: 16, InkDescent: 4}

	got := devOpRect(&op, 1)
	if got.Y != 18 || got.H != 16 {
		t.Fatalf("text rect = %+v, want the line box at y 18", got)
	}
}

// TestDevOpRectUsesCanvasPixels pins the conversion direction: op coordinates
// are points, and PointsPerPixel divides them to CSS pixels. The old inspector
// multiplied, so the outline sat at 56% of the paint.
func TestDevOpRectUsesCanvasPixels(t *testing.T) {
	t.Parallel()

	op := layout.DisplayOp{Kind: layout.DisplayOpFillRect, W: 90, H: 30}

	if got := devOpRect(&op, 0.75); got != (devRect{W: 120, H: 40}) {
		t.Fatalf("fill rect = %+v, want 120x40", got)
	}
}

// TestDevOpRectCoversStrokeInk checks the outline uses the painted box rather
// than the nominal border box, so half a wide border is not left outside it.
func TestDevOpRectCoversStrokeInk(t *testing.T) {
	t.Parallel()

	op := layout.DisplayOp{Kind: layout.DisplayOpStrokeRect, W: 80, H: 40, Width: 4}

	if got := devOpRect(&op, 1); got != (devRect{X: -2, Y: -2, W: 84, H: 44}) {
		t.Fatalf("stroke rect = %+v, want the ink box", got)
	}
}

// TestDevOpRectFallsBackWhenUnbounded checks a transformed run still gets its
// nominal box, with a text baseline applied, for the list and count rows.
func TestDevOpRectFallsBackWhenUnbounded(t *testing.T) {
	t.Parallel()

	op := layout.DisplayOp{Kind: layout.DisplayOpText, X: 10, Y: 30, W: 80, H: 16, InkDescent: 4, XformSet: true}

	if got := devOpRect(&op, 1); got != (devRect{X: 10, Y: 18, W: 80, H: 16}) {
		t.Fatalf("unbounded text rect = %+v, want the nominal line box", got)
	}
}
