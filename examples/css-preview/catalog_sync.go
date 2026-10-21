package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// syncTree puts the app's state back into the parsed tree after the
// template replaced it. It reports whether the tree changed. The tree is
// fresh when this runs, so the started cards and the selected card are the
// only ones that need their state back.
func (d *demo) syncTree(ctx context.Context) bool {
	changed := false
	if d.readout != "" {
		// The readout is in the tree, so the rebuilt tree carries the
		// initial snapshot again until the readout is patched back.
		changed = d.patchReadout(ctx, d.readout) || changed
	}
	for _, row := range d.v.Catalog {
		if !d.v.Started[row.Button] && d.v.Selected != row.Name {
			continue
		}
		text := d.spanText(row)
		span := d.spanClass(row)
		if d.p.PatchElement(ctx, row.StateID,
			ownframe.ElementEdit{Attribute: "class", Value: span},
			ownframe.ElementEdit{Text: &text}) {
			changed = true
		}
		// The card's class moves the transition target. The engine refuses
		// a card, so the edit lands in the tree and the next layout shows
		// it.
		d.p.PatchElement(ctx, row.CardID,
			ownframe.ElementEdit{Attribute: "class", Value: d.cardClass(row)})
	}
	return changed
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
	if !d.syncTree(ctx) {
		return
	}
	d.treeDirty = false
	d.p.Redraw(ctx)
}
