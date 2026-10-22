package domain

import "time"

// Compute returns next with its derived fields filled from prev. Rates divide
// a counter delta by monotonic elapsed time. The first sample of an era, a
// zero elapsed time, a counter that went backwards, a changed core count, and
// a gap all leave the affected values invalid rather than wrong.
func Compute(prev, next Sample) Sample {
	next.Gap = Gap(prev.Stamp, next.Stamp, DefaultGap)
	next.Mem = DeriveMemory(next.Mem)

	elapsed, ok := Elapsed(prev.Stamp, next.Stamp)
	if !ok || next.Gap {
		return next
	}

	if p, ok := cpuPercent(prev.CPUTotal, next.CPUTotal, elapsed); ok {
		next.CPUPercent = p
	}
	if len(prev.CPUCores) == len(next.CPUCores) {
		next.CorePercent = make([]Value, len(next.CPUCores))
		for i := range next.CPUCores {
			next.CorePercent[i], _ = cpuPercent(prev.CPUCores[i], next.CPUCores[i], elapsed)
		}
	} else if len(next.CPUCores) > 0 {
		// A core came online or went offline: keep one sample of invalid
		// per-core values so the graphs break instead of shifting.
		next.CorePercent = make([]Value, len(next.CPUCores))
	}

	diskRates(prev.Disks, next.Disks, elapsed)
	netRates(prev.Nets, next.Nets, elapsed)

	return next
}

// cpuPercent returns busy time over total time for one interval.
func cpuPercent(prev, next CPUTimes, elapsed time.Duration) (Value, bool) {
	if next.Total <= prev.Total || next.Busy < prev.Busy || elapsed <= 0 {
		return Value{}, false
	}

	total := next.Total - prev.Total
	busy := next.Busy - prev.Busy
	if busy > total {
		return Value{}, false
	}

	return Percent(float64(busy) / float64(total) * percentScale), true
}
