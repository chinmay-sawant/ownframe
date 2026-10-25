package main

import (
	"testing"
)

// BenchmarkCostFrame43Animations ticks the page with all 43 previews
// running and reports what one frame of the page costs. The window's draw
// is measured by TestCostWindowFrames; this is the page side of a frame.
func BenchmarkCostFrame43Animations(b *testing.B) {
	ctx, p, v, d, _ := windowBenchPage(b)
	costFullCatalog(ctx, p, v, d, b)
	costPlayAll(ctx, p, b)
	if err := p.Testing().Tick(ctx); err != nil {
		b.Fatal(err)
	}
	if got := costActive(p) + costManual(d); got < 40 {
		b.Fatalf("%d previews running, want the whole catalog", got)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := p.Testing().Tick(ctx); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(costActive(p)+costManual(d)), "active")
	b.ReportMetric(float64(len(v.Catalog)), "rows")
}

// BenchmarkCostFrameIdle ticks the page with no preview running, the
// frame the window draws when nothing moves.
func BenchmarkCostFrameIdle(b *testing.B) {
	ctx, p, _, d, _ := windowBenchPage(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := p.Testing().Tick(ctx); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(costActive(p)+costManual(d)), "active")
}
