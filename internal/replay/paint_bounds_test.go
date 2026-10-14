package replay

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// TestPaintBoundsDropsTheReplaySlack pins the inspector's entry point: the
// same box the dirty-rect filter uses, without its one-pixel margin.
func TestPaintBoundsDropsTheReplaySlack(t *testing.T) {
	t.Parallel()

	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 80, H: 40}

	padded, ok := opBounds(op, 1)
	if !ok || padded != image.Rect(9, 9, 91, 51) {
		t.Fatalf("padded = %v, %v", padded, ok)
	}

	paint, ok := PaintBounds(op, 1)
	if !ok || paint != image.Rect(10, 10, 90, 50) {
		t.Fatalf("paint = %v, %v", paint, ok)
	}
}

// TestPaintBoundsCoversStrokeAndLine checks that the painted box includes the
// ink outside the nominal geometry: half the stroke on each side of a border,
// and the stroke with its square cap around a rule.
func TestPaintBoundsCoversStrokeAndLine(t *testing.T) {
	t.Parallel()

	border := &layout.DisplayOp{Kind: layout.DisplayOpStrokeRect, X: 0, Y: 0, W: 80, H: 40, Width: 4}
	if got, ok := PaintBounds(border, 1); !ok || got != image.Rect(-2, -2, 82, 42) {
		t.Fatalf("stroke bounds = %v, %v", got, ok)
	}

	rule := &layout.DisplayOp{Kind: layout.DisplayOpLine, X: 0, Y: 10, W: 100, H: 0, Width: 4}
	if got, ok := PaintBounds(rule, 1); !ok || got != image.Rect(-2, 8, 102, 12) {
		t.Fatalf("line bounds = %v, %v", got, ok)
	}
}
