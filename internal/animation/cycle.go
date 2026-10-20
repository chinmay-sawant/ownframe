package animation

import (
	"math"
	"time"
)

func samplePosition(t Timing, raw float64) (float64, int, bool, bool, bool) {
	if t.Duration <= 0 && (t.Timeline == "" || t.Timeline == "auto") {
		return 0, 0, false, false, true
	}
	if raw < 0 {
		return directed(0, 0, t.Direction), 0, true, t.Fill == FillBackwards || t.Fill == FillBoth, true
	}
	if t.Infinite {
		cycle := math.Floor(raw)
		return directed(raw-cycle, int(cycle), t.Direction), int(cycle), true, true, true
	}
	if t.Iterations <= 0 {
		return 0, 0, false, false, true
	}
	if raw >= t.Iterations {
		cycle := int(math.Ceil(t.Iterations) - 1)
		last := t.Iterations - float64(cycle)
		if last == 0 {
			last = 1
		}
		end := t.Iterations + float64(t.EndDelay)/float64(maxDuration(t.Duration))
		return directed(last, cycle, t.Direction), cycle, raw < end,
			t.Fill == FillForwards || t.Fill == FillBoth, true
	}
	cycle := int(math.Floor(raw))
	return directed(raw-float64(cycle), cycle, t.Direction), cycle, true, true, true
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

func maxDuration(d time.Duration) time.Duration {
	if d <= 0 {
		return time.Second
	}
	return d
}
