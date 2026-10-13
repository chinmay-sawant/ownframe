package runtimebehavior

import "time"

// StartedAt stores transition start time.
type TransitionTimer struct {
	Started time.Time
	Ended   time.Time
}

// Begin starts a transition with endpoints and duration.
func (t *Transition) Begin(from, to string, duration int) {
	if t == nil {
		return
	}
	t.From = from
	t.To = to
	t.Duration = duration
	t.Active = true
	t.Phase = "running"
}

// Cancel stops a transition.
func (t *Transition) Cancel() {
	if t == nil {
		return
	}
	t.Active = false
	t.Phase = "cancelled"
}

// Reset restores a transition.
func (t *Transition) Reset() {
	if t == nil {
		return
	}
	t.Active = false
	t.Phase = ""
}

// IsComplete reports whether transition is inactive.
func (t *Transition) IsComplete() bool {
	if t == nil {
		return true
	}
	return !t.Active
}

// BeginPhaseAlias is an alias for BeginPhase.
func (p *PhaseTransition) StartPhase(phase string) {
	p.BeginPhase(phase)
}

// End is an alias for EndPhase.
func (p *PhaseTransition) End() {
	p.EndPhase()
}

// PhaseName returns current phase.
func (p *PhaseTransition) PhaseName() string {
	if p == nil {
		return ""
	}
	return p.Phase
}

// ScrollTo sets clipped content offset.
func (c *ClipRect) ScrollTo(x, y int) {
	if c == nil {
		return
	}
	c.ScrollX = x
	c.ScrollY = y
}

// Offset returns scroll offsets.
func (c ClipRect) Offset() (int, int) {
	return c.ScrollX, c.ScrollY
}

// Intersects reports box intersection.
func (c ClipRect) Intersects(x, y, w, h int) bool {
	return x < c.X+c.W && x+w > c.X && y < c.Y+c.H && y+h > c.Y
}
