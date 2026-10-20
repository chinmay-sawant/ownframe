package animation

import "time"

type effectState struct {
	timing           Timing
	start, pausedAt  time.Time
	offset           time.Duration
	rate             float64
	holdRaw, lastRaw float64
	started, enabled bool
	paused, hasHold  bool
	hasRaw           bool
	finished         bool
}
