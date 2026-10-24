package main

import (
	"context"
	"testing"
)

// TestPlayAllKeepsTheCatalogWindowed checks that Play all starts the previews
// on screen and leaves the rest for their first render, so the catalog stays
// windowed and a click keeps its windowed cost.
func TestPlayAllKeepsTheCatalogWindowed(t *testing.T) {
	ctx := context.Background()
	p, d, v := fixturePage(t)
	m, ok := measureCatalog(p)
	if !ok {
		t.Fatal("could not measure the catalog")
	}
	if err := installCatalogWindow(ctx, p, v, d, m); err != nil {
		t.Fatal(err)
	}
	clickID(t, p, playAllID)
	tickPage(t, p)
	if len(v.Rows) >= len(v.Catalog) {
		t.Fatalf("the catalog rendered %d rows, want the windowed subset", len(v.Rows))
	}
	if got := p.Probe().Motion().Bound; got == 0 {
		t.Fatal("no preview started")
	}
	offscreen := 0
	for _, row := range v.Catalog[len(v.Rows):] {
		if _, found := findBox(p.Boxes(), row.TargetID); !found {
			offscreen++
		}
	}
	if offscreen == 0 {
		t.Fatal("no started preview waits for its row")
	}
	// The first window holds only animation cards. Scroll to the first
	// transition card and start it by hand, which is the path a transition
	// takes until the transition track lands.
	scrollTo(t, ctx, p, m, m.rowTop[4]-200)
	transition := catalogRowByName(t, v, "-webkit-transition")
	clickID(t, p, transition.Button)
	tickPage(t, p)
	if s := d.manual[transition.SampleKey]; s == nil || !s.active {
		t.Fatal("the visible transition did not start in the window")
	}
}
