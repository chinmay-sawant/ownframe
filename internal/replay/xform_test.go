package replay

import (
	"context"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"testing"
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

// TestXformGeomMapsOpOrigin pins the matrix the transformed painter uses: a
// local origin maps to the op's transformed layout origin in canvas pixels.
func TestXformGeomMapsOpOrigin(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayList(context.Background(),
		`<div style="width:60px;height:20px;background:#eee;transform:rotate(30deg)"></div>`,
		200, 120)
	if err != nil {
		t.Fatal(err)
	}

	for i := range display.Ops {
		op := &display.Ops[i]
		if op.Kind != layout.DisplayOpFillRect || !op.XformSet {
			continue
		}

		m := op.Transform()
		wantX, wantY := m.Apply(op.X, op.Y)
		geom := xformGeom(op, 1, 2)
		gotX, gotY := geom.Apply(0, 0)

		if !near(gotX, wantX*pxPerPt+1) || !near(gotY, wantY*pxPerPt+2) {
			t.Fatalf("origin = (%v,%v), want (%v,%v)", gotX, gotY, wantX*pxPerPt+1, wantY*pxPerPt+2)
		}

		return
	}

	t.Fatal("no transformed fill")
}

func near(a, b float64) bool {
	d := a - b

	return d < 1e-9 && d > -1e-9
}
