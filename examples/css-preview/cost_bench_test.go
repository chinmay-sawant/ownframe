package main

import (
	"testing"
)

// BenchmarkCostClickWindowed clicks a Start button with the catalog row
// window active, the state the window keeps between clicks.
func BenchmarkCostClickWindowed(b *testing.B) {
	benchCostClick(b, false)
}

// BenchmarkCostClickFullCatalog clicks a Start button with all 43 rows
// rendered, the state Play All leaves behind.
func BenchmarkCostClickFullCatalog(b *testing.B) {
	benchCostClick(b, true)
}

// benchCostClick clicks one Start button again and again and reports the
// redraw time, the allocation, and the display size the click paid for.
// The window is the desktop one, 1908x999.
func benchCostClick(b *testing.B, full bool) {
	ctx, p, v, d, _ := windowBenchPageSize(b, 1908, 999)
	if full {
		costFullCatalog(ctx, p, v, d, b)
	}
	row := v.Catalog[3]
	redraws := make([]float64, 0, b.N)
	allocs := make([]float64, 0, b.N)
	ops := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		windowBenchReset(ctx, p, v, d, b)
		b.StartTimer()
		dur, alloc := costClick(ctx, p, b, row.Button)
		redraws = append(redraws, float64(dur)/1e6)
		allocs = append(allocs, float64(alloc)/1024)
		ops = p.Stats().Ops
	}
	b.StopTimer()
	costReport(b, "redraw/click", redraws, "ms")
	costReport(b, "alloc/click", allocs, "KB")
	b.ReportMetric(float64(ops), "ops")
	b.ReportMetric(float64(len(v.Catalog)), "rows")
}
