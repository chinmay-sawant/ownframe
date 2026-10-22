//go:build linux

package collector

import (
	"strconv"
	"strings"
)

// mount is one line of /proc/mounts.
type mount struct {
	device string
	mount  string
	fstype string
}

// parseMounts reads real filesystems only, so tmpfs and pseudo filesystems do
// not show up as disks.
func parseMounts(data []byte) []mount {
	var out []mount

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < minMountFields {
			continue
		}
		if !spaceFS(fields[2]) {
			continue
		}

		out = append(out, mount{
			device: unescapeMount(fields[0]),
			mount:  unescapeMount(fields[1]),
			fstype: fields[2],
		})
	}

	return out
}

// spaceFS reports whether a filesystem type has meaningful free space.
func spaceFS(fstype string) bool {
	switch fstype {
	case "ext2", "ext3", "ext4", "xfs", "btrfs", "f2fs", "vfat", "exfat",
		"ntfs", "ntfs3", "zfs", "fuseblk", "overlay", "reiserfs", "jfs":
		return true
	default:
		return false
	}
}

// unescapeMount decodes the octal escapes /proc/mounts uses for spaces.
func unescapeMount(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}

	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+3 < len(s) {
			if v, err := strconv.ParseUint(s[i+1:i+4], 8, 8); err == nil {
				b.WriteByte(byte(v))
				i += 3

				continue
			}
		}
		b.WriteByte(s[i])
	}

	return b.String()
}
