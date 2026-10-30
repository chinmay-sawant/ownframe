//go:build linux

package window

import (
	"os"
	"strconv"
	"strings"
)

// perfRSS reads resident bytes from /proc/self/statm without cgo.
// It returns 0 when the read or parse fails; the caller renders that as
// unknown, since genuine resident memory is never zero.
func perfRSS() uint64 {
	b, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0
	}
	f := strings.Fields(string(b))
	if len(f) < perfMinSamples {
		return 0
	}
	pages, err := strconv.ParseUint(f[1], 10, 64)
	if err != nil {
		return 0
	}
	return pages * uint64(max(os.Getpagesize(), 0))
}
