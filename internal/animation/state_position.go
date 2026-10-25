package animation

import "time"

func statePosition(s *effectState, input FrameInput) (float64, int, bool, bool, bool) {
	t, rate := s.timing, s.rate
	if rate == 0 {
		rate = 1
	}
	duration := resolveDuration(t, input)
	if timeBased(t) && duration <= 0 {
		return zeroDurationPosition(t, currentOffset(s, input.Now))
	}
	raw, ok := rawPosition(t, s.start, input)
	if !ok {
		return 0, 0, false, false, false
	}
	switch {
	case s.hasHold:
		raw = s.holdRaw
	case timeBased(t):
		elapsed := float64(s.offset)
		if !s.paused {
			elapsed += float64(input.Now.Sub(s.start)) * rate
		}
		raw = (elapsed - float64(t.Delay)) / float64(duration)
		if s.paused {
			// The first paused sample freezes the exact position it reached,
			// delay included, so a later Clock handoff carries the phase
			// rather than recomputing it from the clock origin.
			s.holdRaw, s.hasHold = raw, true
		}
	case rate < 0:
		raw = t.Iterations - raw
	}
	s.lastRaw, s.hasRaw = raw, true
	progress, cycle, live, apply, available := samplePosition(t, raw)
	if rate < 0 {
		live = raw > 0
	}
	return progress, cycle, live, apply, available
}

// resolveDuration returns the cycle length the position code divides by.
// animation-duration:auto has no intrinsic time for a document-clock
// animation, so it resolves to zero and the animation applies its end value at
// once. A timeline-driven animation keeps its declared duration; resolving auto
// to the timeline's range belongs to the timeline phase.
func resolveDuration(t Timing, _ FrameInput) time.Duration {
	if !timeBased(t) {
		return t.Duration
	}
	if t.DurationAuto || t.Duration < 0 {
		return 0
	}
	return t.Duration
}

func currentOffset(s *effectState, now time.Time) time.Duration {
	offset := s.offset
	if !s.paused && s.started {
		offset += time.Duration(float64(now.Sub(s.start)) * s.rate)
	}
	return offset
}

// frozenPosition returns the position a pause freezes a time-based effect at,
// with the signed start delay folded in and a reverse rate applied. It reads
// the active time the state already carries, so the caller must have updated
// that first. A timeline effect has no document-clock position to freeze and
// reports false.
func frozenPosition(s *effectState) (float64, bool) {
	t := s.timing
	if !timeBased(t) {
		return 0, false
	}
	duration := resolveDuration(t, FrameInput{})
	if duration <= 0 {
		return 0, false
	}
	raw := (float64(s.offset) - float64(t.Delay)) / float64(duration)
	if s.rate < 0 {
		raw = t.Iterations - raw
	}

	return raw, true
}
