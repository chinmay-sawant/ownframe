package window

import (
	"math"
	"time"
)

const scrollDtMax = .05 // seconds, max frame dt

const scrollVelCap = 5000 // px/s, velocity cap

const scrollStopSpeed = 24 // px/s, stop threshold

const scrollBlendOld = .65 // weight for old velocity

const scrollBlendNew = .35 // weight for new sample

const doubleClickCount = 2

const tripleClickCount = 3

func (m *touchScrollMotion) step(now time.Time) (int, int) {
	if m.dragging || m.sensitivity <= 0 {
		return 0, 0
	}
	dt := min(now.Sub(m.lastStep).Seconds(), scrollDtMax)
	m.lastStep = now
	if dt <= 0 {
		return 0, 0
	}
	keep := math.Exp(-m.decay * dt)
	x, rx := scrollPixels(m.vx*(1-keep)/m.decay + m.rx)
	y, ry := scrollPixels(m.vy*(1-keep)/m.decay + m.ry)
	m.rx, m.ry = rx, ry
	m.vx, m.vy = m.vx*keep, m.vy*keep
	if math.Abs(m.vx)+math.Abs(m.vy) < scrollStopSpeed {
		m.vx, m.vy = 0, 0
	}
	return x, y
}

func scrollVelocity(old, sample float64) float64 {
	if old == 0 {
		return clampFloat(sample, -scrollVelCap, scrollVelCap)
	}
	return clampFloat(old*scrollBlendOld+sample*scrollBlendNew, -scrollVelCap, scrollVelCap)
}

func scrollPixels(v float64) (int, float64) {
	pixels := int(v)
	return pixels, v - float64(pixels)
}

func clampFloat(v, low, high float64) float64 { return max(low, min(v, high)) }
