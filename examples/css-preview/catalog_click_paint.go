package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// clickPaint puts the click state into the parsed tree. It reports that the
// frame is painted, so the page skips the layout the click does not need.
// The card that loses the selection gives its class back the same way.
func (d *demo) clickPaint(ctx context.Context, b ownframe.Box) (bool, error) {
	row, ok := d.rowForButton(b.ID)
	if !ok {
		return false, nil
	}
	painted := d.patchCard(ctx, row)
	if d.prevSelected != "" && d.prevSelected != row.Name {
		if old, found := d.rowForName(d.prevSelected); found {
			d.patchCard(ctx, old)
		}
	}
	if !painted || d.dataChanged {
		// The readout is template data and changed, so the page has to
		// draw again. It rebuilds the tree, and the after-redraw hook
		// puts the state back.
		return false, nil
	}
	if !walkActive() {
		// The tree keeps its state and no draw follows, so the tick lays
		// the page out with the class the engine refused to splice.
		d.p.SetTick(d.tick)
	}
	return true, nil
}
