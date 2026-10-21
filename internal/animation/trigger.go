package animation

import (
	"fmt"
	"time"
)

// TriggerEffect applies an animation-trigger action to one list item.
func (p *Player) TriggerEffect(index int, action string, now time.Time) error {
	if index < 0 || index >= len(p.states) {
		return fmt.Errorf("animation effect index %d is out of range", index)
	}
	s := &p.states[index]
	phase := effectPhaseAt(s, now)
	switch action {
	case "none":
		return nil
	case "reset":
		s.timing = p.effects[index].Timing
		reset(s, now, true)
	case "pause":
		if phase == playing {
			if s.timing.Timeline != "" && s.timing.Timeline != "auto" {
				s.holdRaw, s.hasHold = s.lastRaw, s.hasRaw
			} else {
				s.offset = currentOffset(s, now)
			}
			s.start, s.pausedAt, s.paused = now, now, true
		}
	case "replay":
		s.timing = p.effects[index].Timing
		reset(s, now, false)
	case "play-once":
		if phase != playing && phase != finished {
			s.timing = p.effects[index].Timing
			s.timing.Iterations, s.timing.Infinite = 1, false
			play(s, now, phase, s.rate)
		}
	case "play", "play-forwards", "play-backwards":
		rate := s.rate
		switch action {
		case "play-forwards":
			rate = 1
		case "play-backwards":
			rate = -1
		}
		if phase == playing && action == "play" {
			return nil
		}
		if phase == initial || phase == finished {
			s.timing = p.effects[index].Timing
		}
		if phase == playing {
			s.offset = currentOffset(s, now)
			s.start, s.rate = now, rate
		} else {
			play(s, now, phase, rate)
		}
	default:
		return fmt.Errorf("unknown animation-trigger action %q", action)
	}
	return nil
}

// EffectCount reports the number of CSS animation list items.
func (p *Player) EffectCount() int {
	if p == nil {
		return 0
	}
	return len(p.effects)
}
