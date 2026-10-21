package window

import "time"

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
