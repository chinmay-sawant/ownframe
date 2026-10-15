// Package crash writes a panic report to a text file on this machine.
// The file is not uploaded.
package crash

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync"
)

var (
	mu     sync.Mutex
	custom string
)

// Write creates the folder and writes one UTF-8 report.
// The name starts with a UTC timestamp, and a later call in that
// same second gets a different name.
func Write(title, reason string) (string, error) {
	folder := location()

	if err := os.MkdirAll(folder, 0o700); err != nil {
		return "", err
	}
	if err := os.Chmod(folder, 0o700); err != nil {
		return "", err
	}

	body := fmt.Sprintf(
		"title: %s\nreason: %s\ngo: %s\n%s",
		title,
		reason,
		runtime.Version(),
		debug.Stack(),
	)

	return create(folder, body)
}

func location() string {
	mu.Lock()
	dir := custom
	mu.Unlock()

	if dir != "" {
		return dir
	}

	config, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "ownframe-crashes")
	}

	return filepath.Join(config, "ownframe", "crashes")
}
