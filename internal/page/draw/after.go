package draw

// After is the post-commit hook and the guards that keep it from
// re-entering the redraw that called it.
type After struct {
	fn     func()
	busy   bool
	ask    bool
	served bool
}

// Set registers fn. nil removes the callback.
func (a *After) Set(fn func()) { a.fn = fn }

// Busy reports that Run is inside the callback.
func (a *After) Busy() bool { return a.busy }

// Ask records a redraw the callback requested. The commit that runs the
// callback must not start another one inside it.
func (a *After) Ask() { a.ask = true }

// Run calls the callback. A nested Run while busy returns without calling.
func (a *After) Run() {
	fn := a.fn
	if fn == nil || a.busy {
		return
	}

	a.busy = true
	defer func() { a.busy = false }()

	fn()
}

// Take reports that the caller should draw once for a queued ask.
// A second take before Done is a no-op, so a callback that always asks
// cannot loop. An ask that arrives while one draw is already served is
// dropped, and the served flag stays set.
func (a *After) Take() bool {
	if !a.ask {
		return false
	}

	a.ask = false
	if a.served {
		return false
	}

	a.served = true

	return true
}

// Done clears the one-draw guard after the follow-up redraw returns.
func (a *After) Done() { a.served = false }
