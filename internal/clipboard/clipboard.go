// Package clipboard reads and writes the desktop clipboard.
// Memory-only mode is for tests. Outside that mode the OS is the
// only source, so a failed OS read is an error and never older text.
package clipboard

import (
	"errors"
	"fmt"
	"sync"
)

// ErrUnavailable reports a failed OS clipboard read outside
// memory-only mode. It carries no fallback text.
var ErrUnavailable = errors.New("ownframe: clipboard unavailable")

// ErrWaylandToolsMissing reports a Wayland session with no DISPLAY fallback
// where wl-copy and/or wl-paste are absent from PATH. When Read encounters
// this condition it wraps both ErrUnavailable and ErrWaylandToolsMissing so
// errors.Is(err, ErrUnavailable) remains true.
var ErrWaylandToolsMissing = errors.New("wayland clipboard requires wl-copy and wl-paste on PATH")

// Status describes the active clipboard backend and whether it is usable.
// Calling Diagnose never touches the OS clipboard or mutates memory mode.
type Status struct {
	Backend   string
	Available bool
	Reason    string
}

var (
	mu     sync.Mutex
	mem    string
	memory bool
)

// Diagnose reports which clipboard backend handles Read and Write in the
// current process environment without touching the OS clipboard.
func Diagnose() Status {
	mu.Lock()
	only := memory
	mu.Unlock()
	if only {
		return Status{Backend: "memory", Available: true}
	}

	return diagnoseOS()
}

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
// text; it never falls back to an older Write. The mode and the memory
// value are read under one lock, so a concurrent UseMemory cannot mix
// them. An OS read that races a mode flip returns the memory text.
func Read() (string, error) {
	mu.Lock()
	if memory {
		text := mem
		mu.Unlock()

		return text, nil
	}
	mu.Unlock()

	text, ok := readOS()
	if !ok {
		return "", unavailableErr()
	}

	mu.Lock()
	if memory {
		text = mem
	}
	mu.Unlock()

	return text, nil
}

// Clear empties the clipboard: the memory copy and the OS selection.
// Tests call UseMemory(true) first so Clear never touches the desktop.
func Clear() {
	mu.Lock()
	mem = ""
	only := memory
	mu.Unlock()

	if !only {
		clearOS()
	}
}

func unavailableErr() error {
	st := diagnoseOS()
	if !st.Available && st.Reason == ErrWaylandToolsMissing.Error() {
		return fmt.Errorf("%w: %w", ErrUnavailable, ErrWaylandToolsMissing)
	}

	return ErrUnavailable
}
