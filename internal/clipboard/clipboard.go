// Package clipboard reads and writes the desktop clipboard.
// Memory-only mode is for tests. Outside that mode the OS is the
// only source, so a failed OS read is an error and never older text.
package clipboard

import (
	"errors"
	"sync"
)

// ErrUnavailable reports a failed OS clipboard read outside
// memory-only mode. It carries no fallback text.
var ErrUnavailable = errors.New("ownframe: clipboard unavailable")

var (
	mu     sync.Mutex
	mem    string
	memory bool
)

// UseMemory forces the in-memory clipboard. Tests call this
// before any Write so they never touch the desktop clipboard.
func UseMemory(on bool) {
	mu.Lock()
	memory = on
	mu.Unlock()
}

// Write stores text in memory-only mode, and tries the OS clipboard
// otherwise. OS-mode text never lands in memory, so a later failed
// Read cannot resurface it as stale text.
func Write(text string) {
	mu.Lock()
	only := memory
	if only {
		mem = text
	}
	mu.Unlock()

	if !only {
		writeOS(text)
	}
}

// Read returns the memory text in memory-only mode, and the OS
// clipboard otherwise. An OS failure returns ErrUnavailable and no
// text; it never falls back to an older Write.
func Read() (string, error) {
	mu.Lock()
	only := memory
	mu.Unlock()

	if only {
		mu.Lock()
		defer mu.Unlock()

		return mem, nil
	}

	if text, ok := readOS(); ok {
		return text, nil
	}

	return "", ErrUnavailable
}
