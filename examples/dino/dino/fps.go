package dino

import "time"

// meter counts frames per second over a half-second window.
type meter struct {
	frames int
	since  time.Time
	fps    int
}

// add records one frame at now and updates the rate when the window closes.
func (m *meter) add(now time.Time) {
	if m.since.IsZero() {
		m.since = now

		return
	}

	m.frames++

	span := now.Sub(m.since)
	if span < 500*time.Millisecond {
		return
	}

	m.fps = int(float64(m.frames)/span.Seconds() + fpsRound)
	m.frames = 0
	m.since = now
}
