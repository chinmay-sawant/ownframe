package board

import (
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/frame"
)

// bindBar binds the refresh fill in the livery accent.
func (f *flipper) bindBar(d *ownframe.Display, boxes []ownframe.Box, accent [3]float64) {
	f.bar, f.barFullW = nil, 0
	if d == nil {
		return
	}

	b, ok := boxByID(boxes, "refresh-fill")
	if !ok {
		return
	}

	f.bar = frame.Fill(d, b, accent)
	_, _, f.barFullW, _ = frame.BoxUnits(d, b)
}

// boxByID returns the first hit-test box with id.
func boxByID(boxes []ownframe.Box, id string) (ownframe.Box, bool) {
	for _, b := range boxes {
		if b.ID == id {
			return b, true
		}
	}

	return ownframe.Box{}, false
}
