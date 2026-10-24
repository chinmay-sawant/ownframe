package main

import (
	"strings"
	"testing"
)

// TestClickRendersTheCardState checks that a click puts the card's state into
// the template data: the state span carries the text and the class, and the
// row's box did not move.
func TestClickRendersTheCardState(t *testing.T) {
	p, _, v := fixturePage(t)
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
	clickID(t, p, row.Button)
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
	// The selected marker belongs to the class the template renders.
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
}
