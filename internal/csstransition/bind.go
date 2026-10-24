package csstransition

import (
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/transition"
)

// Bind builds a transition renderer for one CSS property on a box. The
// destination comes from the display, so a caller that rebinds later wants
// BindTo with the paints it captured on the first bind. Bind trusts the
// captured indices, because the paint it writes as the destination is the
// one the display holds. BindTo checks them, because it writes a paint the
// caller captured on an earlier display.
func Bind(d *layout.Display, box layout.Box, property string,
	indices []int, initial []cssmotion.Paint, start time.Time) (*transition.Renderer, error) {
	return BindTo(d, box, property, indices, initial, nil, start)
}
