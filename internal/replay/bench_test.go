package replay

import (
	"context"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"image"
	"image/color"
	"os"
	"strings"
	"testing"
)

// BenchmarkDrawRectScan times the op filter on the platform example, once
// for a counter-sized dirty rect and once for the whole frame.
func BenchmarkDrawRectScan(b *testing.B) {
	html, err := os.ReadFile("testdata/platform.html")
	if err != nil {
		b.Fatal(err)
	}

	display, err := render.DisplayListState(context.Background(), string(html), 480, 640, render.State{})
	if err != nil {
		b.Fatal(err)
	}

	box, found := boxByID(display.Boxes, "count")
	if !found {
		b.Fatal("no #count box")
	}

	ppt := display.PointsPerPixel
	dirty := image.Rect(int(box.X)-8, int(box.Y)-8, int(box.X+box.W)+8, int(box.Y+box.H)+8)
	full := image.Rect(0, 0, display.Width, display.Height)

	b.Logf("ops=%d dirty=%v hits=%d", len(display.Ops), dirty, countTouches(display, ppt, dirty))

	b.Run("dirty", func(b *testing.B) { scanOps(b, display, ppt, dirty) })
	b.Run("full", func(b *testing.B) { scanOps(b, display, ppt, full) })
}

func countTouches(display *layout.Display, ppt float64, rect image.Rectangle) int {
	n := 0

	for i := range display.Ops {
		if opTouches(&display.Ops[i], ppt, rect) {
			n++
		}
	}

	return n
}

func scanOps(b *testing.B, display *layout.Display, ppt float64, rect image.Rectangle) {
	for i := 0; i < b.N; i++ {
		for j := range display.Ops {
			opTouches(&display.Ops[j], ppt, rect)
		}
	}
}

func boxByID(boxes []layout.Box, id string) (layout.Box, bool) {
	for _, b := range boxes {
		if b.ID == id {
			return b, true
		}
	}

	return layout.Box{}, false
}

// BenchmarkChunkCardDamage damages one card of a 20-card grid and draws it
// into an image, direct and chunked, reporting the draws and time per call.
func BenchmarkChunkCardDamage(b *testing.B) {
	d := cardBenchDisplay()
	rect := image.Rect(benchGap-2, benchGap-2,
		benchGap+benchCardW+2, benchGap+benchCardH+2)
	dst := ebiten.NewImage(d.Width, d.Height)
	defer dst.Deallocate()

	replayCounts.on.Store(true)
	defer replayCounts.on.Store(false)

	_, runs, members := planStats(chunkPlanFor(d))
	b.Logf("ops=%d runs=%d run-ops=%d", len(d.Ops), runs, members)

	b.Run("direct", func(b *testing.B) {
		b.ReportAllocs()
		replayCounts.ops.Store(0)
		replayCounts.blits.Store(0)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			drawDirectRect(dst, d, rect)
		}

		b.StopTimer()
		b.ReportMetric(float64(replayCounts.ops.Load())/float64(b.N), "ops/draw")
		b.ReportMetric(float64(replayCounts.blits.Load())/float64(b.N), "blits/draw")
	})

	b.Run("chunked", func(b *testing.B) {
		b.ReportAllocs()
		replayCounts.ops.Store(0)
		replayCounts.blits.Store(0)
		DrawRect(dst, d, rect, 0, 0)
		bakeOps := replayCounts.ops.Load()
		bakeBlits := replayCounts.blits.Load()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			DrawRect(dst, d, rect, 0, 0)
		}

		b.StopTimer()
		b.ReportMetric(float64(replayCounts.ops.Load()-bakeOps)/float64(b.N), "ops/draw")
		b.ReportMetric(float64(replayCounts.blits.Load()-bakeBlits)/float64(b.N), "blits/draw")
		b.Logf("bake ops=%d", bakeOps)
	})
}

// drawDirectRect is the pre-chunk damage path: filter the order and draw
// every op that touches the rect.
func drawDirectRect(dst *ebiten.Image, d *layout.Display, rect image.Rectangle) {
	ppt := d.PointsPerPixel

	for _, i := range d.Order {
		op := &d.Ops[i]
		if op.Fixed || !opTouches(op, ppt, rect) {
			continue
		}

		countOp()
		drawOp(dst, op, 0, 0)
	}
}

const (
	benchCards = 20
	benchCardW = 200
	benchCardH = 90
	benchGap   = 10
	benchCols  = 4
)

