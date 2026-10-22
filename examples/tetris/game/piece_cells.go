package game

// baseCells is each piece's spawn orientation. I uses a 4x4 box, O a 2x2
// box, and the rest a 3x3 box, so the cells rotate like SRS defines.
var baseCells = [PieceCount + 1][]Point{
	PieceI: {{0, 1}, {1, 1}, {2, 1}, {3, 1}},
	PieceO: {{0, 0}, {1, 0}, {0, 1}, {1, 1}},
	PieceT: {{1, 0}, {0, 1}, {1, 1}, {2, 1}},
	PieceS: {{1, 0}, {2, 0}, {0, 1}, {1, 1}},
	PieceZ: {{0, 0}, {1, 0}, {1, 1}, {2, 1}},
	PieceJ: {{0, 0}, {0, 1}, {1, 1}, {2, 1}},
	PieceL: {{2, 0}, {0, 1}, {1, 1}, {2, 1}},
}

// Cells returns the piece's cells in its own box for the rotation. The
// result is fresh, so callers may keep it.
func (p Piece) Cells(r Rotation) []Point {
	out := make([]Point, len(baseCells[p]))
	copy(out, baseCells[p])

	n := p.box()
	for i := uint8(0); i < uint8(r%rotationCount); i++ {
		for j := range out {
			out[j] = Point{X: n - 1 - out[j].Y, Y: out[j].X}
		}
	}

	return out
}

// box returns the rotation box size.
func (p Piece) box() int {
	switch p {
	case PieceI:
		return boxI
	case PieceO:
		return boxO
	default:
		return boxDefault
	}
}
