package game

import "time"

// Start begins play: it spawns on the ready screen, resumes a paused
// game, and behaves like Restart after a top-out.
func (g *Game) Start() []Event {
	switch g.Phase {
	case PhaseReady:
		g.Phase = PhaseRunning

		return g.spawn()
	case PhasePaused:
		g.Phase = PhaseRunning

		return []Event{{Kind: EventPause}}
	case PhaseOver:
		return g.Restart()
	}

	return nil
}

// Restart clears every transient value and begins a fresh run. The new
// seed comes from the current generator, so runs differ after a restart.
func (g *Game) Restart() []Event {
	seed := g.rng.next()
	g.reset(seed)

	return g.spawn()
}

// reset returns the game to a running empty board with a new ID.
func (g *Game) reset(seed uint64) {
	*g = Game{
		ID:    newID(),
		Seed:  seed,
		Phase: PhaseRunning,
		Level: 1,
		rng:   rng{state: seed},
	}
}

// SetPaused pauses a running game and resumes a paused one.
func (g *Game) SetPaused(paused bool) []Event {
	switch {
	case paused && g.Phase == PhaseRunning:
		g.Phase = PhasePaused

		return []Event{{Kind: EventPause}}
	case !paused && g.Phase == PhasePaused:
		g.Phase = PhaseRunning

		return []Event{{Kind: EventPause}}
	}

	return nil
}

// TogglePause flips the paused state of a running or paused game.
func (g *Game) TogglePause() []Event {
	if g.Phase == PhasePaused {
		return g.SetPaused(false)
	}

	return g.SetPaused(true)
}

// gravity returns the time per row at the current level, bounded at both
// ends.
func (g *Game) gravity() time.Duration {
	ms := gravityBase - gravityStep*(g.Level-1)
	if ms < gravityMin {
		ms = gravityMin
	}

	return time.Duration(ms) * time.Millisecond
}
