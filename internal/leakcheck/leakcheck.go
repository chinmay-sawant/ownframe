// Package leakcheck fails a test binary that leaks goroutines (GEN-26).
// It snapshots the count before m.Run, then polls after: leaked workers
// become a slow DoS in long-lived desktop apps.
package leakcheck

import (
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	settle    = 5 * time.Second
	pollStep  = 50 * time.Millisecond
	stackSize = 1 << 20
)

// Verify runs m and reports whether goroutines drained afterwards. It
// returns the test exit code, or 1 on a leak after dumping stacks.
// Under -fuzz it just runs: the fuzz engine owns its workers after m.
func Verify(m interface{ Run() int }) int {
	if fuzzing() {
		return m.Run()
	}
	base := runtime.NumGoroutine()
	code := m.Run()
	if code != 0 {
		return code
	}
	deadline := time.Now().Add(settle)
	for runtime.NumGoroutine() > base && time.Now().Before(deadline) {
		runtime.GC()
		time.Sleep(pollStep)
	}
	if runtime.NumGoroutine() > base {
		buf := make([]byte, stackSize)
		n := runtime.Stack(buf, true)
		_, _ = os.Stderr.Write(buf[:n])
		return 1
	}

	return 0
}

// fuzzing reports whether the binary runs under -fuzz (any -test.fuzz*
// flag, including workers). The engine owns its workers after m, so the
// leak gate stands down. os.Args is used because TestMain runs before
// the testing flags are parsed.
func fuzzing() bool {
	for _, a := range os.Args {
		if strings.HasPrefix(a, "-test.fuzz") {
			return true
		}
	}

	return false
}
