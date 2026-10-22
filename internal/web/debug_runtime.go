package web

import "runtime"

// runtimeSnapshot samples only an explicitly opted-in server's debug request.
// Each server owns its perf flag; no global callback can enable another server.
func runtimeSnapshot() map[string]any {
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	return map[string]any{
		"heap_alloc":  memory.HeapAlloc,
		"heap_sys":    memory.HeapSys,
		"total_alloc": memory.TotalAlloc,
		"goroutines":  runtime.NumGoroutine(),
		"num_gc":      memory.NumGC,
	}
}
