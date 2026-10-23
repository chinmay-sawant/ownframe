package draw

import "runtime"

// AllocStart samples the allocator before a tracked Redraw.
func AllocStart() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return m
}

// AllocUsed is the bytes allocated since AllocStart sampled m0.
func AllocUsed(m0 runtime.MemStats) uint64 {
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	if m1.TotalAlloc < m0.TotalAlloc {
		return 0
	}

	return m1.TotalAlloc - m0.TotalAlloc
}
