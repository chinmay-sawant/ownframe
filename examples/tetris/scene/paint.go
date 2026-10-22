package scene

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

// bitmapEvery throttles full repaints on a page that fell back to the
// bitmap path, where every visible change is a Redraw.
const bitmapEvery = 50 * time.Millisecond

// paint draws the current frame from the retained operations. Routine
// frames take this path; nothing here parses or lays out the page.
func (s *Scene) paint(ctx context.Context) error {
	if s.hist.open {
		return nil
	}

	d := s.page.Display()
	if d == nil {
		return s.paintBitmap(ctx)
	}

	if s.overlays {
		s.overlays = false

		if err := s.redraw(ctx); err != nil {
			return err
		}

		d = s.page.Display()
		if d == nil {
			return nil
		}
	}

	if s.page.Generation() != s.bound {
		s.bind(d)
	}

	s.paintBoard(s.frame)
	s.paintNext(s.frame)
	s.paintTexts(s.frame)

	return nil
}

// paintBitmap rebuilds the picture on a page the replay rejected, where
// the only way to change the frame is a Redraw.
func (s *Scene) paintBitmap(ctx context.Context) error {
	if !s.dirty {
		return nil
	}

	now := s.now()
	if !s.bitmap.IsZero() && now.Sub(s.bitmap) < bitmapEvery {
		return nil
	}

	s.bitmap = now

	return s.redraw(ctx)
}

// paintBoard sets the colour of every board cell from the frame.
func (s *Scene) paintBoard(f Frame) {
	for r := 0; r < game.Rows; r++ {
		for c := 0; c < game.Cols; c++ {
			op := s.ops.board[r][c]
			if op == nil {
				continue
			}

			k, ghost := f.cellKind(r, c)
			or, og, ob, oa := cellColor(k, ghost, s.dark)
			op.R, op.G, op.B, op.Alpha = or, og, ob, oa
		}
	}
}

// paintNext sets the preview cells for the next piece.
func (s *Scene) paintNext(f Frame) {
	set := nextSet(f.Next)

	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			op := s.ops.next[r][c]
			if op == nil {
				continue
			}

			k := game.Piece(0)
			if set[r][c] {
				k = f.Next
			}

			or, og, ob, oa := cellColor(k, false, s.dark)
			op.R, op.G, op.B, op.Alpha = or, og, ob, oa
		}
	}
}
