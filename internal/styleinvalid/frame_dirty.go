package styleinvalid

// MaxStyleDirtyPassesPerFrame caps style dirty passes before paint.
const MaxStyleDirtyPassesPerFrame = 2

// FrameDirtyBudget tracks style dirty passes in one frame.
type FrameDirtyBudget struct {
	max  int
	used int
}

// NewFrameDirtyBudget returns a budget capped at max (min 1).
func NewFrameDirtyBudget(max int) *FrameDirtyBudget {
	if max <= 0 {
		max = MaxStyleDirtyPassesPerFrame
	}
	return &FrameDirtyBudget{max: max}
}

// Allow reports whether another dirty pass fits.
func (b *FrameDirtyBudget) Allow() bool {
	if b == nil {
		return true
	}
	if b.used >= b.max {
		return false
	}
	b.used++
	return true
}

// Used returns consumed passes.
func (b *FrameDirtyBudget) Used() int {
	if b == nil {
		return 0
	}
	return b.used
}

// Reset clears consumed passes for next frame.
func (b *FrameDirtyBudget) Reset() {
	if b == nil {
		return
	}
	b.used = 0
}

// BeginFrame starts a new frame's dirty budget.
func (t *Tracker) BeginFrame() {
	if t == nil {
		return
	}
	t.framePasses = 0
}

// DirtyPass records one pass and reports whether allowed.
func (t *Tracker) DirtyPass() bool {
	if t == nil {
		return false
	}
	if t.framePasses >= MaxStyleDirtyPassesPerFrame {
		return false
	}
	t.framePasses++
	return true
}

// CanDirty reports whether another pass is allowed.
func (t *Tracker) CanDirty() bool {
	if t == nil {
		return false
	}
	return t.framePasses < MaxStyleDirtyPassesPerFrame
}

// DirtyPasses returns consumed passes.
func (t *Tracker) DirtyPasses() int {
	if t == nil {
		return 0
	}
	return t.framePasses
}

// MaxDirtyPasses returns the cap.
func (t *Tracker) MaxDirtyPasses() int { return MaxStyleDirtyPassesPerFrame }

// DirtyTwice reports whether at most two passes can run.
func (t *Tracker) DirtyTwice() bool { return MaxStyleDirtyPassesPerFrame == 2 }

// ResetFrame is an alias for BeginFrame.
func (t *Tracker) ResetFrame() { t.BeginFrame() }
