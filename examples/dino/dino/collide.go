package dino

// dinoBox is the collision box for the current pose. It is a little
// smaller than the sprite, so near misses count as misses.
func (g *game) dinoBox() (x, y, w, h float64) {
	top := groundY - g.feet

	if g.onGround && g.ducking {
		return dinoX + dinoBoxInset, top - duckBoxH, duckBoxW, duckBoxH
	}

	return dinoX + dinoBoxInset, top - standBoxH, standBoxW, standBoxH
}

// hits reports whether the dinosaur overlaps an obstacle.
func (g *game) hits() bool {
	dx, dy, dw, dh := g.dinoBox()

	for _, ob := range g.obstacles {
		ox := ob.x + obInset
		oy := groundY - ob.bottom - ob.h + obInset
		ow := ob.w - obShrink
		oh := ob.h - obShrink

		if dx < ox+ow && ox < dx+dw && dy < oy+oh && oy < dy+dh {
			return true
		}
	}

	return false
}
