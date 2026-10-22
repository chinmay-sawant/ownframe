package game

// Rune returns the piece letter.
func (p Piece) Rune() byte {
	switch p {
	case PieceI:
		return 'I'
	case PieceO:
		return 'O'
	case PieceT:
		return 'T'
	case PieceS:
		return 'S'
	case PieceZ:
		return 'Z'
	case PieceJ:
		return 'J'
	case PieceL:
		return 'L'
	}

	return '?'
}

// String returns the piece letter.
func (p Piece) String() string { return string(p.Rune()) }

// Valid reports whether p is one of the seven kinds.
func (p Piece) Valid() bool { return p >= PieceI && p <= PieceL }

// CW returns the next rotation clockwise.
func (r Rotation) CW() Rotation { return (r + 1) % rotationCount }

// CCW returns the next rotation counter-clockwise.
func (r Rotation) CCW() Rotation { return (r + ccwOffset) % rotationCount }

// String returns the rotation name.
func (r Rotation) String() string {
	return [...]string{"0", "R", "2", "L"}[r%rotationCount]
}

// String returns the phase name.
func (p Phase) String() string {
	return [...]string{"ready", "running", "paused", "over"}[p%4]
}
