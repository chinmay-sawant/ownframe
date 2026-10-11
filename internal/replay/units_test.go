package replay

import (
	"context"
	"math"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// TestCanvasUnitsUsePointsPerPixel locks the conversion the op filter relies
// on: a canvas of display.Width CSS pixels is display.Width * PointsPerPixel
// wide in op points.
func TestCanvasUnitsUsePointsPerPixel(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayListState(context.Background(),
		`<html><body style="margin:0;background:#ffffff;height:100px"></body></html>`,
		320, 200, render.State{})
	if err != nil {
		t.Fatal(err)
	}

	wantPts := float64(display.Width) * display.PointsPerPixel

	var found *layout.DisplayOp
	for i := range display.Ops {
		op := &display.Ops[i]
		if op.Kind == layout.DisplayOpFillRect && op.X == 0 && op.Y == 0 {
			found = op
		}
	}

	if found == nil {
		t.Fatal("no canvas fill")
	}

	if math.Abs(found.W-wantPts) > 0.5 {
		t.Fatalf("fill W = %v points, want %v", found.W, wantPts)
	}

	// In canvas pixels the fill must span the whole canvas.
	box, ok := opBounds(found, display.PointsPerPixel)
	if !ok {
		t.Fatal("fill unbounded")
	}

	if box.Dx() < display.Width {
		t.Fatalf("bounds %v are not in canvas pixels, want width %d", box, display.Width)
	}
}
