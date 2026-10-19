package animation

import (
	"math"
	"time"
)

func (p *Player) position(now time.Time) (float64, bool, bool) {
	t := p.timing
	raw := float64(now.Sub(p.start)-t.Delay) / float64(t.Duration)
	if raw < 0 {
		apply := t.Fill == FillBackwards || t.Fill == FillBoth
		return directed(0, 0, t.Direction), true, apply
	}
	if t.Infinite {
		cycle := math.Floor(raw)
		return directed(raw-cycle, int(cycle), t.Direction), true, true
	}
	if t.Iterations <= 0 {
		return 0, false, false
	}
	if raw >= t.Iterations {
		last := math.Mod(t.Iterations, 1)
		if last == 0 {
			last = 1
		}
		if t.Fill == FillForwards || t.Fill == FillBoth {
			end := t.Iterations + float64(t.EndDelay)/float64(t.Duration)
			return directed(last, int(math.Ceil(t.Iterations)-1), t.Direction), raw < end, true
		}
		return 0, false, false
	}
	cycle := math.Floor(raw)
	return directed(raw-cycle, int(cycle), t.Direction), true, true
}

func directed(progress float64, cycle int, direction Direction) float64 {
	reverse := direction == Reverse || direction == AlternateReverse
	if (direction == Alternate || direction == AlternateReverse) && cycle%2 != 0 {
		reverse = !reverse
	}
	if reverse {
		return 1 - progress
	}
	return progress
}
