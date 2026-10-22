//go:build linux

package collector

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// parseMemInfo reads the kB values from /proc/meminfo.
func parseMemInfo(data []byte) domain.Memory {
	vals := make(map[string]uint64)

	for _, line := range strings.Split(string(data), "\n") {
		key, rest, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}

		v, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			continue
		}
		vals[key] = v * kibibyte
	}

	m := domain.Memory{}
	if v, ok := vals["MemTotal"]; ok {
		m.Total = domain.Bytes(float64(v))
	}
	if v, ok := vals["MemAvailable"]; ok {
		m.Available = domain.Bytes(float64(v))
	}
	if v, ok := vals["MemFree"]; ok {
		m.Free = domain.Bytes(float64(v))
	}
	if v, ok := vals["Cached"]; ok {
		m.Cached = domain.Bytes(float64(v))
	}
	if total, ok := m.Total.Get(); ok {
		if avail, ok := m.Available.Get(); ok {
			m.Used = domain.Bytes(total - avail)
		}
	}
	if v, ok := vals["SwapTotal"]; ok {
		m.SwapTotal = domain.Bytes(float64(v))
	}
	if free, ok := vals["SwapFree"]; ok {
		if total, ok := m.SwapTotal.Get(); ok && total >= float64(free) {
			m.SwapUsed = domain.Bytes(total - float64(free))
		}
	}

	return domain.DeriveMemory(m)
}
