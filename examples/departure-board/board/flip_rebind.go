package board

import (
	"github.com/chinmay-sawant/ownframe"
)

// rowDelay is the frame offset between two rows of the flap cascade. Every
// cell in one screen row starts together, so one frame's changed cells stay
// inside one or two row bands. The window takes one dirty rect per frame, and
// a ripple that touches all eight rows unions into the whole frame.
const rowDelay = 8

// flapSpin is how many frames a tile spins before it settles. It lands in 8 to
// 12, rising with the tile's place across its row and down the board, so a row
// settles left to right and the last row settles at most
// rowDelay*(rowCount-1)+13 ticks after the redraw.
func flapSpin(row, col int) int { return rowDelay + (row+col/6)%5 }

// rebind reacquires the tick's ops after a Redraw and starts the cascade on
// every tile whose glyph changed. The Redraw marked its own dirty region, so
// rebind names nothing.
func (f *flipper) rebind(d *ownframe.Display, boxes []ownframe.Box, rows []Row) {
	if d == nil {
		return
	}

	byID := make(map[string]ownframe.Box)
	for _, b := range boxes {
		byID[b.ID] = b
	}

	var order []string
	next := make(map[string]*flap)

	for r, row := range rows {
		for k, c := range row.Cells {
			b, ok := byID[c.ID]
			if !ok {
				continue
			}

			op := cellOp(d, b)

			fl := f.cells[c.ID]
			if fl != nil && fl.want == c.Ch {
				fl.op = op
				fl.box = flapRect(b)
			} else {
				cur := op.Text
				if fl != nil {
					cur = fl.op.Text
				}

				fl = &flap{op: op, box: flapRect(b), want: c.Ch, cur: cur,
					delay: r * rowDelay, spin: flapSpin(r, k), done: cur == c.Ch}
			}

			next[c.ID] = fl
			order = append(order, c.ID)
		}
	}

	f.cells, f.order = next, order
	f.clock = nil

	if b, ok := byID["clock"]; ok {
		f.clock = ownframe.Text(d, b)
	}
}
