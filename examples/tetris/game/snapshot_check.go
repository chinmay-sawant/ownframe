package game

import "fmt"

// Validate rejects a snapshot outside the model bounds.
func (s Snapshot) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("%w: missing id", ErrInvalidSnapshot)
	}

	if s.Ruleset != Ruleset {
		return fmt.Errorf("%w: ruleset %q", ErrInvalidSnapshot, s.Ruleset)
	}

	if s.FixtureVersion != FixtureVersion {
		return fmt.Errorf("%w: fixture version %d", ErrInvalidSnapshot, s.FixtureVersion)
	}

	if !s.Piece.Valid() || s.Rot > RotL {
		return fmt.Errorf("%w: piece %d rotation %d", ErrInvalidSnapshot, s.Piece, s.Rot)
	}

	board, err := ParseBoard(s.Board)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidSnapshot, err)
	}

	if s.X < -4 || s.X > Cols || s.Y < -4 || s.Y > Rows {
		return fmt.Errorf("%w: position %d,%d", ErrInvalidSnapshot, s.X, s.Y)
	}

	scratch := &Game{Board: board}
	if scratch.collides(s.Piece, s.Rot, s.X, s.Y) {
		return fmt.Errorf("%w: piece overlaps the board", ErrInvalidSnapshot)
	}

	return s.validateValues()
}

// validateValues checks numbers and queues.
func (s Snapshot) validateValues() error {
	if s.Score < 0 || s.Score > MaxScore || s.Lines < 0 || s.Lines > MaxLines {
		return fmt.Errorf("%w: score %d lines %d", ErrInvalidSnapshot, s.Score, s.Lines)
	}

	if s.Level < 1 || s.Level > MaxLevel || s.Pieces < 0 || s.Pieces > MaxPieces {
		return fmt.Errorf("%w: level %d pieces %d", ErrInvalidSnapshot, s.Level, s.Pieces)
	}

	if s.ElapsedMS < 0 || s.ElapsedMS > MaxDuration.Milliseconds() {
		return fmt.Errorf("%w: elapsed %d", ErrInvalidSnapshot, s.ElapsedMS)
	}

	if s.FallMS < 0 || s.LockMS < 0 {
		return fmt.Errorf("%w: timers %d %d", ErrInvalidSnapshot, s.FallMS, s.LockMS)
	}

	if len(s.Next) > Preview || len(s.Bag) > PieceCount {
		return fmt.Errorf("%w: queues %d %d", ErrInvalidSnapshot, len(s.Next), len(s.Bag))
	}

	for _, p := range append(append([]Piece(nil), s.Next...), s.Bag...) {
		if !p.Valid() {
			return fmt.Errorf("%w: bad queued piece %d", ErrInvalidSnapshot, p)
		}
	}

	return nil
}
