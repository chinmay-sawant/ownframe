package scene

import (
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

// ops caches the display operations the paint step changes. A Redraw or a
// resize replaces the display list, so bind rebuilds the cache when the
// page generation changes.
type ops struct {
	board [game.Rows][game.Cols]*ownframe.DisplayOp
	next  [4][4]*ownframe.DisplayOp

	score  *ownframe.DisplayOp
	level  *ownframe.DisplayOp
	lines  *ownframe.DisplayOp
	status *ownframe.DisplayOp
	theme  *ownframe.DisplayOp
}

// bind caches the operations the paint step changes.
func (s *Scene) bind(d *ownframe.Display) {
	boxes := s.page.Boxes()

	for r := 0; r < game.Rows; r++ {
		for c := 0; c < game.Cols; c++ {
			s.ops.board[r][c] = fillAt(d, boxes, cellID("b", r, c))
		}
	}

	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			s.ops.next[r][c] = fillAt(d, boxes, cellID("n", r, c))
		}
	}

	s.ops.score = textAt(d, boxes, "t-score")
	s.ops.level = textAt(d, boxes, "t-level")
	s.ops.lines = textAt(d, boxes, "t-lines")
	s.ops.status = textAt(d, boxes, "t-status")
	s.ops.theme = textAt(d, boxes, "t-theme")

	s.bound = s.page.Generation()
}
