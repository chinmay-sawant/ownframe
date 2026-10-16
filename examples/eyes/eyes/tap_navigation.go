package eyes

import (
	"time"
)

const doubleTapWindow = 260 * time.Millisecond

func (a *App) onSideTap(direction int, now time.Time) {
	s := &a.st
	if s.lastTapDir == direction && now.Sub(s.lastTapAt) <= doubleTapWindow {
		s.lastTapAt = now
		return
	}
	a.stepMood(direction, now)
	s.lastTapDir, s.lastTapAt = direction, now
}

func (a *App) stepMood(direction int, now time.Time) {
	index := (a.st.mood + direction + len(poses)) % len(poses)
	a.selectMood(index, now)
}
