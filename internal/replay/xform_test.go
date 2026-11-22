package replay

import (
	"context"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"image"
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

// TestTransformedImageOpBounds pins xformPointsBox for DisplayOpImage: a
// transformed image returns tight bounds and culls disjoint rects instead of
// reporting unbounded.
func TestTransformedImageOpBounds(t *testing.T) {
	t.Parallel()

	op := layout.DisplayOp{
		Kind: layout.DisplayOpImage,
		X:    10, Y: 20, W: 40, H: 20,
	}
	// Translate by (+100, +50) in points (ppt = 1 for simple pixel math).
	op.SetXform(layout.Matrix2D{A: 1, D: 1, E: 100, F: 50})

	box, ok := opBounds(&op, 1)
	if !ok {
		t.Fatal("transformed image reported unbounded")
	}
	want := image.Rect(109, 69, 151, 91)
	if box != want {
		t.Fatalf("transformed image opBounds = %v, want %v", box, want)
	}
	if opTouches(&op, 1, image.Rect(0, 0, 50, 50)) {
		t.Fatal("transformed image failed to cull a disjoint rect")
	}
	if !opTouches(&op, 1, image.Rect(115, 75, 130, 85)) {
		t.Fatal("transformed image skipped its mapped rect")
	}
}

// TestTextBoundsExpandForTextStroke pins that -webkit-text-stroke grows both
// direct textBounds and transformed xformPointsBox by half the stroke width.
func TestTextBoundsExpandForTextStroke(t *testing.T) {
	t.Parallel()

	plainDisplay, err := render.DisplayList(context.Background(),
		`<body style="margin:0"><span style="font-size:20px">Hi</span></body>`, 200, 100)
	if err != nil {
		t.Fatal(err)
	}
	strokedDisplay, err := render.DisplayList(context.Background(),
		`<body style="margin:0"><span style="font-size:20px;-webkit-text-stroke:8px red">Hi</span></body>`, 200, 100)
	if err != nil {
		t.Fatal(err)
	}

	var plainOp, strokedOp *layout.DisplayOp
	for i := range plainDisplay.Ops {
		if plainDisplay.Ops[i].Kind == layout.DisplayOpText {
			plainOp = &plainDisplay.Ops[i]
			break
		}
	}
	for i := range strokedDisplay.Ops {
		if strokedDisplay.Ops[i].Kind == layout.DisplayOpText {
			strokedOp = &strokedDisplay.Ops[i]
			break
		}
	}
	if plainOp == nil || strokedOp == nil {
		t.Fatal("missing text ops from rendered displays")
	}
	width, _, ok := strokedOp.TextStroke()
	if !ok || width <= 0 {
		t.Fatalf("expected positive text stroke width, got %v (ok=%v)", width, ok)
	}

	ppt := strokedDisplay.PointsPerPixel
	plainBox, ok := opBounds(plainOp, ppt)
	if !ok {
		t.Fatal("plain text unbounded")
	}
	strokedBox, ok := opBounds(strokedOp, ppt)
	if !ok {
		t.Fatal("stroked text unbounded")
	}
	if strokedBox.Min.X >= plainBox.Min.X || strokedBox.Max.X <= plainBox.Max.X ||
		strokedBox.Min.Y >= plainBox.Min.Y || strokedBox.Max.Y <= plainBox.Max.Y {
		t.Fatalf("stroked textBounds %v did not expand beyond plain %v", strokedBox, plainBox)
	}

	// Transformed text must also expand its bounds by the stroke radius.
	plainXform := *plainOp
	plainXform.SetXform(layout.Matrix2D{A: 1, D: 1, E: 10, F: 0})
	strokedXform := *strokedOp
	strokedXform.SetXform(layout.Matrix2D{A: 1, D: 1, E: 10, F: 0})

	pxBox, ok1 := opBounds(&plainXform, ppt)
	sxBox, ok2 := opBounds(&strokedXform, ppt)
	if !ok1 || !ok2 {
		t.Fatalf("transformed text unbounded: plain=%v stroked=%v", ok1, ok2)
	}
	if sxBox.Min.X >= pxBox.Min.X || sxBox.Max.X <= pxBox.Max.X ||
		sxBox.Min.Y >= pxBox.Min.Y || sxBox.Max.Y <= pxBox.Max.Y {
		t.Fatalf("stroked xform bounds %v did not expand beyond plain %v", sxBox, pxBox)
	}
}
