package motion

import (
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

// BindTransition builds a transition renderer for one CSS property on a box. The
// destination comes from the display, so a caller that rebinds later wants
// BindTransitionTo with the paints it captured on the first bind. BindTransition trusts the
// captured indices, because the paint it writes as the destination is the
// one the display holds. BindTransitionTo checks them, because it writes a paint the
// caller captured on an earlier display.
func BindTransition(d *layout.Display, box layout.Box, property string,
	indices []int, initial []Paint, start time.Time) (*Renderer, error) {
	return BindTransitionTo(d, box, property, indices, initial, nil, start)
}
