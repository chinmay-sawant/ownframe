package main

import (
	"sort"
	"testing"
	"time"
)

// BenchmarkWindowedPointerCycle measures the real 43-click desktop walk:
// windowed catalog, perf on, batched press/focus/click, and the page redraw
// reported by LastRedraw. The number is the redraw the page reports, not the
// press to painted interval the click script measures.
func BenchmarkWindowedPointerCycle(b *testing.B) {
	ctx, p, v, d, m := windowBenchPage(b)
	if !p.Window().Perf() || len(v.Rows) >= len(v.Catalog) {
		b.Fatal("catalog window or perf mode is not active")
	}
	windowBenchPass(ctx, p, v, d, m, b)

	// The pass resets the page before it walks, and the reset is two full
	// redraws that no click pays for. The reset is timed on its own and
	// taken out of the pass, so ns/op counts the clicks.
	b.StopTimer()
	resetStart := time.Now()
	windowBenchReset(ctx, p, v, d, b)
	resetCost := time.Since(resetStart)
	b.StartTimer()

	b.ResetTimer()
	values := make([]float64, 0, b.N*len(v.Catalog))
	sum := 0.0
	for i := 0; i < b.N; i++ {
		for _, r := range windowBenchPass(ctx, p, v, d, m, b) {
			ms := float64(r) / 1e6
			values = append(values, ms)
			sum += ms
		}
	}
	b.StopTimer()

	clicks := b.Elapsed().Nanoseconds() - int64(b.N)*resetCost.Nanoseconds()
	if clicks < 0 {
		clicks = 0
	}
	b.ReportMetric(float64(clicks)/float64(b.N), "ns/op")

	sort.Float64s(values)
	if len(values) == 0 {
		b.Fatal("no clicks measured")
	}
	avg := sum / float64(len(values))
	b.ReportMetric(avg, "ms-redraw/click")
	b.ReportMetric(values[len(values)/2], "p50-redraw/click")
	b.ReportMetric(values[len(values)*95/100], "p95-redraw/click")
	b.ReportMetric(values[len(values)-1], "max-redraw/click")
	b.ReportMetric(sum/1000, "s-redraw/pass")
}
