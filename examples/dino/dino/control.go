package dino

// jump starts a run from the ready screen, a new run after a crash, or a
// leap while running.
func (g *game) jump() {
	switch g.phase {
	case ready:
		g.phase = running
	case over:
		g.again()
	default:
		if g.onGround && !g.ducking {
			g.launch()
		}
	}
}

// again starts a fresh run after a crash.
func (g *game) again() {
	g.restart()
	g.phase = running
}

// launch lifts the dinosaur off the ground.
func (g *game) launch() {
	g.vy = 660
	g.onGround = false
}

// restart resets the run and keeps the best score.
func (g *game) restart() {
	high := max(g.high, g.score)
	*g = newGame()
	g.high = high
}

// setDuck holds or releases the crouch. Pressing down in the air dives.
func (g *game) setDuck(on bool) {
	if g.phase != running {
		return
	}

	if !on {
		g.ducking = false

		return
	}

	g.ducking = true

	if !g.onGround && g.vy > diveVelocity {
		g.vy = diveVelocity
	}
}
