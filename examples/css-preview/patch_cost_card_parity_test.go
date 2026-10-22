package main

import (
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// TestPatchCardRowMatchesAFullRedraw is the same gate for the class change
// on a card row. The engine splices the card row, so the test takes the
// patched operations, redraws the page, and compares the card operations of
// the patch with the card operations of the full redraw. A refused patch
// fails the test: a gate that passes without a patch measures nothing.
func TestPatchCardRowMatchesAFullRedraw(t *testing.T) {
	ctx, p, v := patchCostPage(t)
	card := patchCostCardOf(v)
	if err := p.PatchElement(ctx, card.id, ownframe.ElementEdit{Attribute: "class", Value: card.cardOn}); err != nil {
		t.Fatal(err)
	}
	patched := patchCostOps(p.Display(), card.id)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	comparePatchOps(t, "card", patched, patchCostOps(p.Display(), card.id))
}
