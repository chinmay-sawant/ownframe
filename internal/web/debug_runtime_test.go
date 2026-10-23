package web

import (
	"runtime"
	"testing"
)

func TestRuntimeSnapshotRSS(t *testing.T) {
	snapshot := runtimeSnapshot()
	rss, ok := snapshot["rss_bytes"].(uint64)
	if runtime.GOOS == "linux" && (!ok || rss == 0) {
		t.Fatal("Linux runtime snapshot lacks resident memory")
	}
	if runtime.GOOS != "linux" && ok {
		t.Fatal("unsupported RSS must be absent rather than a zero sample")
	}
}
