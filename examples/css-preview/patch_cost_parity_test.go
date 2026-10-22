package main

import (
	"testing"
)

// TestPatchElementMatchesAFullRedraw is the element patch parity gate. The
// patched display must hold the same operations for the card and for the
// state span as a full redraw of the same state, and every other operation
// must keep its place.
func TestPatchElementMatchesAFullRedraw(t *testing.T) {
	ctx, p, v := patchCostPage(t)
	card := patchCostCardOf(v)
	if !patchCostStateSpan(ctx, p, card, true) {
		t.Fatal("the state span patch was refused")
	}
	patchedCard := patchCostOps(p.Display(), card.id)
	if len(patchedCard) < 2 {
		t.Fatalf("the card owns %d operations, want the card's own paint", len(patchedCard))
	}
	patchedSpan := patchCostOps(p.Display(), card.state)
	patchedBox, ok := patchCostBox(p.Display(), card.state)
	if !ok {
		t.Fatal("the state span has no box")
	}
	patchedOther := patchCostOtherOps(p.Display(), card.id)

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	comparePatchOps(t, "card", patchedCard, patchCostOps(p.Display(), card.id))
	comparePatchOps(t, "state span", patchedSpan, patchCostOps(p.Display(), card.state))
	fullBox, ok := patchCostBox(p.Display(), card.state)
	if !ok {
		t.Fatal("the state span has no box after the redraw")
	}
	comparePatchBox(t, patchedBox, fullBox)
	comparePatchOps(t, "rest of the page", patchedOther,
		patchCostOtherOps(p.Display(), card.id))
}
