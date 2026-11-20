package motion

import (
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"strings"
	"time"
)

func triggerActions(value, name string) []string {
	var matched []string
	var currentName string
	var current []string
	for _, token := range strings.Fields(value) {
		if isAction(token) {
			if currentName != "" {
				current = append(current, token)
			}
			continue
		}
		if currentName == name {
			matched = append(matched[:0], current...)
		}
		currentName, current = token, nil
	}
	if currentName == name {
		matched = append(matched[:0], current...)
	}
	return matched
}

func isAction(value string) bool {
	switch value {
	case "none", "play", "play-once", "play-forwards", "play-backwards", "pause", "reset", "replay":
		return true
	default:
		return false
	}
}

// DispatchTrigger applies matching trigger actions to each box's animation player.
func DispatchTrigger(boxes []layout.Box, event TriggerEvent, players map[int]*Player, now time.Time) error {
	for _, action := range resolveTriggers(boxes, event) {
		player := players[action.BoxIndex]
		if player == nil {
			return fmt.Errorf("csstrigger: no player for box %d", action.BoxIndex)
		}
		if err := player.TriggerEffect(action.AnimationIndex, action.Value, now); err != nil {
			return fmt.Errorf("csstrigger: box %d animation %d: %w", action.BoxIndex, action.AnimationIndex, err)
		}
	}
	return nil
}

// TriggerEvent is a named host event tied to the element that produced it.
type TriggerEvent struct {
	Name   string
	Source int
	Active bool
}

// Action identifies one animation action selected by a matching event.
type Action struct {
	BoxIndex, AnimationIndex int
	Name, Value              string
}

// minActions is the trigger action count kept: active uses the first,
// inactive uses the second.
const minActions = 2

// resolveTriggers matches a named event to CSS animation-trigger records.
func resolveTriggers(boxes []layout.Box, event TriggerEvent) []Action {
	if event.Source < 0 || event.Source >= len(boxes) || event.Name == "" {
		return nil
	}
	var out []Action
	for boxIndex, box := range boxes {
		if !inScope(boxes, event.Source, boxIndex, event.Name) {
			continue
		}
		for animationIndex, animation := range box.Animations {
			actions := triggerActions(animation.Trigger, event.Name)
			if len(actions) == 0 || len(actions) > minActions {
				continue
			}
			choice := 0
			if !event.Active {
				if len(actions) < minActions {
					continue
				}
				choice = 1
			}
			out = append(out, Action{
				BoxIndex:       boxIndex,
				AnimationIndex: animationIndex,
				Name:           event.Name,
				Value:          actions[choice],
			})
		}
	}
	return out
}

func inScope(boxes []layout.Box, source, target int, name string) bool {
	for index := source; index >= 0; index = boxes[index].ParentIndex {
		if scopesName(boxes[index].TriggerScope, name) && !descendant(boxes, index, target) {
			return false
		}
	}
	for index := target; index >= 0; index = boxes[index].ParentIndex {
		if scopesName(boxes[index].TriggerScope, name) && !descendant(boxes, index, source) {
			return false
		}
	}
	return true
}

func scopesName(scope, name string) bool {
	for _, token := range strings.Fields(strings.ReplaceAll(scope, ",", " ")) {
		if token == "all" || token == name {
			return true
		}
	}
	return false
}

func descendant(boxes []layout.Box, ancestor, child int) bool {
	for child >= 0 {
		if child == ancestor {
			return true
		}
		child = boxes[child].ParentIndex
	}
	return false
}

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
