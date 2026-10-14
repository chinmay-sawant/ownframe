package runtimebehavior

import "time"

// Watchdog samples update frequency to catch feedback loops.
type Watchdog struct {
	Frames int
	Last   int
	LastAt time.Time
	Window time.Duration
	Limit  int
	Over   bool
}

// Tick increments frames and reports whether frequency exceeded limit.
func (w *Watchdog) Tick(limit int) bool {
	if w == nil {
		return false
	}
	w.Frames++
	if limit <= 0 {
		w.Over = false
		return false
	}
	over := w.Frames-w.Last > limit
	if over {
		w.Last = w.Frames
		w.Over = true
	} else {
		w.Over = false
	}
	return over
}

// Sample records a frame at now and reports a feedback loop when the
// observed frame count exceeds limit within the sampling window.
func (w *Watchdog) Sample(now time.Time, window time.Duration, limit int) bool {
	if w == nil {
		return false
	}
	if window <= 0 {
		window = time.Second
	}
	if w.LastAt.IsZero() {
		w.LastAt = now
		w.Last = 0
		w.Frames = 1
		w.Window = window
		w.Limit = limit
		w.Over = false
		return false
	}
	if now.Before(w.LastAt) {
		now = w.LastAt
	}
	if now.Sub(w.LastAt) < window {
		w.Frames++
		if limit > 0 && w.Frames-w.Last > limit {
			w.Over = true
			return true
		}
		w.Over = false
		return false
	}
	w.LastAt = now
	w.Last = 0
	w.Frames = 1
	w.Window = window
	w.Limit = limit
	w.Over = false
	return false
}

// Frequency reports frames per second for the last sample window.
func (w *Watchdog) Frequency(window time.Duration) float64 {
	if w == nil || window <= 0 {
		return 0
	}
	return float64(w.Frames) / window.Seconds()
}

// Reset clears watchdog state.
func (w *Watchdog) Reset() {
	if w == nil {
		return
	}
	w.Frames = 0
	w.Last = 0
	w.LastAt = time.Time{}
	w.Over = false
}

// IsOver reports whether the last sample exceeded the limit.
func (w *Watchdog) IsOver() bool {
	if w == nil {
		return false
	}
	return w.Over
}

// Count returns the frame count for the current sample.
func (w *Watchdog) Count() int {
	if w == nil {
		return 0
	}
	return w.Frames
}
