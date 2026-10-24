package animation

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// Start starts every untriggered effect at the same clock time.
func (p *Player) Start(now time.Time) {
	p.ensureBound()
	p.start, p.started = now, true
	p.resetCursors()
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
// Effects are sampled once and their value sets are applied to every target,
// so a frame does no per-target keyframe work and allocates nothing.
func (p *Player) StepWithInput(input FrameInput) bool {
	if !p.started {
		p.Start(input.Now)
	}
	if input.Now.IsZero() {
		input.Now = p.start
	}
	p.ensureBound()
	p.resetRestamp()
	for i := range p.current {
		p.current[i] = p.targets[i].From
	}
	active, hasAvailable := false, false
	for i := range p.effects {
		if len(p.effects[i].Frames) == 0 {
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
		table := &p.tables[i]
		if table.wasPaused && !state.paused {
			table.resetCursors()
		}
		table.wasPaused = state.paused
		if !table.lastNow.IsZero() && now.Before(table.lastNow) {
			table.resetCursors()
		}
		table.lastNow = now
		if sampleAvailable && apply {
			if table.transform {
				for target := range p.targets {
					current, sample := composeSampled(&p.targets[target],
						p.current[target], p.effects[i], progress)
					p.current[target] = current
					p.noteRestamp(target, sample)
				}
				continue
			}
			var set valueSet
			table.sample(progress, &set)
			for target := range p.targets {
				set.applyTo(&p.current[target], p.targets[target].From, table.mode)
			}
		}
	}
	p.collectRestamp()
	for target := range p.current {
		p.current[target].NeedsRestamp = p.restampSlots[target].active()
	}
	if hasAvailable || p.touched {
		for target := range p.targets {
			cssmotion.Write(p.targets[target].Op, p.current[target])
		}
	}
	p.touched = p.touched || hasAvailable
	p.active = active
	return active
}

func withNow(in FrameInput, now time.Time) FrameInput { in.Now = now; return in }

// RestampValue is one target's latest sampled values that the 2D matrix path
// could not write. Transform is the canonical CSS transform list when the
// sample included a transform; it stays empty for a perspective-only sample.
// Perspective is the camera distance in layout points. PerspectiveOriginX and
// PerspectiveOriginY are the sampled perspective-origin in layout points, as an
// offset from the element's border-box origin; HasPerspectiveOrigin reports
// whether the keyframe declared one. Target indexes the targets the player was
// built with, so a scene maps it to the element and operation it stands for.
type RestampValue struct {
	Target    int
	Transform string

	Perspective          float64
	HasPerspective       bool
	PerspectiveOriginX   float64
	PerspectiveOriginY   float64
	HasPerspectiveOrigin bool
}

// RestampValues returns the samples the latest Step could not write as a 2D
// matrix, one per target. A 2D-only sample returns nil. The slice is reused by
// the next Step, so a caller consumes it before stepping again.
func (p *Player) RestampValues() []RestampValue {
	if p == nil || len(p.restamp) == 0 {
		return nil
	}

	return p.restamp
}

// resetRestamp sizes the per-target slots to the target list and empties the
// exposed slice. Slots are cleared rather than reallocated, so a live 3D or
// perspective sample keeps its backing arrays.
func (p *Player) resetRestamp() {
	if len(p.restampSlots) != len(p.targets) {
		p.restampSlots = make([]restampSlot, len(p.targets))
		p.restamp = make([]RestampValue, 0, len(p.targets))
	}
	for i := range p.restampSlots {
		p.restampSlots[i].reset()
	}
	p.restamp = p.restamp[:0]
}

// noteRestamp records the values sampled for one target. A transform sample
// replaces the previous transform value even when this effect wrote a 2D
// matrix, so a perspective that needs the same element restamped carries the
// animated transform the operation already shows. A later perspective-less
// 2D transform still clears the 3D request because transformRestamp is false.
func (p *Player) noteRestamp(target int, sample restampSample) {
	if target < 0 || target >= len(p.restampSlots) {
		return
	}
	slot := &p.restampSlots[target]
	if sample.hasTransform {
		slot.transform = sample.transform
		slot.hasTransform = true
		slot.transformRestamp = sample.transformRestamp
	}
	if sample.hasPerspective {
		slot.perspective = sample.perspective
		slot.hasPerspective = true
	}
}

// collectRestamp rebuilds the exposed slice from the per-target slots,
// serializing each transform once. A player with no 3D transform and no
// perspective leaves it empty, so RestampValues reports nil.
func (p *Player) collectRestamp() {
	p.restamp = p.restamp[:0]
	for i := range p.restampSlots {
		slot := &p.restampSlots[i]
		if !slot.active() {
			continue
		}
		value := RestampValue{Target: i}
		if slot.hasPerspective {
			value.HasPerspective = slot.perspective.HasDistance
			value.Perspective = slot.perspective.Distance
			value.PerspectiveOriginX = slot.perspective.OriginX
			value.PerspectiveOriginY = slot.perspective.OriginY
			value.HasPerspectiveOrigin = slot.perspective.HasOrigin
		}
		if slot.hasTransform {
			p.restampBuf = AppendTransformCSS(p.restampBuf[:0], slot.transform)
			value.Transform = string(p.restampBuf)
		}
		p.restamp = append(p.restamp, value)
	}
}

// Active reports whether the latest sample leaves work to do.
func (p *Player) Active() bool { return p != nil && p.active }
