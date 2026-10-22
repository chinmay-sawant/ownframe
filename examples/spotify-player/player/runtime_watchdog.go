package player

import "time"

type watchdog struct {
	Frames int
	Last   int
	LastAt time.Time
	Window time.Duration
	Limit  int
	Over   bool
}

func (w *watchdog) Sample(now time.Time, window time.Duration, limit int) bool {
	if w == nil {
		return false
	}
	if window <= 0 {
		window = time.Second
	}
	if w.LastAt.IsZero() {
		w.LastAt, w.Last, w.Frames = now, 0, 1
		w.Window, w.Limit, w.Over = window, limit, false
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
	w.LastAt, w.Last, w.Frames = now, 0, 1
	w.Window, w.Limit, w.Over = window, limit, false
	return false
}

func (w *watchdog) Reset() {
	if w == nil {
		return
	}
	w.Frames, w.Last, w.LastAt = 0, 0, time.Time{}
	w.Over = false
}

func (w *watchdog) Count() int {
	if w == nil {
		return 0
	}
	return w.Frames
}
