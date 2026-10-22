package web

import (
	"fmt"
	"log"
	"sync"
)

// runtimeSnapshotMu guards runtimeSnapshotFn. All access goes through
// SetRuntimeSnapshot/GetRuntimeSnapshot so replacement cannot race with
// a request.
var runtimeSnapshotMu sync.RWMutex

// runtimeSnapshotFn reports window runtime stats for /debug/state. The
// window package sets it; it stays nil in tests and headless builds.
var runtimeSnapshotFn func() map[string]any

// SetRuntimeSnapshot installs fn as the runtime stats hook used by
// /debug/state. Passing nil clears the hook. Safe for any goroutine.
func SetRuntimeSnapshot(fn func() map[string]any) {
	runtimeSnapshotMu.Lock()
	defer runtimeSnapshotMu.Unlock()

	runtimeSnapshotFn = fn
}

// GetRuntimeSnapshot returns the current hook without invoking it.
func GetRuntimeSnapshot() func() map[string]any {
	runtimeSnapshotMu.RLock()
	defer runtimeSnapshotMu.RUnlock()

	return runtimeSnapshotFn
}

// runtimeSnapshot safely invokes the hook. A panic in the hook is
// recovered so one bad sampler cannot abort the debug request.
func runtimeSnapshot() map[string]any {
	runtimeSnapshotMu.RLock()
	fn := runtimeSnapshotFn
	runtimeSnapshotMu.RUnlock()

	if fn == nil {
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("runtime snapshot panic: %s", sanitizeLine(fmt.Sprint(r)))
		}
	}()

	return fn()
}
