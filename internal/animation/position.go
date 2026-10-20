package animation

import "time"

func position(t Timing, start time.Time, input FrameInput) (float64, int, bool, bool, bool) {
	raw, available := rawPosition(t, start, input)
	if !available {
		return 0, 0, false, false, false
	}
	return samplePosition(t, raw)
}

func rawPosition(t Timing, start time.Time, input FrameInput) (float64, bool) {
	raw, available := timelineProgress(t, input)
	if !available {
		return 0, false
	}
	if t.Timeline == "" || t.Timeline == "auto" {
		if t.Duration <= 0 {
			return 0, false
		}
		raw = float64(input.Now.Sub(start)-t.Delay) / float64(t.Duration)
	} else {
		lo, hi, ok := timelineRange(t, input)
		if !ok || hi <= lo {
			return 0, false
		}
		raw = centerProgress((raw-lo)/(hi-lo), t.RangeCenter, input, lo, hi)
		if t.Iterations > 0 && !t.Infinite {
			raw *= t.Iterations
		}
		if t.Duration > 0 {
			raw -= float64(t.Delay) / float64(t.Duration)
		}
	}
	return raw, true
}
