//go:build !windows

package reader

import (
	"os"
	"strconv"
	"syscall"
)

// pathIdentity names the file a path currently points at, for example
// "1a2b:3c4d" from device and inode. An empty string means the platform
// cannot tell and the caller falls back to size checks.
func pathIdentity(path string) (string, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	return statIdentity(fi), nil
}

func handleIdentity(f *os.File) (string, error) {
	fi, err := f.Stat()
	if err != nil {
		return "", err
	}

	return statIdentity(fi), nil
}

func statIdentity(fi os.FileInfo) string {
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return ""
	}

	dev := strconv.FormatUint(uint64(st.Dev), 16) //nolint:unconvert // reason: Dev is int32 on darwin, conversion keeps unix builds portable
	ino := strconv.FormatUint(uint64(st.Ino), 16) //nolint:unconvert // reason: Ino is uint64 on linux but smaller on other unix, conversion keeps builds portable

	return dev + ":" + ino
}

// openShared opens path for sequential reading. Unix keeps the handle valid
// after a rename or unlink, so rotation drains through it.
func openShared(path string) (*os.File, error) { return os.Open(path) } //nolint:gosec // reason: path comes from the watched log file configuration, reading it is intended
