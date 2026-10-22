//go:build linux

package collector

import (
	"time"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// stateName expands the single-letter process state.
func stateName(state string) string {
	switch state {
	case "R":
		return "running"
	case "S":
		return "sleeping"
	case "D":
		return "disk sleep"
	case "Z":
		return "zombie"
	case "T":
		return "stopped"
	case "t":
		return "tracing"
	case "X", "x":
		return "dead"
	case "I":
		return "idle"
	default:
		return "unknown"
	}
}

// startedAt converts a start tick count to wall clock time.
func startedAt(boot time.Time, startTicks uint64) time.Time {
	return boot.Add(time.Duration(startTicks) * time.Second / userHZ) //nolint:gosec // reason: boot-relative start ticks fit int64 for any realistic uptime.
}

// rowOfStat builds the table row for one stat read. The user name is left
// empty here; the detail view fills it from /etc/passwd.
func (s *procSource) rowOfStat(pid int32, st procStat, boot time.Time) domain.Process {
	return domain.Process{
		ID:        domain.ProcessIdentity{PID: pid, Start: st.startTicks},
		Name:      st.comm,
		State:     stateName(st.state),
		StartedAt: startedAt(boot, st.startTicks),
		CPUTime:   (st.utime + st.stime) * tickNS,
		Memory:    domain.Bytes(float64(st.rss) * float64(s.pageSize)),
		Threads:   domain.Count(float64(st.threads)),
		Priority:  domain.Count(float64(st.nice)),
	}
}
