package web

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

const procStatMinFields = 2

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

func runtimeRSS() (uint64, bool) {
	if runtime.GOOS != "linux" {
		return 0, false
	}
	data, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0, false
	}
	fields := strings.Fields(string(data))
	if len(fields) < procStatMinFields {
		return 0, false
	}
	pages, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0, false
	}
	return pages * uint64(max(os.Getpagesize(), 0)), true
}
