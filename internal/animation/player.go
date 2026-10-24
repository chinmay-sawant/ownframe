package animation

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// Start starts every untriggered effect at the same clock time. An effect
// declared animation-play-state:paused starts enabled but frozen: it keeps
// sampling its first pose instead of disappearing until a resume.
func (p *Player) Start(now time.Time) {
	p.ensureBound()
	p.start, p.started = now, true
	p.resetCursors()
	for i := range p.states {
		if p.effects[i].Trigger == "" || p.effects[i].Trigger == "none" {
			s := &p.states[i]
			s.start, s.offset = now, 0
			s.rate, s.finished = 1, false
			s.started, s.enabled = true, true
			s.paused = s.timing.Play == Paused
			s.pausedAt = time.Time{}
			s.hasHold = false
			if s.paused {
				s.pausedAt = now
			}
		}
	}
}

// Pause freezes every running effect at now. The active time an effect has
// accumulated is kept, so a later Resume continues the same phase instead of
// starting over; an effect already paused keeps its earlier freeze, and a
// triggered effect that never started is left for its trigger.
//
// A time-based effect also records the exact position it froze at. The position
// is the sampled phase with the signed start delay folded in, so a Clock taken
// after the pause carries it across a rebind without recomputing it from the
// clock origin.
func (p *Player) Pause(now time.Time) {
	if p == nil {
		return
	}
	for i := range p.states {
		s := &p.states[i]
		if !s.started || s.paused {
			continue
		}
		s.offset = currentOffset(s, now)
		if raw, ok := frozenPosition(s); ok {
			s.holdRaw, s.hasHold = raw, true
			s.lastRaw, s.hasRaw = raw, true
		}
		s.paused, s.pausedAt = true, now
		s.enabled = true
	}
}

// Resume continues every paused effect from the phase it froze at. The time
// spent paused is not counted: the current run starts at now with the frozen
// offset, so a sample after a resume matches an uninterrupted run at the same
// active time.
func (p *Player) Resume(now time.Time) {
	if p == nil {
		return
	}
	for i := range p.states {
		s := &p.states[i]
		if !s.started || !s.paused {
			continue
		}
		s.start, s.paused, s.pausedAt = now, false, time.Time{}
		s.enabled = true
		if timeBased(s.timing) {
			// The frozen sample the pause recorded no longer holds: the next
			// sample advances from the active time it reached.
			s.hasHold = false
		}
		if i < len(p.tables) {
			p.tables[i].resetCursors()
		}
	}
}

// Paused reports whether every started effect is frozen. A player with no
// started effect, or with any running effect, reports false.
func (p *Player) Paused() bool {
	if p == nil {
		return false
	}
	started := false
	for i := range p.states {
		if !p.states[i].started {
			continue
		}
		started = true
		if !p.states[i].paused {
			return false
		}
	}

	return started
}

// Clock is the animation clock a player carries across a rebuild: the clock
// origin of the current run, the active time accumulated before it, and the
// freeze a pause left. A page that rebuilds a player on every redraw restores
// this so the animation keeps its phase instead of starting or snapping.
//
// Position is the exact paused sample: the position in iterations, with the
// signed start delay folded in and a reverse rate already applied. HasPosition
// reports that the paused effect had a sample to carry. A paused rebind
// restores it instead of recomputing the phase from Start and Offset, so a
// negative delay cannot be counted twice and a timing value that changed with
// the new stylesheet cannot move the frozen pose.
type Clock struct {
	Start    time.Time
	Offset   time.Duration
	Paused   bool
	PausedAt time.Time

	Position    float64
	HasPosition bool
}

// Clock reports the clock of the first started effect and false when no effect
// has started. The CSS animations on one element share that clock; an effect
// that a trigger gave its own start or rate is not represented. The clock does
// not fold in the time since Start; a caller that needs the phase at a later
// time adds that elapsed time to Offset.
func (p *Player) Clock() (Clock, bool) {
	if p == nil {
		return Clock{}, false
	}
	for i := range p.states {
		s := &p.states[i]
		if !s.started {
			continue
		}
		clock := Clock{Start: s.start, Offset: s.offset, Paused: s.paused, PausedAt: s.pausedAt}
		if s.paused && timeBased(s.timing) && s.hasHold {
			clock.Position, clock.HasPosition = s.holdRaw, true
		}

		return clock, true
	}

	return Clock{}, false
}

// SetClock puts a freshly started player on a clock another player carried. A
// running clock keeps its origin and offset, so a rebind does not restart the
// phase. A paused clock freezes every effect at the carried offset until
// Resume, which continues from that phase. When the clock carries the paused
// sample itself, the freeze uses it directly: the sampled position then does
// not depend on the new player's start or its timing values.
func (p *Player) SetClock(c Clock) {
	if p == nil {
		return
	}
	for i := range p.states {
		s := &p.states[i]
		if !s.started {
			continue
		}
		s.start, s.offset = c.Start, c.Offset
		s.paused, s.pausedAt = c.Paused, c.PausedAt
		s.enabled = true
		s.finished = false
		if timeBased(s.timing) {
			s.hasHold = c.Paused && c.HasPosition
			if s.hasHold {
				s.holdRaw, s.lastRaw, s.hasRaw = c.Position, c.Position, true
			}
		}
		if i < len(p.tables) {
			p.tables[i].resetCursors()
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
		if state.paused && !state.pausedAt.IsZero() {
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

// TargetPaint returns one target's latest sampled paint and the paint the
// target was bound with, in target order. A compositor layer reads the pair:
// current.Matrix composed with the inverse of base.Matrix is the animation
// delta from the captured pose to the pose the display carries now, and
// current.Opacity is the element's group opacity.
//
// Before the first sample both values are the base paint, which is the pose the
// operation already carries. A target index outside the player reports false.
func (p *Player) TargetPaint(target int) (current, base cssmotion.Paint, ok bool) {
	if p == nil || target < 0 || target >= len(p.targets) {
		return cssmotion.Paint{}, cssmotion.Paint{}, false
	}
	base = p.targets[target].From
	if !p.touched || len(p.current) != len(p.targets) {
		return base, base, true
	}

	return p.current[target], base, true
}

// BasePaints appends the paint every target composes from to buf and returns
// the extended slice. A rebind that finds its display already carrying the
// player's own sample reads this from the player it replaces, so the new
// player composes from the declared paint instead of stacking the sample on
// itself.
func (p *Player) BasePaints(buf []cssmotion.Paint) []cssmotion.Paint {
	if p == nil {
		return buf
	}
	for i := range p.targets {
		buf = append(buf, p.targets[i].From)
	}

	return buf
}

// SetBasePaints replaces the paint each target composes from, in target order.
// A shorter list leaves the remaining targets alone. The base is captured at
// bind, so this is for a rebind that has to repair a base another display
// carried; the composed sample of the next Step starts from it.
func (p *Player) SetBasePaints(bases []cssmotion.Paint) {
	if p == nil {
		return
	}
	for i := range p.targets {
		if i >= len(bases) {
			return
		}
		p.targets[i].From = bases[i]
	}
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
