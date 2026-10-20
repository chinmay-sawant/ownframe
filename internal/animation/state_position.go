package animation

import "time"

func statePosition(s *effectState, input FrameInput) (float64, int, bool, bool, bool) {
	t, rate := s.timing, s.rate
	if rate == 0 {
		rate = 1
	}
	if t.Duration <= 0 && (t.Timeline == "" || t.Timeline == "auto") {
		return position(t, s.start, input)
	}
	raw, ok := rawPosition(t, s.start, input)
	if !ok {
		return 0, 0, false, false, false
	}
	if s.hasHold {
		raw = s.holdRaw
	} else if t.Timeline == "" || t.Timeline == "auto" {
		elapsed := float64(s.offset)
		if !s.paused {
			elapsed += float64(input.Now.Sub(s.start)) * rate
		}
		raw = (elapsed - float64(t.Delay)) / float64(t.Duration)
	} else if rate < 0 {
		raw = t.Iterations - raw
	}
	s.lastRaw, s.hasRaw = raw, true
	progress, cycle, live, apply, available := samplePosition(t, raw)
	if rate < 0 {
		live = raw > 0
	}
	return progress, cycle, live, apply, available
}

func currentOffset(s *effectState, now time.Time) time.Duration {
	offset := s.offset
	if !s.paused && s.started {
		offset += time.Duration(float64(now.Sub(s.start)) * s.rate)
	}
	return offset
}
