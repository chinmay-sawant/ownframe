package collector

import (
	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// dmProc is one simulated process run.
type dmProc struct {
	id                domain.ProcessIdentity
	name, user, state string
	cpu, mem          uint64
	threads, nice     int
	read, write       uint64
}

// spawn adds one process. A retired identity comes back with a new start
// value about a third of the time, which exercises PID reuse.
func (d *Dummy) spawn() {
	var id domain.ProcessIdentity
	if len(d.retired) > 0 && d.rng.Float64() < 0.35 {
		j := d.rng.Intn(len(d.retired))
		id = d.retired[j]
		d.retired = append(d.retired[:j], d.retired[j+1:]...)
		id.Start = d.newStart()
	} else {
		d.nextPID++
		id = domain.ProcessIdentity{PID: d.nextPID, Start: d.newStart()}
	}

	d.procs = append(d.procs, &dmProc{
		id:      id,
		name:    procNames[d.rng.Intn(len(procNames))],
		user:    procUsers[d.rng.Intn(len(procUsers))],
		state:   "sleeping",
		cpu:     uint64(d.rng.Int63n(cpuCapNS)),                  //nolint:gosec // reason: Int63n never returns a negative value.
		mem:     procMemBase + uint64(d.rng.Int63n(procMemSpan)), //nolint:gosec // reason: Int63n never returns a negative value.
		threads: 1 + d.rng.Intn(maxNewThreads),
		nice:    d.rng.Intn(niceSpan),
	})
}
