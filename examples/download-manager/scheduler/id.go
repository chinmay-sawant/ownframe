package scheduler

import (
	"crypto/rand"
	"encoding/hex"
)

// newID returns a stable, collision-safe job ID.
func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}

	return "job-" + hex.EncodeToString(b[:])
}

// signal wakes one waiting worker without blocking.
func (e *Engine) signal() {
	select {
	case e.wake <- struct{}{}:
	default:
	}
}
