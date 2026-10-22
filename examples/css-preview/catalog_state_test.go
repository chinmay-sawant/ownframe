package main

import (
	"context"
	"strings"
	"testing"
)

// TestClickPatchesTheCardInTheTree checks that a click puts the card's
// click state into the parsed tree: the state span carries the text and the
// class, and the row's box did not move.
func TestClickPatchesTheCardInTheTree(t *testing.T) {
	ctx := context.Background()
	p, d, v := fixturePage(t)
	row := v.Catalog[0]
	card, ok := findBox(p.Boxes(), row.CardID)
	if !ok {
		t.Fatalf("missing card %q", row.CardID)
	}
	span, ok := findBox(p.Boxes(), row.StateID)
	if !ok {
		t.Fatalf("missing state span %q", row.StateID)
	}
	before := capture(p.Probe().Display(), span)
	button, ok := findBox(p.Boxes(), row.Button)
	if !ok {
		t.Fatalf("missing button %q", row.Button)
	}
	if err := p.Click(ctx, button.X+button.W/2, button.Y+button.H/2); err != nil {
		t.Fatal(err)
	}
	span, ok = findBox(p.Boxes(), row.StateID)
	if !ok {
		t.Fatalf("missing state span %q", row.StateID)
	}
	if !strings.Contains(span.Text, "Started / selected") {
		t.Fatalf("the state span reads %q", span.Text)
	}
	after := capture(p.Probe().Display(), span)
	if len(before) == 0 || len(after) == 0 {
		t.Fatal("the state span paints nothing")
	}
	// The class only shows when the tree carries it: the marker colour
	// belongs to the rule that matches the patched span.
	if !nearColor(after[0], 0x2f, 0x6f, 0x5f) {
		t.Fatalf("the state span paints %+v, want the selected marker", after[0])
	}
	moved, ok := findBox(p.Boxes(), row.CardID)
	if !ok {
		t.Fatalf("missing card %q", row.CardID)
	}
	if moved.X != card.X || moved.Y != card.Y || moved.W != card.W || moved.H != card.H {
		t.Fatalf("the row moved: %+v then %+v", card, moved)
	}
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
}
