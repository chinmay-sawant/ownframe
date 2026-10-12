package board_test

import (
	"github.com/chinmay-sawant/ownframe/examples/departure-board/board"
	"github.com/chinmay-sawant/ownframe/internal/frame"
)

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

			if op := frame.Text(d, b); op == nil || op.Text != ch {
				return false
			}
		}

		if !found {
			return false
		}
	}

	return true
}
