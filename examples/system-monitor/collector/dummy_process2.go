package collector

import "time"

// evolveProc advances one process: CPU time, resident memory, IO counters,
// and state.
func (d *Dummy) evolveProc(p *dmProc) {
	roll := d.rng.Float64()

	delta := time.Duration(d.rng.Int63n(evolveSpanNS))
	if roll < settleChance {
		delta = settleBase + time.Duration(d.rng.Int63n(int64(settleExtra)))
	}
	if d.spike && roll > spikeEvolve {
		delta = spikeEvolveDT
	}
	p.cpu += uint64(delta)

	step := int64(d.rng.Intn(procJitter)) - procJitterMid
	p.mem = clampU64(int64(p.mem)+step, procMemFloor, procMemCeil) //nolint:gosec // reason: resident memory stays within the clamped 1 MiB to 2 GiB range.

	if d.rng.Float64() < readChance {
		p.read += uint64(d.rng.Int63n(readSpan))   //nolint:gosec // reason: Int63n never returns a negative value.
		p.write += uint64(d.rng.Int63n(writeSpan)) //nolint:gosec // reason: Int63n never returns a negative value.
	}

	if d.rng.Float64() < runChance {
		p.state = "running"
	} else {
		p.state = "sleeping"
	}
}

// advanceProcs evolves every process, then exits some and starts others so
// the table changes between samples.
func (d *Dummy) advanceProcs() {
	for _, p := range d.procs {
		d.evolveProc(p)
	}

	d.churn()
}

// churn exits a few processes and replaces them, keeping the table size.
func (d *Dummy) churn() {
	exits := 1 + d.rng.Intn(exitSpan)

	for range exits {
		if len(d.procs) <= procFloor {
			break
		}

		idx := d.rng.Intn(len(d.procs))
		d.retired = append(d.retired, d.procs[idx].id)
		if len(d.retired) > retireCap {
			d.retired = d.retired[1:]
		}
		d.procs = append(d.procs[:idx], d.procs[idx+1:]...)
	}

	for range exits {
		d.spawn()
	}
}
