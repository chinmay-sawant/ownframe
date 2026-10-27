package replay

import (
	"context"
	"image"
	"image/color"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// setupTextBench provides a rendered DisplayOp and its GoTextFace for text benchmarking.
func setupTextBench(b *testing.B, sampleText string, letterSpacing float64) (*Renderer, *layout.DisplayOp, *text.GoTextFace, *ebiten.Image) {
	b.Helper()
	d, err := render.DisplayList(context.Background(),
		`<html><body style="margin:0;font-size:16px"><p>Sample</p></body></html>`,
		400, 100)
	if err != nil {
		b.Fatal(err)
	}

	op := firstTextOp(d)
	if op == nil {
		b.Fatal("no text op found")
	}

	op.Text = sampleText
	op.LetterSpacing = letterSpacing

	r := &Renderer{}
	face := r.textFace(op)
	if face == nil {
		b.Fatal("no face resolved")
	}

	dst := newParityImage(600, 200)
	b.Cleanup(func() {
		dst.Deallocate()
		r.Close()
	})

	return r, op, face, dst
}

// -----------------------------------------------------------------------------
// 1. Vector & Raster Cache Blits
// -----------------------------------------------------------------------------

func BenchmarkBlitRaster(b *testing.B) {
	dst := newParityImage(400, 400)
	src := newParityImage(64, 64)
	tint := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	defer dst.Deallocate()
	defer src.Deallocate()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blitRaster(dst, src, 10, 20, tint)
	}
}

func BenchmarkCachedRoundedFill_Warm(b *testing.B) {
	r := &Renderer{}
	defer r.Close()
	dst := newParityImage(400, 400)
	defer dst.Deallocate()

	op := &layout.DisplayOp{
		Kind:  layout.DisplayOpFillRect,
		X:     10,
		Y:     10,
		W:     120,
		H:     60,
		R:     100,
		G:     150,
		B:     200,
		Alpha: 1,
	}
	radii := [4]float64{8, 8, 8, 8}

	// Warm cache
	if !r.cachedRoundedFill(dst, op, radii, 0, 0) {
		b.Fatal("failed to cache rounded fill")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.cachedRoundedFill(dst, op, radii, 0, 0)
	}
}

func BenchmarkCachedRoundedFill_Cold(b *testing.B) {
	r := &Renderer{}
	defer r.Close()
	dst := newParityImage(400, 400)
	defer dst.Deallocate()

	op := &layout.DisplayOp{
		Kind:  layout.DisplayOpFillRect,
		X:     10,
		Y:     10,
		W:     120,
		H:     60,
		R:     100,
		G:     150,
		B:     200,
		Alpha: 1,
	}
	radii := [4]float64{8, 8, 8, 8}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Flush cache each iteration to simulate cold bake
		r.rasters.mu.Lock()
		for _, img := range r.rasters.entries {
			img.Deallocate()
		}
		clear(r.rasters.entries)
		r.rasters.order = nil
		r.rasters.pixels = 0
		r.rasters.mu.Unlock()

		r.cachedRoundedFill(dst, op, radii, 0, 0)
	}
}

func BenchmarkCachedStroke_Warm(b *testing.B) {
	r := &Renderer{}
	defer r.Close()
	dst := newParityImage(400, 400)
	defer dst.Deallocate()

	op := &layout.DisplayOp{
		Kind:       layout.DisplayOpStrokeRect,
		X:          10,
		Y:          10,
		W:          120,
		H:          60,
		Width:      2,
		StrokeMask: 0x0F,
		R:          50,
		G:          50,
		B:          50,
		Alpha:      1,
	}

	// Warm cache
	if !r.cachedStroke(dst, op, 0, 0) {
		b.Fatal("failed to cache stroke")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.cachedStroke(dst, op, 0, 0)
	}
}

func BenchmarkCachedStroke_Cold(b *testing.B) {
	r := &Renderer{}
	defer r.Close()
	dst := newParityImage(400, 400)
	defer dst.Deallocate()

	op := &layout.DisplayOp{
		Kind:       layout.DisplayOpStrokeRect,
		X:          10,
		Y:          10,
		W:          120,
		H:          60,
		Width:      2,
		StrokeMask: 0x0F,
		R:          50,
		G:          50,
		B:          50,
		Alpha:      1,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.rasters.mu.Lock()
		for _, img := range r.rasters.entries {
			img.Deallocate()
		}
		clear(r.rasters.entries)
		r.rasters.order = nil
		r.rasters.pixels = 0
		r.rasters.mu.Unlock()

		r.cachedStroke(dst, op, 0, 0)
	}
}

