package csstransition

import (
	"errors"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/transition"
)

// BindTo builds a transition renderer with an explicit destination paint.
// A nil final reads the destination from the display, which is what a first
// bind wants. A caller that rebinds later passes the paints it captured on
// the first bind, so a redraw in the middle of the transition does not make
// the destination the value the transition had reached.
//
// The captured paints describe the operations they were read from. A rebind
// checks that the operations are still the ones the box owns, and a field
// the layout changed keeps the operation's current value. See bind_verify.go
// and bind_paint.go.
func BindTo(d *layout.Display, box layout.Box, property string,
	indices []int, initial, final []cssmotion.Paint, start time.Time) (*transition.Renderer, error) {
	if d == nil {
		return nil, errors.New("csstransition: display is nil")
	}
	record, ok := transitionRecordOf(box, property)
	if !ok {
		return nil, errors.New("csstransition: box has no matching transition metadata")
	}
	ops := cssmotion.Resolve(d, indices)
	if len(ops) != len(indices) {
		return nil, errors.New("csstransition: captured operations are missing from display")
	}
	if len(ops) == 0 || len(initial) != len(ops) {
		return nil, errors.New("csstransition: target operation paints do not match")
	}
	if final != nil && len(final) != len(ops) {
		return nil, errors.New("csstransition: destination paints do not match")
	}
	if final != nil {
		if err := ownedByBox(d, box, ops); err != nil {
			return nil, err
		}
	}
	renderer, err := transition.New(buildTracks(ops, initial, final, record, property))
	if err != nil {
		return nil, err
	}
	renderer.Start(start)
	return renderer, nil
}
