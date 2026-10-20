package main

import (
	"sort"
	"testing"
)

// BenchmarkWindowedPointerCycle measures the real 43-click desktop walk:
// windowed catalog, perf on, batched press/focus/click, and the page redraw
// reported by LastRedraw. It mirrors scripts/css-preview-click-test.py.
func BenchmarkWindowedPointerCycle(b *testing.B) {
	ctx, p, v, d, m := windowBenchPage(b)
	if !p.Perf() || len(v.Rows) >= len(v.Catalog) {
		b.Fatal("catalog window or perf mode is not active")
	}
	windowBenchPass(ctx, p, v, d, m, b)

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
