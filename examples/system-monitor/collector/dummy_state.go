package collector

import (
	"time"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// init builds the fixture's starting state. It runs once per source, so the
// same seed always starts from the same counters.
func (d *Dummy) init() {
	d.cores = 8
	d.cpu = make([]domain.CPUTimes, d.cores)
	for i := range d.cpu {
		d.cpu[i] = domain.CPUTimes{Busy: coreBusyStep * uint64(i+coreBusyOffset), Total: coreTotalBase}
	}

	d.load = [3]float64{1.1, 1.4, 1.6}
	d.memTotal = memTotalInit
	d.memUsed = memUsedInit
	d.swapUsed = swapInit
	d.disks = []dmDisk{
		{name: "sda", mount: "/", total: diskSdaTotal, free: diskSdaFree},
		{name: "nvme0n1", mount: "/home", total: diskNvmeTotal, free: diskNvmeFree},
	}
	d.nets = []dmNet{{name: "eth0", up: true}, {name: "wlan0", up: true}}

	d.nextPID = 100
	d.nextStart = 1000

	for range d.opts.Processes {
		d.spawn()
	}
}

// advance moves the simulation to stamp and returns the machine sample.
// Callers hold d.mu.
func (d *Dummy) advance(stamp domain.Stamp) domain.Sample {
	elapsed := stamp.Mono - d.lastMono
	if d.tick == 0 || elapsed <= 0 {
		elapsed = time.Second
	}
	d.lastMono = stamp.Mono

	if d.boot.IsZero() {
		d.boot = stamp.At.Add(-bootBackHours * time.Hour)
	}

	d.tick++
	d.spike = d.tick%spikeEvery < spikeWidth

	d.advanceCPU(elapsed)
	d.advanceMem()
	d.advanceDisks()
	d.advanceNets()
	d.advanceProcs()

	return d.machineSample(stamp)
}

// clampU64 clamps v into [lo, hi].
func clampU64(v, lo, hi int64) uint64 {
	if v < lo {
		v = lo
	}
	if v > hi {
		v = hi
	}

	return uint64(v) //nolint:gosec // reason: every caller passes a non-negative lo, so the clamped result stays non-negative.
}
