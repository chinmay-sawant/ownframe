//go:build linux

package collector

import (
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// kibibyte scales the KiB counters /proc reports up to bytes.
const kibibyte = 1024

// Field counts guard the /proc line shapes before indexing.
const (
	minCPUFields    = 5
	minDiskFields   = 10
	minMountFields  = 3
	minStatFields   = 22
	minPasswdFields = 3
	// sectorBytes is the 512-byte sector unit diskstats counters use.
	sectorBytes = 512
)

// parseLoadAvg reads the three load averages.
func parseLoadAvg(data []byte) [3]domain.Value {
	var out [3]domain.Value

	fields := strings.Fields(string(data))
	for i := 0; i < 3 && i < len(fields); i++ {
		v, err := strconv.ParseFloat(fields[i], 64)
		if err == nil {
			out[i] = domain.Count(v)
		}
	}

	return out
}

// parseUptime reads the seconds field of /proc/uptime.
func parseUptime(data []byte) time.Duration {
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0
	}

	v, err := strconv.ParseFloat(fields[0], 64)
	if err != nil || v < 0 {
		return 0
	}

	return time.Duration(v * float64(time.Second))
}
