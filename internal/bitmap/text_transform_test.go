package bitmap

import (
	"image"
	"image/color"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"golang.org/x/image/math/f64"
)

func TestQuarterTurnTextTransformUsesExactIntegerCoefficients(t *testing.T) {
	t.Parallel()

	// Verify that quarter-turn rotations in paintTextOp produce exact {-1,0,1}
	// sin/cos without math.Cos(math.Pi/2) floating-point drift, and that
	// paintQuarterTurnText maps a single lit pixel to its exact rotated pixel.
	src := image.NewNRGBA(image.Rect(0, 0, 40, 40))
	src.SetNRGBA(20, 10, color.NRGBA{R: 255, G: 0, B: 0, A: 255})

	// 90-degree clockwise rotation around (20, 20):
	// (x', y') = (-(y - 20) + 20, (x - 20) + 20) = (-y + 40, x).
	// Pixel center (20.5, 10.5) maps to (29.5, 20.5) -> pixel (29, 20).
	dst := image.NewNRGBA(image.Rect(0, 0, 40, 40))
	matrix := f64.Aff3{0, -1, 40, 1, 0, 0}
	paintQuarterTurnText(dst, src, tightAlphaBounds(src), matrix)

	got := dst.NRGBAAt(29, 20)
	if got.R != 255 || got.A != 255 {
		t.Fatalf("rotated pixel at (29,20) = %v, want opaque red", got)
	}
	if dst.NRGBAAt(28, 20).A != 0 || dst.NRGBAAt(30, 20).A != 0 {
		t.Fatal("quarter-turn leaked into adjacent pixels")
	}
}

func TestGroupOpBoundsTightensNonMaskedBlendGroup(t *testing.T) {
	t.Parallel()

	g := &layout.DisplayGroup{ID: 1, Mode: "multiply"}
	op := &layout.DisplayOp{
		Kind: layout.DisplayOpFillRect,
		X:    15, Y: 15, W: 30, H: 30,
	}
	items := map[*layout.DisplayGroup][]paintItem{
		g: {{op: op}},
	}
	bounds, ok := groupOpBounds(g, items, nil)
	if !ok {
		t.Fatal("expected bounded rect for fill-rect blend group")
	}
	// 15..45 pt * (96/72) = 20..60 px, padded by 2px -> 18..62 px.
	want := image.Rect(18, 18, 62, 62)
	if bounds != want {
		t.Fatalf("groupOpBounds = %v, want %v", bounds, want)
	}
	canvas := image.Rect(0, 0, 800, 600)
	if !bounds.In(canvas) || bounds.Dx()*bounds.Dy() >= canvas.Dx()*canvas.Dy()/10 {
		t.Fatalf("groupOpBounds %v did not tighten relative to canvas %v", bounds, canvas)
	}
}
