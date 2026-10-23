package draw

import (
	"runtime"
	"time"
)

// Sample is the clock for one Redraw and, when MemOK is set, an allocator reading.
// Track stays set for every Perf redraw. MemOK is set only on a redraw that
// read the allocator.
type Sample struct {
	Track bool
	MemOK bool
	Mem   runtime.MemStats
	Start time.Time
}

// BeginSample reads the allocator when perf is on and memDue is true.
// clear is true when perf is off: the caller zeros opt-in counters before
// starting the clock, so an untracked redraw drops stale numbers and the
// clock does not include that zeroing. A Perf redraw that is not due keeps
// Track set, leaves MemOK false, and does not read the allocator.
func BeginSample(perf, memDue bool) (Sample, bool) {
	if !perf {
		return Sample{}, true
	}
	if !memDue {
		return Sample{Track: true}, false
	}

	return Sample{Track: true, MemOK: true, Mem: AllocStart()}, false
}

// StartClock marks the beginning of the redraw. Call it after any counter
// clear so the clear is not part of LastRedraw.
func (s *Sample) StartClock() { s.Start = time.Now() }
