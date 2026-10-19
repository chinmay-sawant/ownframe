package csspreview

import (
	"math"
	"time"
)

func (m *motion) position(now time.Time) (float64, bool, bool) {
	raw := float64(now.Sub(m.start)-m.delay) / float64(m.duration)
	if raw < 0 {
		apply := m.kind == "transition" || m.fill == "backwards" || m.fill == "both"
		return m.directed(0, 0), true, apply
	}
	if m.kind == "transition" {
		if raw >= 1 {
			return 1, false, true
		}
		return raw, true, true
	}
	if m.infinite {
		cycle := math.Floor(raw)
		return m.directed(raw-cycle, int(cycle)), true, true
	}
	if m.iterations <= 0 {
		return 0, false, false
	}
	end := m.iterations + float64(m.endDelay)/float64(m.duration)
	if raw >= m.iterations {
		last := math.Mod(m.iterations, 1)
		if last == 0 {
			last = 1
		}
		if m.fill == "forwards" || m.fill == "both" {
			return m.directed(last, int(math.Ceil(m.iterations)-1)), raw < end, true
		}
		return 0, false, false
	}
	cycle := math.Floor(raw)
	return m.directed(raw-cycle, int(cycle)), true, true
}

func (m *motion) directed(p float64, cycle int) float64 {
	reverse := m.direction == "reverse" || m.direction == "alternate-reverse"
	if (m.direction == "alternate" || m.direction == "alternate-reverse") && cycle%2 != 0 {
		reverse = !reverse
	}
	if reverse {
		return 1 - p
	}
	return p
}
