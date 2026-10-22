//go:build linux

package collector

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// userHZ is the tick rate /proc uses for CPU times. Linux exposes it through
// getconf, but every supported architecture reports 100 for these counters.
const userHZ = 100

// tickNS converts one /proc CPU tick to nanoseconds.
const tickNS = uint64(time.Second) / userHZ

// parseCPUStat reads the "cpu" and "cpuN" lines. Busy is total minus idle and
// iowait, so a kernel that reports extra fields still sums correctly.
func parseCPUStat(data []byte) (domain.CPUTimes, []domain.CPUTimes, error) {
	var total domain.CPUTimes

	var cores []domain.CPUTimes

	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "cpu") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < minCPUFields {
			continue
		}

		var sum, idle uint64
		for i, raw := range fields[1:] {
			v, err := strconv.ParseUint(raw, 10, 64)
			if err != nil {
				v = 0
			}
			sum += v
			if i == 3 || i == 4 {
				idle += v
			}
		}

		c := domain.CPUTimes{Total: sum * tickNS}
		if sum > idle {
			c.Busy = (sum - idle) * tickNS
		}

		if fields[0] == "cpu" {
			total = c
		} else {
			cores = append(cores, c)
		}
	}

	if total.Total == 0 {
		return total, nil, errors.New("procfs: no cpu line in /proc/stat")
	}

	return total, cores, nil
}
