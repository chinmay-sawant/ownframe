//go:build linux

package web

import (
	"os"
	"strconv"
	"strings"
)

const procStatMinFields = 2

func runtimeRSS() (uint64, bool) {
	data, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0, false
	}
	fields := strings.Fields(string(data))
	if len(fields) < procStatMinFields {
		return 0, false
	}
	pages, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0, false
	}
	return pages * uint64(max(os.Getpagesize(), 0)), true
}
