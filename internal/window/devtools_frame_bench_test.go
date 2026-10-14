package window

import "testing"

// BenchmarkDevFrameRows measures the Frame tab's per-frame row build, which
// runs on every frame while the dock is open.
func BenchmarkDevFrameRows(b *testing.B) {
	s := newDevShell(newDevScreen())
	b.ReportAllocs()

	for b.Loop() {
		if len(s.devFrameRows(320)) == 0 {
			b.Fatal("no rows")
		}
	}
}
