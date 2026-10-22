package store

import (
	"net/url"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// dsn builds the driver name. The driver takes a file: URI, so spaces,
// Unicode, and punctuation in the data directory are percent-encoded. On
// Windows a drive path becomes file:///C:/..., the form SQLite parses.
func dsn(dir string, memory bool) (string, error) {
	if memory {
		return ":memory:", nil
	}

	abs, err := filepath.Abs(filepath.Join(dir, "jobs.sqlite"))
	if err != nil {
		return "", err
	}

	slash := filepath.ToSlash(abs)
	if runtime.GOOS == "windows" && !strings.HasPrefix(slash, "/") {
		slash = "/" + slash
	}

	u := url.URL{Scheme: "file", Path: slash}

	values := url.Values{}
	values.Add("_pragma", "busy_timeout("+strconv.FormatInt(busyMillis(), 10)+")")
	values.Add("_pragma", "foreign_keys(1)")
	u.RawQuery = values.Encode()

	return u.String(), nil
}

// busyMillis is the connection lock wait.
func busyMillis() int64 { return busyTimeout.Milliseconds() }
