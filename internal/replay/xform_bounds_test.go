package replay

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// TestPaintBoundsMapsTransform pins the transformed answer: the local box is
// mapped through the baked matrix, so a quarter turn bounds tall instead of
// wide.
func TestPaintBoundsMapsTransform(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayList(context.Background(),
		`<div style="width:120px;height:40px;background:#eee;transform:rotate(90deg)"></div>`,
		200, 200)
	if err != nil {
		t.Fatal(err)
	}

	for i := range display.Ops {
		op := &display.Ops[i]
		if op.Kind != layout.DisplayOpFillRect || !op.XformSet {
			continue
		}

		box, ok := PaintBounds(op, display.PointsPerPixel)
		if !ok || box.Empty() {
			t.Fatalf("transformed bounds = %v, %v", box, ok)
		}

		if box.Dy() <= box.Dx() {
			t.Fatalf("rotated bounds %v are not tall", box)
		}

		return
	}

	t.Fatal("no transformed fill")
}

// TestPaintBoundsRejectsRotatedGlyphs pins the run the painter still cannot
// place: a rotated glyph carries no matrix the bounds can map.
func TestPaintBoundsRejectsRotatedGlyphs(t *testing.T) {
	t.Parallel()

	op := &layout.DisplayOp{Kind: layout.DisplayOpText, X: 1, Y: 2, W: 3, H: 4, RotateDeg: 20}
	if box, ok := PaintBounds(op, 1); ok || !box.Empty() {
		t.Fatalf("rotated bounds = %v, %v", box, ok)
	}
}
