package draw

import (
	"runtime"
	"testing"
)

func TestAllocUsedWrap(t *testing.T) {
	t.Parallel()

	var m0 runtime.MemStats
	m0.TotalAlloc = ^uint64(0)
	if AllocUsed(m0) != 0 {
		t.Fatal("wrap")
	}
}

func TestAllocStartSamples(t *testing.T) {
	t.Parallel()

	if AllocStart().TotalAlloc == 0 {
		t.Fatal("zero")
	}
}
