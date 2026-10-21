package main

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// TestPatchCardRowMatchesAFullRedraw is the same gate for the class change
// on a card row. The engine refuses a card row today, so the test takes the
// refusal and checks that the page keeps the operations it had.
func TestPatchCardRowMatchesAFullRedraw(t *testing.T) {
	ctx, p, v := patchCostPage(t)
	card := patchCostCardOf(v)
	before := patchCostOps(p.Display(), card.id)
	if !p.PatchElement(ctx, card.id, page.ElementEdit{Attribute: "class", Value: card.cardOn}) {
		comparePatchOps(t, "card", patchCostOps(p.Display(), card.id), before)
		return
	}
	patched := patchCostOps(p.Display(), card.id)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	comparePatchOps(t, "card", patched, patchCostOps(p.Display(), card.id))
}
