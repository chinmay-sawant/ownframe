package game

// rng is a splitmix64 generator. Its whole state is one uint64, so a
// snapshot can save and restore the piece sequence exactly.
type rng struct{ state uint64 }

// next returns the next 64 random bits.
func (r *rng) next() uint64 {
	r.state += 0x9e3779b97f4a7c15
	z := r.state
	z = (z ^ (z >> splitShiftA)) * splitMixA
	z = (z ^ (z >> splitShiftB)) * splitMixB

	return z ^ (z >> splitShiftC)
}

// intn returns a value in [0, n).
func (r *rng) intn(n int) int {
	if n <= 0 {
		return 0
	}

	return int(r.next() % uint64(n)) //nolint:gosec // reason: n is guarded positive above, so the int to uint64 conversion cannot wrap.
}

// nextPiece returns the next piece from the 7-bag, shuffling a fresh bag
// when the current one runs out.
func (g *Game) nextPiece() Piece {
	if len(g.bag) == 0 {
		var b [PieceCount]Piece
		for i := range b {
			b[i] = Piece(i + 1)
		}

		for i := len(b) - 1; i > 0; i-- {
			j := g.rng.intn(i + 1)
			b[i], b[j] = b[j], b[i]
		}

		g.bag = b[:]
	}

	p := g.bag[len(g.bag)-1]
	g.bag = g.bag[:len(g.bag)-1]

	return p
}
