package web

import "time"

// Options tunes ServeWithOptions. Perf registers /debug/pprof/* for CPU,
// heap, goroutine, and trace profiles. Debug gates /debug/state (and implies
// Perf-style inspector access). Both are off by default, so a shipped
// server exposes no developer endpoints until a developer opts in.
type Options struct {
	Perf  bool
	Debug bool
}

// maxFormBytes caps the POST form bodies the preview handlers read.
const maxFormBytes = 256 * 1024

// maxHeaderBytes caps request headers at 64 KiB (WEB-4).
const maxHeaderBytes = 1 << 16

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
)
