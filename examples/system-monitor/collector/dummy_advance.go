package collector

import "time"

// advanceCPU adds one interval of scheduler time to every core. A rotating
// core spikes every 37 ticks, which shows up in the overview graph.
func (d *Dummy) advanceCPU(elapsed time.Duration) {
	for i := range d.cpu {
		fraction := loadBase + d.rng.Float64()*loadJitter
		if d.spike && (i+int(d.tick/spikeStride))%spikeParity == 0 {
			fraction = spikeBase + d.rng.Float64()*spikeJitter
		}

		d.cpu[i].Total += uint64(elapsed) //nolint:gosec // reason: tick intervals are non-negative; advance resets non-positive spans to one second.
		d.cpu[i].Busy += uint64(float64(elapsed) * fraction)
	}

	drift := (d.rng.Float64() - driftCenter) * driftAmp
	d.load[0] = clampF(d.load[0]+drift, loadFloor, loadCeil)
	d.load[1] = clampF(d.load[1]*0.97+d.load[0]*0.03, loadFloor, loadCeil)
	d.load[2] = clampF(d.load[2]*0.99+d.load[0]*0.01, loadFloor, loadCeil)
}

// advanceMem walks used memory and swap within sane bounds.
func (d *Dummy) advanceMem() {
	step := int64(d.rng.Intn(memStepRange)) - memStepHalf
	d.memUsed = clampU64(int64(d.memUsed)+step, memLow, memHigh) //nolint:gosec // reason: used memory stays within the clamped GiB range, far below MaxInt64.

	if d.rng.Float64() < swapChance {
		d.swapUsed = clampU64(int64(d.swapUsed)+step/4, 0, swapHigh) //nolint:gosec // reason: swap stays within the clamped GiB range, far below MaxInt64.
	}
}

// clampF clamps a float into [lo, hi].
func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}

	return v
}
