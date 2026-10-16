package render

import "sync"

// The pinned engine replaces shared document readers during Parse and Apply.
// Guard registration and reads together until the engine publishes them once.
// The lock is load-bearing, not a tuning knob: the readers are process-wide
// engine state, so parsing or applying outside this lock would race. Perf
// work belongs in fewer parse calls (the render cache), not in narrowing
// this mutex.
var engineMu sync.Mutex
