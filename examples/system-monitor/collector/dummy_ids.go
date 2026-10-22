package collector

import (
	"time"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// newStart returns the next start value used for process identity.
func (d *Dummy) newStart() uint64 {
	d.nextStart++

	return d.nextStart
}

// startedAt maps a start value to a start time up to three days back.
func (d *Dummy) startedAt(id domain.ProcessIdentity) time.Time {
	const age = bootBackHours * uint64(time.Hour) / uint64(time.Second)

	return d.boot.Add(-time.Duration(id.Start%age) * time.Second)
}
