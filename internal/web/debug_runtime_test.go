package web

import (
	"runtime"
	"testing"
)

func TestRuntimeSnapshotRSS(t *testing.T) {
	snapshot := runtimeSnapshot()
	rss, ok := snapshot["rss_bytes"].(uint64)
	supported := runtime.GOOS == "linux" || runtime.GOOS == "windows"
	if supported && (!ok || rss == 0) {
		t.Fatal("supported runtime snapshot lacks resident memory")
	}
	if !supported && ok {
		t.Fatal("unsupported RSS must be absent rather than a zero sample")
	}
}
