package web

import (
	"runtime"
)

// runtimeSnapshot samples only an explicitly opted-in server's debug request.
// Each server owns its perf flag; no global callback can enable another server.
func runtimeSnapshot() map[string]any {
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	snapshot := map[string]any{
		"heap_alloc":  memory.HeapAlloc,
		"heap_sys":    memory.HeapSys,
		"total_alloc": memory.TotalAlloc,
		"goroutines":  runtime.NumGoroutine(),
		"num_gc":      memory.NumGC,
	}
	if rss, ok := runtimeRSS(); ok {
		snapshot["rss_bytes"] = rss
	}
	return snapshot
}

// runtimeRSS reports resident bytes and whether the platform supplied
// one. It lives in the OS-specific files next to this one.
