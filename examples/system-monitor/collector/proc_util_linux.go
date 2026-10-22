//go:build linux

package collector

import (
	"os"
	"strconv"
)

// pathExists reports whether a path is present. Probing with Stat is enough;
// the contents decide whether a value is available.
func pathExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

// parseInt32 reads a decimal PID or UID. ParseInt with bitSize 32 rejects
// out-of-range input, so the int32 conversion cannot overflow.
func parseInt32(s string) int32 {
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0
	}

	return int32(v)
}

// parseSigned reads a signed decimal field such as nice.
func parseSigned(s string) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}

	return v
}

// parseCount reads a non-negative count that fits int. ParseInt with IntSize
// rejects overflow and negatives map to zero.
func parseCount(s string) int {
	v, err := strconv.ParseInt(s, 10, strconv.IntSize)
	if err != nil || v < 0 {
		return 0
	}

	return int(v)
}

// isDigits reports whether s is a non-empty run of ASCII digits.
func isDigits(s string) bool {
	if s == "" {
		return false
	}

	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return true
}