// cardBenchDisplay builds a 20-card grid; one card folds into one run.
func cardBenchDisplay() *layout.Display {
	pt := func(css float64) float64 { return css * 72.0 / 96.0 }

	rows := benchCards / benchCols
	d := &layout.Display{
		Width:          benchCols*benchCardW + (benchCols+1)*benchGap,
		Height:         rows*benchCardH + (rows+1)*benchGap,
		PointsPerPixel: 72.0 / 96.0,
	}

	for i := range benchCards {
		x := float64(benchGap + (i%benchCols)*(benchCardW+benchGap))
		y := float64(benchGap + (i/benchCols)*(benchCardH+benchGap))
		d.Boxes = append(d.Boxes, layout.Box{
			ID: fmt.Sprintf("card%d", i), X: x, Y: y, W: benchCardW, H: benchCardH,
		})

		base := len(d.Ops)
		d.Ops = append(d.Ops,
			layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: pt(x + 2), Y: pt(y + 2),
				W: pt(benchCardW - 4), H: pt(benchCardH - 4), R: .9, G: .92, B: .96},
			layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: pt(x + 2), Y: pt(y + 2),
				W: pt(benchCardW - 4), H: pt(22), Radius: pt(4), R: .2, G: .4, B: .8},
			layout.DisplayOp{Kind: layout.DisplayOpStrokeRect, X: pt(x + 2), Y: pt(y + 2),
				W: pt(benchCardW - 4), H: pt(benchCardH - 4), Width: pt(2), R: .2, G: .25, B: .3},
			layout.DisplayOp{Kind: layout.DisplayOpLine, X: pt(x + 6), Y: pt(y + 40),
				W: pt(benchCardW - 12), Width: pt(1), R: .7, G: .7, B: .75},
			layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: pt(x + 8), Y: pt(y + 58),
				W: pt(40), H: pt(16), R: .1, G: .6, B: .3},
		)

		for j := base; j < len(d.Ops); j++ {
			d.Order = append(d.Order, j)
		}
	}

	return d
}

// BenchmarkBuildChunks times the plan fold (GO-ALLOC-02 gate).
func BenchmarkBuildChunks(b *testing.B) {
	d := cardBenchDisplay()
	root := buildPlan(d)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildChunks(d, root)
	}
}

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

