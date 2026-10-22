package flappy

import (
	"math"
	"math/rand/v2"
)

// step advances the simulation by dt seconds. The ready screen bobs; a
// crash only lets the bird fall; a run moves, scores, and can end.
func (g *game) step(dt float64, rng *rand.Rand) {
	g.run += dt

	switch g.phase {
	case ready:
		g.birdY = sceneH/halfDiv + math.Sin(g.run*bobFreq)*bobAmp

		return
	case over:
		g.fall(dt)

		return
	}

	g.distance += g.speed * dt
	g.speed = min(speedMax, speedStart+g.distance*0.02)

	g.nextPipe -= g.speed * dt
	if g.nextPipe <= 0 {
		g.spawn(rng)
		g.nextPipe = g.spacing(rng)
	}

	g.move(dt)
	g.fall(dt)
	g.scored()

	if g.hits() {
		g.phase = over
		g.best = max(g.best, g.score)
	}
}

// move slides every pipe left and drops the ones off the screen.
func (g *game) move(dt float64) {
	kept := g.pipes[:0]

	for _, p := range g.pipes {
		p.x -= g.speed * dt

		if p.x+pipeW > -40 {
			kept = append(kept, p)
		}
	}

	g.pipes = kept
}

// fall applies gravity. The ceiling clamps without a crash; after a crash
// the bird also stops at the ground.
func (g *game) fall(dt float64) {
	g.vy = min(maxFall, g.vy+gravity*dt)
	g.birdY += g.vy * dt
	g.flapAge += dt

	if g.birdY < birdR {
		g.birdY = birdR
		g.vy = 0
	}

	if g.phase == over && g.birdY > groundY-birdR {
		g.birdY = groundY - birdR
		g.vy = 0
	}
}
