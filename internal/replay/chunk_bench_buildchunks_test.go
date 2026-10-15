package replay

import "testing"

// BenchmarkBuildChunks times the chunk-plan fold on the 20-card grid:
// flatten plus per-node run packing. buildPlan runs once outside the
// timer so the numbers isolate buildChunks allocs (GO-ALLOC-02 gate).
func BenchmarkBuildChunks(b *testing.B) {
	d := cardBenchDisplay()
	root := buildPlan(d)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildChunks(d, root)
	}
}
