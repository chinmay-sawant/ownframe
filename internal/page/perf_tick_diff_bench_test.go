package page

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// makeBenchDisplay creates a Display with n layout.DisplayOpFillRect ops.
func makeBenchDisplay(n int) *layout.Display {
	ops := make([]DisplayOp, n)
	for i := 0; i < n; i++ {
		ops[i] = DisplayOp{
			Kind:  layout.DisplayOpFillRect,
			X:     float64((i % 40) * 20),
			Y:     float64((i / 40) * 20),
			W:     15,
			H:     15,
			R:     0.2,
			G:     0.4,
			B:     0.6,
			Alpha: 1.0,
		}
	}

	return &layout.Display{
		Width:         1000,
		Height:        1000,
		PixelPerPoint: 1,
		Ops:           ops,
	}
}

// makeBenchPage creates a Page populated with a Display having n ops.
func makeBenchPage(n int) *Page {
	p := &Page{
		pageGeom: pageGeom{width: 1000, height: 1000},
	}
	p.draw.Display = makeBenchDisplay(n)

	return p
}

func benchTickCaptureAndDamage(b *testing.B, opCount int, modify bool) {
	b.Helper()
	p := makeBenchPage(opCount)
	gen := p.Generation()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		snap := p.captureTickOps()
		if modify {
			p.draw.Display.Ops[0].R = float64(i%2) * 0.8
		}
		p.noteTickDamage(gen, snap)
		p.freeOps(snap)
		p.dirty.Rects = p.dirty.Rects[:0]
	}
}

func benchTickDiffByPosition(b *testing.B, opCount int) {
	b.Helper()
	prev := makeBenchDisplay(opCount)
	next := makeBenchDisplay(opCount + 1)
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		diffDisplay(prev, next)
	}
}

func BenchmarkTick_CaptureAndDamage_100Ops_Mod0(b *testing.B) {
	benchTickCaptureAndDamage(b, 100, false)
}

func BenchmarkTick_CaptureAndDamage_100Ops_Mod1(b *testing.B) {
	benchTickCaptureAndDamage(b, 100, true)
}

func BenchmarkTick_CaptureAndDamage_500Ops_Mod0(b *testing.B) {
	benchTickCaptureAndDamage(b, 500, false)
}

func BenchmarkTick_CaptureAndDamage_500Ops_Mod1(b *testing.B) {
	benchTickCaptureAndDamage(b, 500, true)
}

func BenchmarkTick_CaptureAndDamage_2000Ops_Mod0(b *testing.B) {
	benchTickCaptureAndDamage(b, 2000, false)
}

func BenchmarkTick_CaptureAndDamage_2000Ops_Mod1(b *testing.B) {
	benchTickCaptureAndDamage(b, 2000, true)
}

func BenchmarkTick_DiffByPosition_100Ops(b *testing.B) {
	benchTickDiffByPosition(b, 100)
}

func BenchmarkTick_DiffByPosition_500Ops(b *testing.B) {
	benchTickDiffByPosition(b, 500)
}

func BenchmarkTick_DiffByPosition_2000Ops(b *testing.B) {
	benchTickDiffByPosition(b, 2000)
}

func benchFullTick(b *testing.B, opCount int, modify bool) {
	b.Helper()
	p := makeBenchPage(opCount)
	ctx := context.Background()

	i := 0
	p.SetTick(func(ctx context.Context) error {
		if modify {
			p.draw.Display.Ops[0].R = float64(i%2) * 0.8
		}
		return nil
	})

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		i++
		if err := p.Tick(ctx); err != nil {
			b.Fatal(err)
		}
		p.dirty.Rects = p.dirty.Rects[:0]
	}
}

func BenchmarkTick_FullTick_100Ops_Mod0(b *testing.B) {
	benchFullTick(b, 100, false)
}

func BenchmarkTick_FullTick_100Ops_Mod1(b *testing.B) {
	benchFullTick(b, 100, true)
}

func BenchmarkTick_FullTick_500Ops_Mod0(b *testing.B) {
	benchFullTick(b, 500, false)
}

func BenchmarkTick_FullTick_500Ops_Mod1(b *testing.B) {
	benchFullTick(b, 500, true)
}

func BenchmarkTick_FullTick_2000Ops_Mod0(b *testing.B) {
	benchFullTick(b, 2000, false)
}

func BenchmarkTick_FullTick_2000Ops_Mod1(b *testing.B) {
	benchFullTick(b, 2000, true)
}
