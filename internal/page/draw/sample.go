package draw

import (
	"runtime"
	"time"
)

// Sample is the allocator reading and the clock for one Redraw.
type Sample struct {
	Track bool
	Mem   runtime.MemStats
	Start time.Time
}

// BeginSample reads the allocator when perf is on.
// clear is true when perf is off: the caller zeros opt-in counters before
// starting the clock, so an untracked redraw drops stale numbers and the
// clock does not include that zeroing.
func BeginSample(perf bool) (Sample, bool) {
	if perf {
		return Sample{Track: true, Mem: AllocStart()}, false
	}

	return Sample{}, true
}

// StartClock marks the beginning of the redraw. Call it after any counter
// clear so the clear is not part of LastRedraw.
func (s *Sample) StartClock() { s.Start = time.Now() }
