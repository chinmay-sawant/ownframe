package main

import (
	"context"
	"strings"
	"testing"
)

// TestRebuildKeepsTheStartedState checks that a rebuild puts the app's
// state back into the parsed tree, so no card loses its started state when
// the click fell back to the full page path.
func TestRebuildKeepsTheStartedState(t *testing.T) {
	ctx := context.Background()
	p, _, v := fixturePage(t)
	rows := []catalogRow{v.Catalog[0], v.Catalog[1]}
	for _, row := range rows {
		button, ok := findBox(p.Boxes(), row.Button)
		if !ok {
			t.Fatalf("missing button %q", row.Button)
		}
		if err := p.Click(ctx, button.X+button.W/2, button.Y+button.H/2); err != nil {
			t.Fatal(err)
		}
	}
	// A rebuild replaces the parsed tree.
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		span, ok := findBox(p.Boxes(), row.StateID)
		if !ok {
			t.Fatalf("missing state span %q", row.StateID)
		}
		if !strings.Contains(span.Text, "Started / selected") {
			t.Fatalf("the state span of %q reads %q", row.Name, span.Text)
		}
	}
}
