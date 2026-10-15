package web

import (
	"log"
	"sync"
)

// runtimeSnapshotMu guards RuntimeSnapshot. Direct assignment still works, but
// new code should call SetRuntimeSnapshot so replacement cannot race with a
// request.
var runtimeSnapshotMu sync.RWMutex

// RuntimeSnapshot reports window runtime stats for /debug/state. The window
// package sets it; it stays nil in tests and headless builds.
var RuntimeSnapshot func() map[string]any

// SetRuntimeSnapshot installs fn as the runtime stats hook used by /debug/state.
// Passing nil clears the hook. It is safe to call from any goroutine.
func SetRuntimeSnapshot(fn func() map[string]any) {
	runtimeSnapshotMu.Lock()
	defer runtimeSnapshotMu.Unlock()

	RuntimeSnapshot = fn
}

// GetRuntimeSnapshot returns the current hook without invoking it.
func GetRuntimeSnapshot() func() map[string]any {
	runtimeSnapshotMu.RLock()
	defer runtimeSnapshotMu.RUnlock()

	return RuntimeSnapshot
}

// runtimeSnapshot safely invokes the hook. A panic in the hook is recovered so
// one bad sampler cannot abort the debug request.
func runtimeSnapshot() map[string]any {
	runtimeSnapshotMu.RLock()
	fn := RuntimeSnapshot
	runtimeSnapshotMu.RUnlock()

	if fn == nil {
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("runtime snapshot panic: %v", r)
		}
	}()

	return fn()
}
