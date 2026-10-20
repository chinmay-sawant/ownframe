package animation

import "time"

type effectPhase uint8

const (
	initial effectPhase = iota
	playing
	paused
	finished
)

func effectPhaseAt(s *effectState, now time.Time) effectPhase {
	if !s.started {
		return initial
	}
	if s.paused || !s.enabled {
		return paused
	}
	if s.finished {
		return finished
	}
	if s.timing.Timeline == "" || s.timing.Timeline == "auto" {
		_, _, live, _, ok := statePosition(s, FrameInput{Now: now})
		if ok && !live {
			return finished
		}
	}
	return playing
}

func reset(s *effectState, now time.Time, pause bool) {
	s.offset, s.start, s.pausedAt = 0, now, now
	s.started, s.enabled, s.paused, s.finished = true, true, pause, false
	s.hasHold = s.timing.Timeline != "" && s.timing.Timeline != "auto"
	s.holdRaw = -float64(s.timing.Delay) / float64(maxDuration(s.timing.Duration))
}

func play(s *effectState, now time.Time, phase effectPhase, rate float64) {
	if phase == initial || phase == finished {
		s.offset = 0
		if rate < 0 && !s.timing.Infinite {
			s.offset = s.timing.Delay + time.Duration(float64(s.timing.Duration)*s.timing.Iterations) + s.timing.EndDelay
		}
	}
	s.start, s.rate = now, rate
	s.hasHold = false
	s.started, s.enabled, s.paused, s.finished = true, true, false, false
}