// makeLayerBenchDisplay constructs a layout.Display with totalOps display ops
// and 4 promoted layer boxes (cards). If nonContiguous is true, the cards'
// owned operations are interleaved with non-owned background operations so that
// lastRank - firstRank + 1 != len(owned), exercising layerOrderSafe's map allocation.
func makeLayerBenchDisplay(totalOps int, nonContiguous bool) (*layout.Display, []LayerSpec) {
	if totalOps < 20 {
		totalOps = 20
	}

	const (
		numLayers   = 4
		opsPerLayer = 4
		layerWidth  = 100
		layerHeight = 100
	)

	d := &layout.Display{
		Width:          1200,
		Height:         1200,
		PointsPerPixel: 1,
		PixelPerPoint:  1,
		Boxes: []layout.Box{
			{ID: "page", X: 0, Y: 0, W: 1200, H: 1200, ParentIndex: -1},
		},
		Ops:      make([]layout.DisplayOp, 0, totalOps),
		OpOwners: make([]int, 0, totalOps),
		Order:    make([]int, 0, totalOps),
	}

	specs := make([]LayerSpec, numLayers)
	for i := 0; i < numLayers; i++ {
		boxX := float64(50 + i*150)
		boxY := float64(50)
		boxID := fmt.Sprintf("card%d", i)
		d.Boxes = append(d.Boxes, layout.Box{
			ID:          boxID,
			X:           boxX,
			Y:           boxY,
			W:           layerWidth,
			H:           layerHeight,
			ParentIndex: 0,
		})
		specs[i] = LayerSpec{
			ElementID: boxID,
			BoxIndex:  i + 1,
			Matrix:    layerTranslate(float64(4+i*4), 0),
			Alpha:     1,
		}
	}

	// Create 4 ops for each of the 4 layers (16 ops total)
	type layerOpInfo struct {
		op    layout.DisplayOp
		owner int
	}
	var layerOps []layerOpInfo
	for i := 0; i < numLayers; i++ {
		owner := i + 1
		boxX := float64(50 + i*150)
		boxY := float64(50)
		layerOps = append(layerOps,
			// Op 0: Background fill
			layerOpInfo{
				op: layout.DisplayOp{
					Kind: layout.DisplayOpFillRect,
					X:    boxX, Y: boxY, W: layerWidth, H: layerHeight,
					R: 0.1, G: 0.2, B: float64(i+1) * 0.2, Alpha: 1,
				},
				owner: owner,
			},
			// Op 1: Header fill
			layerOpInfo{
				op: layout.DisplayOp{
					Kind: layout.DisplayOpFillRect,
					X:    boxX + 5, Y: boxY + 5, W: layerWidth - 10, H: 25,
					R: 0.8, G: 0.8, B: 0.8, Alpha: 1,
				},
				owner: owner,
			},
			// Op 2: Inner content
			layerOpInfo{
				op: layout.DisplayOp{
					Kind: layout.DisplayOpFillRect,
					X:    boxX + 10, Y: boxY + 35, W: layerWidth - 20, H: 30,
					R: 0.3, G: 0.6, B: 0.9, Alpha: 1,
				},
				owner: owner,
			},
			// Op 3: Badge
			layerOpInfo{
				op: layout.DisplayOp{
					Kind: layout.DisplayOpFillRect,
					X:    boxX + 10, Y: boxY + 70, W: 30, H: 20,
					R: 0.9, G: 0.3, B: 0.2, Alpha: 1,
				},
				owner: owner,
			},
		)
	}

	remainingOps := totalOps - len(layerOps)

	if !nonContiguous {
		// Contiguous layout: all layer ops placed first consecutively
		for _, lo := range layerOps {
			idx := len(d.Ops)
			d.Ops = append(d.Ops, lo.op)
			d.OpOwners = append(d.OpOwners, lo.owner)
			d.Order = append(d.Order, idx)
		}
		// Remaining page background/other ops
		for i := 0; i < remainingOps; i++ {
			idx := len(d.Ops)
			d.Ops = append(d.Ops, layout.DisplayOp{
				Kind: layout.DisplayOpFillRect,
				X:    float64(10 + (i%20)*50),
				Y:    float64(300 + (i/20)*30),
				W:    40, H: 20,
				R: 0.5, G: 0.5, B: 0.5, Alpha: 1,
			})
			d.OpOwners = append(d.OpOwners, 0)
			d.Order = append(d.Order, idx)
		}
	} else {
		// Non-contiguous layout: interleave background ops between layer ops so
		// that each layer's ops are separated in d.Order by non-overlapping page ops.
		bgIndex := 0
		for layerIdx := 0; layerIdx < numLayers; layerIdx++ {
			// Add 2 ops for this layer
			for k := 0; k < 2; k++ {
				lo := layerOps[layerIdx*opsPerLayer+k]
				idx := len(d.Ops)
				d.Ops = append(d.Ops, lo.op)
				d.OpOwners = append(d.OpOwners, lo.owner)
				d.Order = append(d.Order, idx)
			}
			// Insert an intervening background op outside the cards (e.g. Y >= 300)
			if bgIndex < remainingOps {
				idx := len(d.Ops)
				d.Ops = append(d.Ops, layout.DisplayOp{
					Kind: layout.DisplayOpFillRect,
					X:    float64(10 + (bgIndex%20)*50),
					Y:    float64(300 + (bgIndex/20)*30),
					W:    40, H: 20,
					R: 0.5, G: 0.5, B: 0.5, Alpha: 1,
				})
				d.OpOwners = append(d.OpOwners, 0)
				d.Order = append(d.Order, idx)
				bgIndex++
			}
			// Add remaining 2 ops for this layer
			for k := 2; k < 4; k++ {
				lo := layerOps[layerIdx*opsPerLayer+k]
				idx := len(d.Ops)
				d.Ops = append(d.Ops, lo.op)
				d.OpOwners = append(d.OpOwners, lo.owner)
				d.Order = append(d.Order, idx)
			}
		}
		// Append remaining background ops
		for ; bgIndex < remainingOps; bgIndex++ {
			idx := len(d.Ops)
			d.Ops = append(d.Ops, layout.DisplayOp{
				Kind: layout.DisplayOpFillRect,
				X:    float64(10 + (bgIndex%20)*50),
				Y:    float64(300 + (bgIndex/20)*30),
				W:    40, H: 20,
				R: 0.5, G: 0.5, B: 0.5, Alpha: 1,
			})
			d.OpOwners = append(d.OpOwners, 0)
			d.Order = append(d.Order, idx)
		}
	}

	return d, specs
}

func TestLayerBenchmarkSetup(t *testing.T) {
	for _, count := range []int{100, 500, 1000} {
		for _, nonContig := range []bool{false, true} {
			d, specs := makeLayerBenchDisplay(count, nonContig)
			var renderer Renderer
			defer renderer.Close()

			damage := renderer.BeginLayers(d, 1, specs)
			if len(damage) == 0 {
				t.Fatalf("count=%d nonContig=%v: expected damage", count, nonContig)
			}
			stats := renderer.Layers()
			if stats.Active != 4 {
				t.Fatalf("count=%d nonContig=%v: expected 4 active layers, got %d (stats: %+v)", count, nonContig, stats.Active, stats)
			}
		}
	}
}

