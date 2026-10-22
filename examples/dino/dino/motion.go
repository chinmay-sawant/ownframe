package dino

// move slides every obstacle left and drops the ones off the screen.
func (g *game) move(dt float64) {
	kept := g.obstacles[:0]

	for _, ob := range g.obstacles {
		ob.x -= g.speed * dt
		ob.flap += dt

		if ob.x+ob.w > -40 {
			kept = append(kept, ob)
		}
	}

	g.obstacles = kept
}

// fall applies gravity while the dinosaur is off the ground. A held jump
// bounces again on landing, as the browser game does.
func (g *game) fall(dt float64) {
	if g.onGround {
		return
	}

	g.vy -= g.gravity() * dt
	g.feet += g.vy * dt

	if g.vy > 0 {
		return
	}

	if g.feet > 0 {
		return
	}

	g.feet = 0
	g.vy = 0
	g.onGround = true

	if g.jumpHeld && !g.ducking {
		g.launch()
	}
}

// gravity is lighter while a jump is held and the dinosaur is rising, so
// holding jump reaches higher.
func (g *game) gravity() float64 {
	if g.vy > 0 && g.jumpHeld {
		return heldGravity
	}

	return fallGravity
}
