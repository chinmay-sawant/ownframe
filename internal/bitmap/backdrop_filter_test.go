package bitmap

import (
	"context"
	"image"
	"image/color"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/backdrop"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestApplyBackdropFilterKeepsOffscreenRoundedCorner(t *testing.T) {
	t.Parallel()

	dst := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	dst.SetNRGBA(0, 10, color.NRGBA{R: 255, A: 255})
	applyBackdropFilter(dst, image.Rect(-10, 10, 40, 50), 12, backdrop.Filter{Saturation: 0})

	got := dst.NRGBAAt(0, 10)
	if got.R != got.G || got.G != got.B {
		t.Fatalf("partially clipped rounded corner = %v, want filtered grayscale from original shape", got)
	}
}

func TestApplyBackdropFilterBlursPremultipliedPixels(t *testing.T) {
	t.Parallel()

	dst := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	dst.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 128})
	dst.SetNRGBA(1, 0, color.NRGBA{B: 255, A: 0})
	applyBackdropFilter(dst, dst.Bounds(), 0, backdrop.Filter{BlurPx: 1, Saturation: 1})

	got := dst.NRGBAAt(1, 0)
	if got.R < 250 || got.G > 1 || got.B > 1 || got.A < 63 || got.A > 65 {
		t.Fatalf("blurred transparent edge = %v, want red with alpha 64", got)
	}
}

func TestApplyBackdropFilterSaturatesTranslucentPixelsWithoutChangingAlpha(t *testing.T) {
	t.Parallel()

	dst := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	dst.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 128})
	dst.SetNRGBA(1, 0, color.NRGBA{R: 255, A: 128})
	applyBackdropFilter(dst, dst.Bounds(), 0, backdrop.Filter{Saturation: 0})

	got := dst.NRGBAAt(1, 0)
	if got.A != 128 || got.R < 52 || got.R > 56 || got.G < 52 || got.G > 56 || got.B < 52 || got.B > 56 {
		t.Fatalf("saturated translucent pixel = %v, want grayscale near 54 with alpha 128", got)
	}
}

func TestApplyBackdropFilterRefractsNearRoundedGlassEdge(t *testing.T) {
	t.Parallel()

	dst := image.NewNRGBA(image.Rect(0, 0, 100, 60))
	for y := 0; y < 60; y++ {
		for x := 0; x < 100; x++ {
			if x < 42 {
				dst.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
			} else {
				dst.SetNRGBA(x, y, color.NRGBA{B: 255, A: 255})
			}
		}
	}

	applyBackdropFilter(dst, image.Rect(40, 0, 80, 60), 0, backdrop.Filter{Saturation: 1, RefractionPx: 5})
	got := dst.NRGBAAt(40, 30)
	if got.B < 240 || got.R > 20 {
		t.Fatalf("refracted boundary pixel = %v, want the backdrop bent inward to blue", got)
	}
	center := dst.NRGBAAt(60, 30)
	if center.B < 240 || center.R > 20 {
		t.Fatalf("glass center pixel = %v, want unchanged blue backdrop", center)
	}
}

func TestPaintBackdropBlurMixesAdjacentColors(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayList(context.Background(), `<body style="margin:0;background:white">`+
		`<div style="position:absolute;left:0;top:0;width:50px;height:100px;background:#ff0000"></div>`+
		`<div style="position:absolute;left:50px;top:0;width:50px;height:100px;background:#0000ff"></div>`+
		`<div style="position:absolute;left:40px;top:0;width:20px;height:100px;`+
		`backdrop-filter:blur(10px);background:transparent"></div></body>`, 100, 100)
	if err != nil {
		t.Fatal(err)
	}

	painted, err := Paint(display)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := color.NRGBAModel.Convert(painted.At(49, 50)).(color.NRGBA)
	if !ok {
		t.Fatal("pixel model changed")
	}
	if got.B < 35 {
		t.Fatalf("filtered boundary pixel = %v, want blue from the adjacent backdrop", got)
	}
}

func TestPaintEmptyBackdropFilterSaturatesItsBackdrop(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayList(context.Background(), `<body style="margin:0">`+
		`<div style="position:absolute;left:0;top:0;width:100px;height:100px;background:#804020"></div>`+
		`<div style="position:absolute;left:20px;top:20px;width:60px;height:60px;`+
		`backdrop-filter:saturate(0);background:transparent"></div></body>`, 100, 100)
	if err != nil {
		t.Fatal(err)
	}

	painted, err := Paint(display)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := color.NRGBAModel.Convert(painted.At(50, 50)).(color.NRGBA)
	if !ok {
		t.Fatal("pixel model changed")
	}
	if got.A < 250 || max(got.R, got.G, got.B)-min(got.R, got.G, got.B) > 2 || got.R < 70 || got.R > 80 {
		t.Fatalf("saturated empty-filter pixel = %v, want opaque grayscale near luma 75", got)
	}
	outside, ok := color.NRGBAModel.Convert(painted.At(10, 10)).(color.NRGBA)
	if !ok {
		t.Fatal("pixel model changed")
	}
	if outside != (color.NRGBA{R: 128, G: 64, B: 32, A: 255}) {
		t.Fatalf("pixel outside empty filter = %v, want original backdrop color", outside)
	}
}
