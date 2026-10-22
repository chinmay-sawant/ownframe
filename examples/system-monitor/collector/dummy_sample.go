package collector

import (
	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// machineSample copies the current machine state into a raw sample. Derived
// fields stay invalid; the manager computes them from consecutive samples.
func (d *Dummy) machineSample(stamp domain.Stamp) domain.Sample {
	s := domain.Sample{Stamp: stamp, Host: "dummy-host", OS: "dummy"}
	if !d.boot.IsZero() && stamp.At.After(d.boot) {
		s.Uptime = stamp.At.Sub(d.boot)
	}

	var total domain.CPUTimes
	for _, c := range d.cpu {
		total.Busy += c.Busy
		total.Total += c.Total
	}
	s.CPUTotal = total
	s.CPUCores = append([]domain.CPUTimes(nil), d.cpu...)
	s.Load = [3]domain.Value{
		domain.Count(d.load[0]),
		domain.Count(d.load[1]),
		domain.Count(d.load[2]),
	}
	s.Mem = domain.Memory{
		Total:     domain.Bytes(float64(d.memTotal)),
		Used:      domain.Bytes(float64(d.memUsed)),
		Available: domain.Bytes(float64(d.memTotal - d.memUsed)),
		Cached:    domain.Bytes(float64(d.memTotal / memQuarterDiv)),
		SwapTotal: domain.Bytes(swapTotal),
		SwapUsed:  domain.Bytes(float64(d.swapUsed)),
	}

	for _, disk := range d.disks {
		s.Disks = append(s.Disks, domain.Disk{
			Device:     disk.name,
			Mount:      disk.mount,
			Total:      domain.Bytes(float64(disk.total)),
			Free:       domain.Bytes(float64(disk.free)),
			ReadBytes:  disk.read,
			WriteBytes: disk.write,
		})
	}

	for _, n := range d.nets {
		s.Nets = append(s.Nets, domain.Net{
			Name:     n.name,
			Up:       n.up,
			RXBytes:  n.rx,
			TXBytes:  n.tx,
			RXErrors: n.errs,
		})
	}

	threads := 0
	for _, p := range d.procs {
		threads += p.threads
	}
	s.Procs = domain.Count(float64(len(d.procs)))
	s.Threads = domain.Count(float64(threads))

	// Two of the three sensors are deliberately unreadable.
	s.Temps = sampleTemps(d.rng)

	return s
}
