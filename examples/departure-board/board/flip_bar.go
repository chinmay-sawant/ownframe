package board

import (
	"github.com/chinmay-sawant/ownframe"
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

	f.bar = ownframe.Fill(d, b, accent)
	_, _, f.barFullW, _ = ownframe.BoxUnits(d, b)
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
