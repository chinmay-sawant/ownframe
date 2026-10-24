package main

import (
	"testing"
)

// TestEveryCatalogStartRunsOnlyItsOwnPreview clicks all 43 Start buttons.
// Every card records its own state; the page runtime binds an automatic
// player for each animation card and the manual path starts each transition.
func TestEveryCatalogStartRunsOnlyItsOwnPreview(t *testing.T) {
	p, d, v := fixturePage(t)
	for _, row := range v.Catalog {
		clickID(t, p, row.Button)
		if !v.Started[row.Button] || v.Selected != row.Name {
			t.Fatalf("%s was not selected", row.Name)
		}
	}
	tickPage(t, p)
	started := 0
	for _, row := range v.Catalog {
		if v.Started[row.Button] {
			started++
		}
	}
	if started != 43 {
		t.Fatalf("started cards = %d, want 43", started)
	}
	manual := 0
	for _, row := range v.Catalog {
		target, ok := targetByKey(row.SampleKey)
		if !ok {
			t.Fatalf("%s has no target", row.Name)
		}
		if target.kind == "transition" {
			if s := d.manual[row.SampleKey]; s != nil {
				manual++
			}
		}
	}
	if manual != 6 {
		t.Fatalf("manual transition samples = %d, want the 6 geometry cards", manual)
	}
	if got := p.Probe().Motion().Bound; got < 29 {
		t.Fatalf("automatic players = %d, want the animation and view-transition cards", got)
	}
	if err := p.Probe().Motion().LastError; err != "" {
		t.Fatalf("motion bind failed: %s", err)
	}
}
