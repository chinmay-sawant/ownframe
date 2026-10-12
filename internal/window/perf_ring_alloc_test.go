package window

import (
	"testing"
	"time"
)

// TestPerfRingSummaryNoAlloc pins the point of the scratch buffer: a summary
// on a full ring must not allocate, because the frame loop runs it.
func TestPerfRingSummaryNoAlloc(t *testing.T) {
	var r perfRing
	for i := range perfRingCap {
		r.record(time.Duration(i) * time.Millisecond)
	}

	allocs := testing.AllocsPerRun(100, func() {
		avg, p95, p99, long := r.summary()
		if avg < 0 || p95 < 0 || p99 < 0 || long < 0 {
			t.Fatal("negative summary")
		}
	})

	if allocs != 0 {
		t.Fatalf("summary allocates %.1f times per call, want 0", allocs)
	}
}
