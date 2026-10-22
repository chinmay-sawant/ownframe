package domain

import "time"

// ComputeProcesses fills CPU percent on next from a previous cumulative read,
// matching by identity. A new identity, a reused PID, or a backwards counter
// keeps an invalid value. CPU is percent of one logical core, so a process
// busy on several cores can exceed 100.
func ComputeProcesses(prev map[ProcessIdentity]uint64, next []Process, elapsed time.Duration) {
	if elapsed <= 0 {
		return
	}

	for i := range next {
		before, ok := prev[next[i].ID]
		if !ok || next[i].CPUTime < before {
			continue
		}

		next[i].CPU = Percent(float64(next[i].CPUTime-before) / float64(elapsed) * percentScale)
	}
}

// ComputeDetail fills the IO rates of one tracked process from its previous
// read. A different identity or a backwards counter keeps them invalid.
func ComputeDetail(prev, next ProcessDetail, elapsed time.Duration) ProcessDetail {
	if next.ID != prev.ID || elapsed <= 0 {
		return next
	}

	if read, ok := prev.ReadBytes.Get(); ok && next.ReadBytes.Valid && next.ReadBytes.N >= read {
		next.ReadRate = Rate((next.ReadBytes.N - read) / elapsed.Seconds())
	}
	if write, ok := prev.WriteBytes.Get(); ok && next.WriteBytes.Valid && next.WriteBytes.N >= write {
		next.WriteRate = Rate((next.WriteBytes.N - write) / elapsed.Seconds())
	}

	return next
}
