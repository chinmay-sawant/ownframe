package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// patchCard puts one card's click state into the parsed tree and lays the
// card out again. The card carries the classes the sample needs, and the
// state span inside it carries the text and the marker. One patch covers
// both, because the card's patch lays the whole subtree out again.
func (d *demo) patchCard(ctx context.Context, row catalogRow) bool {
	text := d.spanText(row)
	painted := d.p.PatchElement(ctx, row.CardID,
		ownframe.ElementEdit{Attribute: "class", Value: d.cardClass(row)},
		ownframe.ElementEdit{ID: row.StateID, Attribute: "class", Value: d.spanClass(row)},
		ownframe.ElementEdit{ID: row.StateID, Text: &text})
	if !painted {
		// The engine refused the card, so the classes land in the tree and
		// the next layout shows them.
		d.treeDirty = true
	}

	return painted
}

// patchReadout puts the readout into the parsed tree. The readout scrolls in
// a fixed box, so its text changes without moving the box that holds it.
func (d *demo) patchReadout(ctx context.Context, text string) bool {
	return d.p.PatchElement(ctx, readoutID, ownframe.ElementEdit{Text: &text})
}
