package main

import (
	"context"
	"testing"
)

// TestPlayAllKeepsTheCatalogWindowed checks that the Play all button starts
// the previews on screen and leaves the rest waiting for their first render,
// so the catalog stays windowed and a click keeps its windowed cost.
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
	all, ok := findBox(p.Boxes(), playAllID)
	if !ok {
		t.Fatal("the play all button is missing")
	}
	if err := p.Click(ctx, all.X+all.W/2, all.Y+all.H/2); err != nil {
		t.Fatal(err)
	}
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
	if len(v.Rows) >= len(v.Catalog) {
		t.Fatalf("the catalog rendered %d rows, want the windowed subset", len(v.Rows))
	}

	active, waiting := 0, 0
	for _, s := range d.samples {
		if s.active {
			active++
		}
		if s.base == nil {
			waiting++
		}
	}
	if active == 0 {
		t.Fatal("no preview started")
	}
	if waiting == 0 {
		t.Fatal("no preview waits for its row")
	}
	if active+waiting < len(v.Catalog) {
		t.Fatalf("started %d and waiting %d, want the whole catalog", active, waiting)
	}
}