// -----------------------------------------------------------------------------
// 2. Text Drawing & Letter Spacing
// -----------------------------------------------------------------------------

const sampleRunText = "Quick brown fox jumps over lazy dog"

func BenchmarkText_NoSpacing_Warm(b *testing.B) {
	r, op, face, dst := setupTextBench(b, sampleRunText, 0)

	// Warm the cache
	r.cachedTextRun(dst, op, face, op.Text, 20, 40)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.cachedTextRun(dst, op, face, op.Text, 20, 40)
	}
}

func BenchmarkText_NoSpacing_Cold(b *testing.B) {
	r, op, face, dst := setupTextBench(b, sampleRunText, 0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.textRuns.flush()
		r.cachedTextRun(dst, op, face, op.Text, 20, 40)
	}
}

func BenchmarkText_NoSpacing_Direct(b *testing.B) {
	r, op, face, dst := setupTextBench(b, sampleRunText, 0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.drawTextRun(dst, face, op.Text, 20, 40, op)
	}
}

func BenchmarkText_LetterSpacing_Warm(b *testing.B) {
	r, op, face, dst := setupTextBench(b, sampleRunText, 2)

	// Warm the cache
	r.cachedTextRun(dst, op, face, op.Text, 20, 40)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.cachedTextRun(dst, op, face, op.Text, 20, 40)
	}
}

func BenchmarkText_LetterSpacing_Cold(b *testing.B) {
	r, op, face, dst := setupTextBench(b, sampleRunText, 2)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.textRuns.flush()
		r.cachedTextRun(dst, op, face, op.Text, 20, 40)
	}
}

func BenchmarkText_LetterSpacing_Direct(b *testing.B) {
	r, op, face, dst := setupTextBench(b, sampleRunText, 2)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.drawTextRun(dst, face, op.Text, 20, 40, op)
	}
}

// Microbenchmarks for text allocations:

func BenchmarkMicro_RuneToString(b *testing.B) {
	content := sampleRunText
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range content {
			s := string(r)
			_ = s
		}
	}
}

func BenchmarkMicro_AppendLazyGlyphs_FullString(b *testing.B) {
	_, _, face, _ := setupTextBench(b, sampleRunText, 0)
	content := sampleRunText
	var buf []text.LazyGlyph

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = text.AppendLazyGlyphs(buf[:0], content, face, nil)
	}
}

func BenchmarkMicro_AppendLazyGlyphs_PerRune(b *testing.B) {
	_, _, face, _ := setupTextBench(b, sampleRunText, 0)
	content := sampleRunText
	var buf []text.LazyGlyph

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range content {
			glyph := string(r)
			buf = text.AppendLazyGlyphs(buf[:0], glyph, face, nil)
		}
	}
}

func BenchmarkMicro_Measure_FullString(b *testing.B) {
	_, _, face, _ := setupTextBench(b, sampleRunText, 0)
	content := sampleRunText

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = text.Measure(content, face, 0)
	}
}

func BenchmarkMicro_Measure_PerRune(b *testing.B) {
	_, _, face, _ := setupTextBench(b, sampleRunText, 0)
	content := sampleRunText

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range content {
			glyph := string(r)
			_, _ = text.Measure(glyph, face, 0)
		}
	}
}

// -----------------------------------------------------------------------------
// 3. Mix-Blend Shader Compositing
// -----------------------------------------------------------------------------

func BenchmarkCompositeLayer_Normal(b *testing.B) {
	dst := newParityImage(200, 200)
	layer := newParityImage(100, 100)
	defer dst.Deallocate()
	defer layer.Deallocate()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compositeLayer(dst, layer, image.Pt(20, 20), "normal")
	}
}

func BenchmarkCompositeLayer_MixBlend_Multiply(b *testing.B) {
	dst := newParityImage(200, 200)
	layer := newParityImage(100, 100)
	defer dst.Deallocate()
	defer layer.Deallocate()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compositeLayer(dst, layer, image.Pt(20, 20), "multiply")
	}
}

func BenchmarkCompositeLayer_MixBlend_Screen(b *testing.B) {
	dst := newParityImage(200, 200)
	layer := newParityImage(100, 100)
	defer dst.Deallocate()
	defer layer.Deallocate()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compositeLayer(dst, layer, image.Pt(20, 20), "screen")
	}
}

// Microbenchmarks for composite allocations:

func BenchmarkMicro_NewImageAndDeallocate(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img := ebiten.NewImage(100, 100)
		img.Deallocate()
	}
}

func BenchmarkMicro_UniformsMapAlloc(b *testing.B) {
	modeID := 1
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := map[string]any{"Mode": float32(modeID)}
		_ = m
	}
}
