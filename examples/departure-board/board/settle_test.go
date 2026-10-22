package board_test

import (
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/departure-board/board"
)

// tileOp finds a tile's text op the way the board does. The 14px gap tiles
// center the wider braille filler left of their box, outside ownframe.Text's
// half-point tolerance, so a miss retries with the box shifted left.
func tileOp(d *ownframe.Display, b ownframe.Box) *ownframe.DisplayOp {
	if op := ownframe.Text(d, b); op != nil {
		return op
	}

	b.X -= 3

	return ownframe.Text(d, b)
}

// settled reports whether every glyph cell's text op matches its view target.
func settled(app *board.App, want map[string]string) bool {
	d := app.Page().Display()
	if d == nil {
		return false
	}

	for id, ch := range want {
		found := false

		for _, b := range app.Boxes() {
			if b.ID != id {
				continue
			}

			found = true

			if op := tileOp(d, b); op == nil || op.Text != ch {
				return false
			}
		}

		if !found {
			return false
		}
	}

	return true
}
