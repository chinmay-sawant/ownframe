package flappy

// birdBox is the collision box, a little smaller than the sprite.
func (g *game) birdBox() (x, y, w, h float64) {
	return birdX - collideDX, g.birdY - collideDY, collideW, collideH
}

// hits reports whether the bird touched the ground or a pipe.
func (g *game) hits() bool {
	x, y, w, h := g.birdBox()

	if y+h >= groundY {
		return true
	}

	for _, p := range g.pipes {
		if p.x+pipeW <= x || p.x >= x+w {
			continue
		}

		if y < p.gapY-p.gap/2 || y+h > p.gapY+p.gap/2 {
			return true
		}
	}

	return false
}
