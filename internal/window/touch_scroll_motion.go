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
