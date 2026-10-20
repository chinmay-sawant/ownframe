package csstransition

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/transition"
)

// Bind builds a transition renderer for one CSS property on a box. The
// destination comes from the display, so a caller that rebinds later wants
// BindTo with the paints it captured on the first bind.
func Bind(d *ownframe.Display, box ownframe.Box, property string,
	indices []int, initial []cssmotion.Paint, start time.Time) (*transition.Renderer, error) {
	return BindTo(d, box, property, indices, initial, nil, start)
}