// BenchmarkLayerBeginLayers measures steady-state per-frame allocations and latency
// of BeginLayers during animation with 4 retained layers over 100, 500, and 1000 operations.
func BenchmarkLayerBeginLayers(b *testing.B) {
	for _, count := range []int{100, 500, 1000} {
		b.Run(fmt.Sprintf("ops_%d/contiguous", count), func(b *testing.B) {
			d, specs := makeLayerBenchDisplay(count, false)
			var renderer Renderer
			defer renderer.Close()

			// Warm up: initial promotion & bake
			renderer.BeginLayers(d, 1, specs)
			if renderer.Layers().Active != 4 {
				b.Fatalf("expected 4 active layers, got %d", renderer.Layers().Active)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Simulate per-frame 2D transform animation: matrix translation moves each frame
				specs[0].Matrix = layerTranslate(float64(4+(i&31)), 0)
				renderer.BeginLayers(d, 1, specs)
			}
		})

		b.Run(fmt.Sprintf("ops_%d/non_contiguous", count), func(b *testing.B) {
			d, specs := makeLayerBenchDisplay(count, true)
			var renderer Renderer
			defer renderer.Close()

			// Warm up: initial promotion & bake
			renderer.BeginLayers(d, 1, specs)
			if renderer.Layers().Active != 4 {
				b.Fatalf("expected 4 active layers, got %d", renderer.Layers().Active)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Simulate per-frame 2D transform animation: matrix translation moves each frame
				specs[0].Matrix = layerTranslate(float64(4+(i&31)), 0)
				renderer.BeginLayers(d, 1, specs)
			}
		})
	}
}

// BenchmarkLayerDrawLayers measures allocation overhead when compositing retained
// layers onto the destination buffer (DrawLayers).
func BenchmarkLayerDrawLayers(b *testing.B) {
	for _, count := range []int{100, 500, 1000} {
		b.Run(fmt.Sprintf("ops_%d/normal", count), func(b *testing.B) {
			d, specs := makeLayerBenchDisplay(count, false)
			var renderer Renderer
			defer renderer.Close()

			renderer.BeginLayers(d, 1, specs)
			if renderer.Layers().Active != 4 {
				b.Fatalf("expected 4 active layers, got %d", renderer.Layers().Active)
			}

			dst := ebiten.NewImage(d.Width, d.Height)
			defer dst.Deallocate()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				renderer.DrawLayers(dst, d, 0, 0)
			}
		})
	}

	b.Run("blend_mode", func(b *testing.B) {
		d, specs := makeLayerBenchDisplay(100, false)
		specs[0].Mode = "multiply"
		var renderer Renderer
		defer renderer.Close()

		renderer.BeginLayers(d, 1, specs)
		if renderer.Layers().Active != 4 {
			b.Fatalf("expected 4 active layers, got %d", renderer.Layers().Active)
		}

		dst := ebiten.NewImage(d.Width, d.Height)
		defer dst.Deallocate()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			renderer.DrawLayers(dst, d, 0, 0)
		}
	})
}

// textBenchDisplay builds a text-heavy page: 48 short label rows, the shape a
// scrolled list or a card grid repaints.
func textBenchDisplay(b *testing.B) *layout.Display {
	var sb strings.Builder
	sb.WriteString(`<html><body style="margin:0;font-size:13px">`)
	for i := range 48 {
		fmt.Fprintf(&sb, `<p style="margin:2px">Row %02d service west-1 healthy 12 ms</p>`, i)
	}
	sb.WriteString(`</body></html>`)

	d, err := render.DisplayList(context.Background(), sb.String(), 520, 900)
	if err != nil {
		b.Fatal(err)
	}

	return d
}

// BenchmarkTextDamageRedraw times a damage repaint over text rows: the same
// display list drawn again, so the second draw finds the runs already baked.
// warm redraws at one position, miss flushes the cache first, and scroll
// moves the page one whole pixel per step, which the key's sub-pixel grid has
// to absorb.
func BenchmarkTextDamageRedraw(b *testing.B) {
	d := textBenchDisplay(b)
	dst := ebiten.NewImage(d.Width, d.Height)
	defer dst.Deallocate()
	rect := image.Rect(0, 0, d.Width, d.Height)

	b.ReportAllocs()
	b.Run("warm", func(b *testing.B) {
		DrawRect(dst, d, rect, 0, 0)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			DrawRect(dst, d, rect, 0, 0)
		}
	})
	b.Run("miss", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			textRuns.flush()
			DrawRect(dst, d, rect, 0, 0)
		}
	})
	b.Run("scroll", func(b *testing.B) {
		DrawRect(dst, d, rect, 0, 0)

		for i := 0; i < b.N; i++ {
			DrawRect(dst, d, rect, 0, -float64(i%800))
		}
	})
}
