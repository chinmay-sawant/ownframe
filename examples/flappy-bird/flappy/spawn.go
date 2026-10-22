package flappy

import "math/rand/v2"

// spacing is the travel before the next pipe pair, shrinking with distance.
func (g *game) spacing(rng *rand.Rand) float64 {
	return max(spacingMin, spacingStart-g.distance*0.03) + rng.Float64()*40
}

// spawn adds one pipe pair at the right edge. The gap shrinks with score.
func (g *game) spawn(rng *rand.Rand) {
	gap := max(gapMin, gapStart-float64(g.score)*3)
	top := margin + rng.Float64()*(groundY-2*margin-gap)

	g.pipes = append(g.pipes, pipe{x: sceneW + spawnOffset, gapY: top + gap/2, gap: gap})
}

// scored counts a pipe once when its right edge passes the bird.
func (g *game) scored() {
	for i := range g.pipes {
		p := &g.pipes[i]

		if !p.scored && p.x+pipeW < birdX {
			p.scored = true
			g.score++
		}
	}
}
