//go:build linux

package collector

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// Detail reads one process in full. A PID that was reused for another run
// returns ErrGone, so a stale selection never shows another process's data.
func (s *procSource) Detail(ctx context.Context, id domain.ProcessIdentity) (domain.ProcessDetail, error) {
	if err := ctx.Err(); err != nil {
		return domain.ProcessDetail{}, err
	}

	base := "/proc/" + strconv.Itoa(int(id.PID))
	data, err := os.ReadFile(base + "/stat") //nolint:gosec // reason: fixed /proc/<pid>/stat path built from a numeric PID.
	if err != nil {
		return domain.ProcessDetail{}, domain.ErrGone
	}

	st, err := parseProcStat(data)
	if err != nil {
		return domain.ProcessDetail{}, domain.ErrGone
	}
	if st.startTicks != id.Start {
		return domain.ProcessDetail{}, domain.ErrGone
	}

	boot := time.Now().Add(-parseUptime(readFile("/proc/uptime")))
	d := domain.ProcessDetail{Process: s.rowOfStat(id.PID, st, boot)}

	d.Command = trimNull(readFile(base + "/cmdline"))
	if exe, err := os.Readlink(base + "/exe"); err == nil {
		d.Exe = exe
	}
	if cwd, err := os.Readlink(base + "/cwd"); err == nil {
		d.Cwd = cwd
	}
	if pdata := readFile("/proc/" + strconv.Itoa(int(st.ppid)) + "/stat"); pdata != nil {
		if pst, err := parseProcStat(pdata); err == nil {
			d.Parent = domain.ProcessIdentity{PID: st.ppid, Start: pst.startTicks}
			d.ParentName = pst.comm
		}
	}

	if status := readFile(base + "/status"); status != nil {
		applyStatus(&d, status)
	}
	if io := readFile(base + "/io"); io != nil {
		applyIO(&d, io)
	}

	return d, nil
}
