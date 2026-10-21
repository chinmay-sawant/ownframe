//go:build linux && !android

package clipboard

import (
	"os"
	"strconv"
	"strings"
)

func displaySpec(d string) (string, string, int, bool) {
	i := strings.LastIndex(d, ":")
	if i < 0 {
		return "", "", 0, false
	}

	host := d[:i]
	rest := d[i+1:]
	scr := "0"

	if j := strings.IndexByte(rest, '.'); j >= 0 {
		scr = rest[j+1:]
		rest = rest[:j]
	}

	if !digits(rest) || !digits(scr) {
		return "", "", 0, false
	}

	n, err := strconv.Atoi(scr)
	if err != nil {
		return "", "", 0, false
	}

	if host == "unix" {
		host = ""
	}

	return host, rest, n, true
}

func digits(s string) bool {
	if s == "" {
		return false
	}

	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

func loadAuth(host, num string) (string, []byte) {
	path := os.Getenv("XAUTHORITY")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", nil
		}

		path = home + "/.Xauthority"
	}

	b, err := os.ReadFile(path) //nolint:gosec // reason: path is the user's own XAUTHORITY file or ~/.Xauthority, local trust.
	if err != nil {
		return "", nil
	}

	if host == "" || host == "localhost" {
		if h, err := os.Hostname(); err == nil {
			host = h
		}
	}

	return pickAuth(b, host, num)
}
