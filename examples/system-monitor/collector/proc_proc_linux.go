//go:build linux

package collector

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// Processes lists /proc. A PID that exits between the directory read and the
// stat read is skipped; access to one process never fails the whole table.
func (s *procSource) Processes(ctx context.Context) ([]domain.Process, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	boot := time.Now().Add(-parseUptime(readFile("/proc/uptime")))
	rows := make([]domain.Process, 0, len(entries))

	for i, entry := range entries {
		if i%256 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}

		pid64, err := strconv.ParseInt(entry.Name(), 10, 32)
		if err != nil || pid64 <= 0 || !entry.IsDir() {
			continue
		}

		data, err := os.ReadFile("/proc/" + entry.Name() + "/stat")
		if err != nil {
			continue
		}

		st, err := parseProcStat(data)
		if err != nil {
			continue
		}

		rows = append(rows, s.rowOfStat(int32(pid64), st, boot))
	}

	return rows, nil
}

// countProcs counts the numeric entries in /proc, which is the running
// process count.
func countProcs() (int, bool) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0, false
	}

	n := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := strconv.ParseUint(entry.Name(), 10, 32); err == nil {
			n++
		}
	}

	return n, true
}

// readFile returns a file's trimmed contents, or nil when it is unreadable.
func readFile(path string) []byte {
	data, err := os.ReadFile(path) //nolint:gosec // reason: every caller passes a fixed /proc or /sys path, some with a numeric PID segment.
	if err != nil {
		return nil
	}

	return data
}

// trimNull joins NUL-separated /proc fields into one line.
func trimNull(data []byte) string {
	return strings.TrimSpace(strings.ReplaceAll(string(data), "\x00", " "))
}
