package domain

import "time"

// diskRates fills space percentages and IO rates in place, matching devices by
// name. A device that just appeared keeps an invalid rate.
func diskRates(prev, next []Disk, elapsed time.Duration) {
	before := make(map[string]Disk, len(prev))
	for _, d := range prev {
		before[d.Device] = d
	}

	for i := range next {
		d := &next[i]
		if total, ok := d.Total.Get(); ok && total > 0 {
			if free, ok := d.Free.Get(); ok {
				d.UsedPercent = Percent((total - free) / total * percentScale)
			}
		}

		p, ok := before[d.Device]
		if !ok {
			continue
		}
		if d.ReadBytes >= p.ReadBytes {
			d.ReadRate = Rate(float64(d.ReadBytes-p.ReadBytes) / elapsed.Seconds())
		}
		if d.WriteBytes >= p.WriteBytes {
			d.WriteRate = Rate(float64(d.WriteBytes-p.WriteBytes) / elapsed.Seconds())
		}
	}
}

// netRates fills interface rates in place, matching by name. A hot-plugged
// interface keeps an invalid rate for its first sample.
func netRates(prev, next []Net, elapsed time.Duration) {
	before := make(map[string]Net, len(prev))
	for _, n := range prev {
		before[n.Name] = n
	}

	for i := range next {
		n := &next[i]

		p, ok := before[n.Name]
		if !ok {
			continue
		}
		if n.RXBytes >= p.RXBytes {
			n.RXRate = Rate(float64(n.RXBytes-p.RXBytes) / elapsed.Seconds())
		}
		if n.TXBytes >= p.TXBytes {
			n.TXRate = Rate(float64(n.TXBytes-p.TXBytes) / elapsed.Seconds())
		}
	}
}
