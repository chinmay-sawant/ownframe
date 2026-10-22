package dino

// rect is one fill in CSS pixels.
type rect struct {
	x, y, w, h float64
}

// dinoPose returns the eight fills of the dinosaur: tail, body, neck,
// head, eye, arm, and the two legs.
func (g *game) dinoPose() [8]rect {
	if g.onGround && g.ducking {
		return duckPose()
	}

	top := groundY - g.feet - dinoStandH

	switch {
	case g.phase != running:
		return standPose(top, legLong, legLong)
	case !g.onGround:
		return standPose(top, legShort, legShort)
	case int(g.run/stridePeriod)%strideCycle == 0:
		return standPose(top, legLong, legShort)
	default:
		return standPose(top, legShort, legLong)
	}
}

// standPose is the upright dinosaur with the given leg lengths.
func standPose(top, leg1, leg2 float64) [8]rect {
	return [8]rect{
		{dinoX, top + 20, 8, 8},
		{dinoX + 6, top + 20, 26, 24},
		{dinoX + 26, top + 12, 10, 14},
		{dinoX + 24, top, 22, 16},
		{dinoX + 38, top + 4, 4, 4},
		{dinoX + 32, top + 26, 10, 4},
		{dinoX + 10, top + 44, 8, leg1},
		{dinoX + 22, top + 44, 8, leg2},
	}
}

// duckPose is the low crouch under a high bird.
func duckPose() [8]rect {
	top := float64(groundY - duckHeight)

	return [8]rect{
		{dinoX, top + 14, 10, 8},
		{dinoX + 8, top + 12, 30, 18},
		{dinoX + 28, top + 10, 10, 12},
		{dinoX + 32, top + 6, 22, 14},
		{dinoX + 46, top + 10, 4, 4},
		{dinoX + 38, top + 22, 8, 4},
		{dinoX + 12, top + 26, 8, 4},
		{dinoX + 24, top + 26, 8, 4},
	}
}
