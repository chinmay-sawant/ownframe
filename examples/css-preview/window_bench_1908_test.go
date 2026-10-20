package main

import (
	"sort"
	"testing"
)

// BenchmarkWindowedPointerCycle1908 measures the same walk in the 1908x999
// window used for the desktop perf checks.
func BenchmarkWindowedPointerCycle1908(b *testing.B) {
	ctx, p, v, d, m := windowBenchPageSize(b, 1908, 999)
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
	b.ReportMetric(sum/float64(len(values)), "ms-redraw/click")
	b.ReportMetric(values[len(values)/2], "p50-redraw/click")
	b.ReportMetric(values[len(values)*95/100], "p95-redraw/click")
	b.ReportMetric(values[len(values)-1], "max-redraw/click")
	b.ReportMetric(sum/1000, "s-redraw/pass")
}
