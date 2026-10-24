package animation

import (
	"math"
	"time"
)

func samplePosition(t Timing, raw float64) (float64, int, bool, bool, bool) {
	if raw < 0 {
		return directed(0, 0, t.Direction), 0, true, t.Fill == FillBackwards || t.Fill == FillBoth, true
	}
	if t.Infinite {
		cycle := math.Floor(raw)
		return directed(raw-cycle, int(cycle), t.Direction), int(cycle), true, true, true
	}
	if t.Iterations <= 0 {
		// Iteration count zero never reaches the active interval; forwards
		// fill still applies the value the animation would have ended on.
		return directed(1, 0, t.Direction), 0, false,
			t.Fill == FillForwards || t.Fill == FillBoth, true
	}
	if raw >= t.Iterations {
		// A fractional iteration count ends inside the last cycle, so last
		// is the fractional tail rather than an implicit whole keyframe.
		cycle := int(math.Ceil(t.Iterations) - 1)
		last := t.Iterations - float64(cycle)
		if last == 0 {
			last = 1
		}
		end := t.Iterations + float64(t.EndDelay)/float64(maxDuration(t.Duration))
		if raw < end {
			// End delay holds the last sampled value for the declared delay,
			// whatever the fill mode decides after it.
			return directed(last, cycle, t.Direction), cycle, true, true, true
		}
		return directed(last, cycle, t.Direction), cycle, false,
			t.Fill == FillForwards || t.Fill == FillBoth, true
	}
	cycle := int(math.Floor(raw))
	return directed(raw-float64(cycle), cycle, t.Direction), cycle, true, true, true
}

// zeroDurationPosition samples a zero-length time-based animation. There is no
// active interval to divide, so the sample is the end value once the start
// delay has passed; the end delay holds it, and the fill mode decides whether
// it survives afterwards. Such an animation has no time left to play, so it is
// live only while an end delay is holding it.
func zeroDurationPosition(t Timing, elapsed time.Duration) (float64, int, bool, bool, bool) {
	if elapsed < t.Delay {
		return directed(0, 0, t.Direction), 0, false,
			t.Fill == FillBackwards || t.Fill == FillBoth, true
	}
	value := directed(1, 0, t.Direction)
	if elapsed < t.Delay+t.EndDelay {
		return value, 0, true, true, true
	}

	return value, 0, false,
		t.Fill == FillForwards || t.Fill == FillBoth, true
}

// timeBased reports whether a timing uses the document clock rather than a
// declared timeline.
func timeBased(t Timing) bool { return t.Timeline == "" || t.Timeline == "auto" }

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
