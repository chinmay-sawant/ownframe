package frame

import "github.com/chinmay-sawant/ownframe/internal/page"

// Text returns the first text operation inside box, or nil. A caller can
// change its Text field to show a new value on the next frame.
func Text(d *page.Display, box page.Box) *page.DisplayOp {
	if d == nil {
		return nil
	}

	x, y, w, h := BoxUnits(d, box)

	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind != page.DisplayOpText {
			continue
		}

		// A text op carries its baseline in Y, not the line-box top, so test
		// the baseline against the box instead of the center fills use.
		if op.X >= x-0.5 && op.X <= x+w+0.5 && op.Y >= y-0.5 && op.Y <= y+h+0.5 {
			return op
		}
	}

	return nil
}
