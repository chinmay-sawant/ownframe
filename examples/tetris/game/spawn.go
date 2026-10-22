package game

// spawnX returns the box origin column: centered, with O straddling the
// middle columns.
func spawnX(p Piece) int {
	if p == PieceO {
		return spawnOriginO
	}

	return spawnOrigin
}

// spawnY returns the box origin row. Occupied cells start in rows -1 and
// 0; cells above the board are legal, so a piece can spawn partly hidden.
func spawnY(Piece) int { return -1 }

// spawn presents the next piece and tops out when it cannot appear.
func (g *Game) spawn() []Event {
	g.fillPreview()
	g.Piece = g.Next[0]
	g.Next = g.Next[1:]
	g.fillPreview()
	g.Rot = Rot0
	g.X, g.Y = spawnX(g.Piece), spawnY(g.Piece)
	g.fall, g.lock, g.resets, g.grounded = 0, 0, 0, false

	if g.collides(g.Piece, g.Rot, g.X, g.Y) {
		g.topOut()

		return []Event{{Kind: EventTopOut, Score: g.Score}}
	}

	return []Event{{Kind: EventSpawn, Score: g.Score}}
}

// fillPreview keeps the preview queue at Preview pieces.
func (g *Game) fillPreview() {
	for len(g.Next) < Preview {
		g.Next = append(g.Next, g.nextPiece())
	}
}
