package game

// move shifts the piece one cell and restarts the lock delay when it
// stays grounded.
func (g *Game) move(dx, dy int) bool {
	if g.collides(g.Piece, g.Rot, g.X+dx, g.Y+dy) {
		return false
	}

	g.X += dx
	g.Y += dy
	if dy > 0 {
		g.resets = 0
	}

	g.resetLock()

	return true
}

// rotate turns the piece, trying the SRS wall kicks in order.
func (g *Game) rotate(cw bool) bool {
	to := g.Rot.CW()
	if !cw {
		to = g.Rot.CCW()
	}

	for _, k := range kicks(g.Piece, g.Rot, cw) {
		if g.collides(g.Piece, to, g.X+k.dx, g.Y+k.dy) {
			continue
		}

		g.Rot = to
		g.X += k.dx
		g.Y += k.dy
		g.resetLock()

		return true
	}

	return false
}

// hardDrop lands the piece and locks it at once.
func (g *Game) hardDrop() []Event {
	cells := 0
	for g.canDrop() {
		g.Y++
		cells++
	}

	g.Score = addScore(g.Score, hardDropPoints*cells)
	ev := []Event{{Kind: EventHardDrop, Score: g.Score}}

	return append(ev, g.lockPiece()...)
}

// ActiveCells returns the current piece's absolute board cells, in spawn
// order. Cells with negative y sit above the visible board.
func (g *Game) ActiveCells() []Point {
	if !g.Piece.Valid() {
		return nil
	}

	cells := g.Piece.Cells(g.Rot)
	out := make([]Point, len(cells))
	for i, c := range cells {
		out[i] = Point{X: g.X + c.X, Y: g.Y + c.Y}
	}

	return out
}

// GhostCells returns where the current piece would land.
func (g *Game) GhostCells() []Point {
	if !g.Piece.Valid() {
		return nil
	}

	y := g.landY()
	cells := g.Piece.Cells(g.Rot)
	out := make([]Point, len(cells))
	for i, c := range cells {
		out[i] = Point{X: g.X + c.X, Y: y + c.Y}
	}

	return out
}
