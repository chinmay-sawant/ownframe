package board

import (
	"github.com/chinmay-sawant/ownframe"
)

// cellSlack is how far left of its box a centered glyph may paint. The 14px
// gap tiles center the wider braille filler, so its op starts left of the
// box, past ownframe.Text's half-point tolerance.
const cellSlack = 3

// cellOp returns a tile's text op. A miss retries with the box moved left by
// cellSlack. The 14px gap tiles center the wider braille filler, so its op
// starts left of the box, past ownframe.Text's half-point tolerance. Neighbour
// tiles sit a full glyph apart, so the shifted box cannot reach their ops.
func cellOp(d *ownframe.Display, b ownframe.Box) *ownframe.DisplayOp {
	if op := ownframe.Text(d, b); op != nil {
		return op
	}

	b.X -= cellSlack

	return ownframe.Text(d, b)
}
