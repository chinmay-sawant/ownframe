package web

// Options tunes ServeWithOptions. Perf registers /debug/pprof/* for CPU,
// heap, goroutine, and trace profiles. Debug gates /debug/state (and implies
// Perf-style inspector access). Both are off by default, so a shipped
// server exposes no developer endpoints until a developer opts in.
type Options struct {
	Perf  bool
	Debug bool
}
