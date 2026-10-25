package csstransition

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// paintSlack is the room a captured paint keeps when a display hands the
// operation back. A layout that rounds a value differently still describes
// the same operation.
const paintSlack = 0.5

// endpoints returns the start and destination paints of one operation. A
// field the operation still carries inside the captured range keeps the
// captured value, so the transition animates it. A field the layout changed
// keeps the operation's current value, so a rebind does not write a stale
// size or position back.
func endpoints(op *layout.DisplayOp, initial, final cssmotion.Paint) (cssmotion.Paint, cssmotion.Paint) {
	current := cssmotion.Read(op)

	return adjust(initial, initial, final, current), adjust(final, initial, final, current)
}
