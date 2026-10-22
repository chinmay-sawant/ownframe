package main

import (
	"context"
)

// syncTree puts the app's state back into the parsed tree after the
// template replaced it. It reports whether the page has to draw again. The
// tree is fresh when this runs, so the started cards and the selected card
// are the only ones that need their state back.
func (d *demo) syncTree(ctx context.Context) bool {
	redraw := false
	if d.readout != "" {
		// The readout is in the tree, so the rebuilt tree carries the
		// initial snapshot again until the readout is patched back.
		if !d.patchReadout(ctx, d.readout) {
			redraw = true
		}
	}
	if !d.patchGrid(ctx) {
		redraw = true
	}

	return redraw
}

// afterRedraw puts the app's state back into the parsed tree when the
// template replaced it, and asks for the layout that shows it. The page
// calls it after every redraw, so a scroll or a reload does not lose the
// started state of the other cards.
func (d *demo) afterRedraw(ctx context.Context) {
	parses := d.p.Stats().Parses
	if parses == d.treeParses {
		return
	}
	d.treeParses = parses
	d.treeDirty = false
	// The rows that just rendered hold their starting paint until the
	// state patch moves them.
	d.captureBases()
	if !d.syncTree(ctx) {
		return
	}
	if err := d.p.Redraw(ctx); err != nil {
		// The state patch did not display; the next tick lays it out.
		d.treeDirty = true
	}
}
