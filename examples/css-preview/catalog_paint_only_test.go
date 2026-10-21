package main

import (
	"context"
	"testing"
)

// TestClickPatchesWithoutRebuildingThePage checks that a click puts the
// click state into the parsed tree and paints it, without rebuilding the
// page. A rebuild parses the template again, so the parse count is the
// marker.
func TestClickPatchesWithoutRebuildingThePage(t *testing.T) {
	ctx := context.Background()
	p, d, v := fixturePage(t)
	row := v.Catalog[0]
	button, ok := findBox(p.Boxes(), row.Button)
	if !ok {
		t.Fatalf("missing button %q", row.Button)
	}
	parses := p.Stats().Parses
	if err := p.Click(ctx, button.X+button.W/2, button.Y+button.H/2); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Parses; got != parses {
		t.Fatalf("the click rebuilt the page: parses %d then %d", parses, got)
	}
	if d.treeDirty {
		t.Fatal("the tree carries state the engine refused to splice")
	}
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
}
