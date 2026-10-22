package flappy

// wingPoses is the vertical wing offset over one cycle: down, level, up.
var wingPoses = [3]float64{-3, 0, 3}

// birdRects returns the six fills of the bird, in paint order: body,
// belly, wing, eye, pupil, beak. The ready phase flaps with the bob.
func (g *game) birdRects() [birdCount]rect {
	age := g.flapAge
	if g.phase == ready {
		age = g.run
	}

	wing := wingPoses[int(age*wingRate)%3]

	return [birdCount]rect{
		{birdX - 17, g.birdY - 13.5, 34, 27},
		{birdX - 13, g.birdY + 3, 16, 10},
		{birdX - 17, g.birdY - 5 + wing, 13, 12},
		{birdX + 5, g.birdY - 10, 8, 8},
		{birdX + 8, g.birdY - 8, 4, 4},
		{birdX + 14, g.birdY - 3, 9, 7},
	}
}

// pipeRects returns the four fills of one pair: top body, top cap, bottom
// body, bottom cap.
func (g *game) pipeRects(p pipe) [pipePartCount]rect {
	top := p.gapY - p.gap/halfDiv
	bot := p.gapY + p.gap/halfDiv
	side := p.x - (capW-pipeW)/halfDiv

	return [pipePartCount]rect{
		{p.x, 0, pipeW, max(0, top-capH)},
		{side, top - capH, capW, capH},
		{p.x, bot + capH, pipeW, max(0, groundY-bot-capH)},
		{side, bot, capW, capH},
	}
}
