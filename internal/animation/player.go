package animation

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// Start starts every untriggered effect at the same clock time.
func (p *Player) Start(now time.Time) {
	p.start, p.started = now, true
	for i := range p.states {
		if p.effects[i].Trigger == "" || p.effects[i].Trigger == "none" {
			p.states[i].start, p.states[i].offset = now, 0
			p.states[i].rate, p.states[i].finished = 1, false
			p.states[i].started = true
			p.states[i].enabled = p.states[i].timing.Play != Paused
		}
	}
}

// Step samples the time-based animation frame at now.
func (p *Player) Step(now time.Time) bool { return p.StepWithInput(FrameInput{Now: now}) }

// StepWithInput samples every effect using the supplied host clock and timelines.
func (p *Player) StepWithInput(input FrameInput) bool {
	if !p.started {
		p.Start(input.Now)
	}
	if input.Now.IsZero() {
		input.Now = p.start
	}
	active, hasAvailable := false, false
	for _, target := range p.targets {
		value := target.From
		for i, effect := range p.effects {
			if len(effect.Frames) == 0 {
				continue
			}
			state := &p.states[i]
			if !state.enabled {
				continue
			}
			now := input.Now
			if state.paused {
				now = state.pausedAt
			}
			progress, _, live, apply, sampleAvailable := statePosition(state, withNow(input, now))
			if sampleAvailable {
				state.finished = !live
			}
			active = active || live && !state.paused
			hasAvailable = hasAvailable || sampleAvailable
			if sampleAvailable && apply {
				value = compose(target.From, value, effect, progress)
			}
		}
		if hasAvailable || p.touched {
			cssmotion.Write(target.Op, value)
		}
	}
	p.touched = p.touched || hasAvailable
	p.active = active
	return active
}

func withNow(in FrameInput, now time.Time) FrameInput { in.Now = now; return in }

// Active reports whether the latest sample leaves work to do.
func (p *Player) Active() bool { return p != nil && p.active }
