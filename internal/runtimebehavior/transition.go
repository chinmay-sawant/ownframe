package runtimebehavior

// Transition describes a CSS-like phase transition.
type Transition struct {
	Property string
	From, To string
	Duration int // ms
	Active   bool
	Phase    string
}

// Start begins the transition.
func (t *Transition) Start() {
	if t == nil {
		return
	}
	t.Active = true
}

// Finish ends the transition.
func (t *Transition) Finish() {
	if t == nil {
		return
	}
	t.Active = false
}

// PhaseTransition describes a mount/update/unmount phase transition.
type PhaseTransition struct {
	Transition
	Phase string
}

// BeginPhase starts a named phase transition.
func (p *PhaseTransition) BeginPhase(phase string) {
	if p == nil {
		return
	}
	p.Phase = phase
	p.Active = true
}

// EndPhase ends the current phase transition.
func (p *PhaseTransition) EndPhase() {
	if p == nil {
		return
	}
	p.Active = false
}

// ClipRect is a component-scoped clipping/scrolling container.
type ClipRect struct {
	X, Y, W, H int
	ScrollX    int
	ScrollY    int
}

// Contains reports whether the point is inside the visible clip rect.
func (c ClipRect) Contains(x, y int) bool {
	return x >= c.X && y >= c.Y && x < c.X+c.W && y < c.Y+c.H
}

// ContentContains reports whether content coordinates are visible.
func (c ClipRect) ContentContains(x, y int) bool {
	return c.Contains(x-c.ScrollX, y-c.ScrollY)
}

// ScrollBy offsets the clipped content.
func (c *ClipRect) ScrollBy(dx, dy int) {
	if c == nil {
		return
	}
	c.ScrollX += dx
	c.ScrollY += dy
}

// ResetScroll clears the clipped content offset.
func (c *ClipRect) ResetScroll() {
	if c == nil {
		return
	}
	c.ScrollX = 0
	c.ScrollY = 0
}
