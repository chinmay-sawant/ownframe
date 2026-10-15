package eyes

import "time"

type motionInput struct {
	x, y  float64
	tilt  bool
	shake bool
}

// SetMotion queues normalized screen-aligned tilt for the next frame.
func (a *App) SetMotion(x, y float64) {
	a.motionMu.Lock()
	a.motion.x, a.motion.y, a.motion.tilt = x, y, true
	a.motionMu.Unlock()
}

// Shake queues a brief spiral pulse on the frame loop.
func (a *App) Shake() {
	a.motionMu.Lock()
	a.motion.shake = true
	a.motionMu.Unlock()
}

// applyMotion transfers sensor callbacks onto the frame loop before paint.
func (a *App) applyMotion(now time.Time) {
	a.motionMu.Lock()
	m := a.motion
	a.motion.tilt, a.motion.shake = false, false
	a.motionMu.Unlock()
	if m.tilt {
		a.st.tiltX, a.st.tiltY = m.x, m.y
		a.st.tiltUntil = now.Add(1500 * time.Millisecond)
	}
	if m.shake {
		s := &a.st
		if s.mood != 3 {
			s.mood, s.moodT, s.spiral = 3, now, 0
			s.blinkOn, s.blinkP = false, 0
			s.tLookX, s.tLookY = 0, 0
		}
		s.spiralUntil = now.Add(650 * time.Millisecond)
	}
}
