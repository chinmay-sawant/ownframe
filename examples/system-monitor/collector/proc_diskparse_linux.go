//go:build linux

package collector

import "strings"

// parseDiskStats reads sector counters, in 512 byte sectors, for whole disks.
func parseDiskStats(data []byte) map[string]diskIO {
	out := make(map[string]diskIO)

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < minDiskFields {
			continue
		}

		name := wholeDiskName(fields[2])
		if name == "" {
			continue
		}

		out[name] = diskIO{
			read:  parseUint(fields[5]) * sectorBytes,
			write: parseUint(fields[9]) * sectorBytes,
		}
	}

	return out
}

// wholeDiskName maps a device or partition name to its whole disk. It returns
// an empty string for devices the monitor should not graph, such as loop and
// ram devices.
func wholeDiskName(name string) string {
	switch {
	case strings.HasPrefix(name, "nvme") || strings.HasPrefix(name, "mmcblk"):
		if i := strings.LastIndex(name, "p"); i > 0 && isDigits(name[i+1:]) {
			name = name[:i]
		}
	case strings.HasPrefix(name, "sd"), strings.HasPrefix(name, "vd"),
		strings.HasPrefix(name, "hd"), strings.HasPrefix(name, "xvd"):
		name = strings.TrimRight(name, "0123456789")
	default:
		return ""
	}

	if name == "" {
		return ""
	}

	return name
}
