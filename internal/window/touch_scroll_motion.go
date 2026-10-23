package window

import (
	"math"
	"time"
)

// touchScrollMotion bounds: sensitivity range, release keep factor, decay range.
const (
	touchSensMin     = .25
	touchSensMax     = 3
	touchReleaseKeep = .5
	touchDecayMin    = 1
	touchDecayMax    = 12
)

type touchScrollMotion struct {
	sensitivity, decay  float64
	vx, vy, rx, ry      float64
	lastTouch, lastStep time.Time
	dragging            bool
}

func newTouchScrollMotion(sensitivity, decay float64) touchScrollMotion {
	if sensitivity <= 0 || math.IsNaN(sensitivity) || math.IsInf(sensitivity, 0) {
		sensitivity = 1
	}
	if decay <= 0 || math.IsNaN(decay) || math.IsInf(decay, 0) {
		decay = 5
	}
	return touchScrollMotion{sensitivity: clampFloat(sensitivity, touchSensMin, touchSensMax), decay: clampFloat(decay, touchDecayMin, touchDecayMax)}
}

func (m *touchScrollMotion) defaults() {
	if m.sensitivity <= 0 {
		*m = newTouchScrollMotion(m.sensitivity, m.decay)
	}
}

func (m *touchScrollMotion) begin(now time.Time) {
	m.defaults()
	m.vx, m.vy, m.rx, m.ry = 0, 0, 0, 0
	m.lastTouch, m.lastStep, m.dragging = now, now, true
}

func (m *touchScrollMotion) configure(sensitivity, decay float64) {
	next := newTouchScrollMotion(sensitivity, decay)
	m.sensitivity, m.decay = next.sensitivity, next.decay
}

func (m *touchScrollMotion) release() {
	m.vx, m.vy = m.vx*touchReleaseKeep, m.vy*touchReleaseKeep
	m.dragging = false
}

func (m *touchScrollMotion) idle(now time.Time) {
	if m.dragging && now.Sub(m.lastTouch) > 45*time.Millisecond {
		m.vx, m.vy = 0, 0
	}
}

func (m *touchScrollMotion) move(dx, dy int, now time.Time) (int, int) {
	m.defaults()
	dt := now.Sub(m.lastTouch).Seconds()
	if dt > 0 && dt < .2 {
		vx, vy := -float64(dx)*m.sensitivity/dt, -float64(dy)*m.sensitivity/dt
		m.vx = scrollVelocity(m.vx, vx)
		m.vy = scrollVelocity(m.vy, vy)
	} else {
		m.vx, m.vy = 0, 0
	}
	m.lastTouch, m.lastStep = now, now
	x, rx := scrollPixels(-float64(dx)*m.sensitivity + m.rx)
	y, ry := scrollPixels(-float64(dy)*m.sensitivity + m.ry)
	m.rx, m.ry = rx, ry
	return x, y
}

func (m *touchScrollMotion) cancel() { m.vx, m.vy, m.dragging = 0, 0, false }
