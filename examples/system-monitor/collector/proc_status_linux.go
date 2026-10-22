//go:build linux

package collector

import (
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// applyStatus fills the values that only /proc/<pid>/status carries.
func applyStatus(d *domain.ProcessDetail, data []byte) {
	for _, line := range strings.Split(string(data), "\n") {
		key, rest, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}

		switch key {
		case "VmSize":
			d.Virtual = domain.Bytes(float64(parseUint(fields[0])) * kibibyte)
		case "VmRSS":
			d.Memory = domain.Bytes(float64(parseUint(fields[0])) * kibibyte)
		case "Threads":
			d.Threads = domain.Count(float64(parseUint(fields[0])))
		case "Uid":
			if name, ok := userNames()[int32(parseUint(fields[0]))]; ok { //nolint:gosec // reason: UIDs are 32-bit values, which fit int32.
				d.User = name
			}
		}
	}
}

// applyIO fills cumulative read and write bytes from /proc/<pid>/io.
func applyIO(d *domain.ProcessDetail, data []byte) {
	for _, line := range strings.Split(string(data), "\n") {
		key, rest, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}

		switch key {
		case "read_bytes":
			d.ReadBytes = domain.Bytes(float64(parseUint(fields[0])))
		case "write_bytes":
			d.WriteBytes = domain.Bytes(float64(parseUint(fields[0])))
		}
	}
}

// userNames maps UIDs to names from /etc/passwd, read once.
var userNames = sync.OnceValue(func() map[int32]string {
	out := make(map[int32]string)

	data, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return out
	}

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Split(line, ":")
		if len(fields) < minPasswdFields {
			continue
		}

		uid, err := strconv.ParseInt(fields[2], 10, 32)
		if err != nil {
			continue
		}
		out[int32(uid)] = fields[0]
	}

	return out
})
