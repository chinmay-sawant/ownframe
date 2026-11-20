package frame

import (
	"image"

	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
)

// OpBox returns op's painted box in CSS pixels: the box to mark when the op
// moves, fades, or changes shape. It honors a baked transform and the text
// baseline, matching the page's own dirty boxes.
func OpBox(d *page.Display, op *page.DisplayOp) image.Rectangle {
	return dirty.OpBounds(op, d)
}
