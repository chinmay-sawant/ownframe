package main

import (
	"context"
	"testing"
)

func TestEveryCatalogStartRunsOnlyItsOwnPreview(t *testing.T) {
	ctx := context.Background()
	p, d, v := fixturePage(t)
	started := 0
	for _, row := range v.Catalog {
		button, ok := findBox(p.Boxes(), row.Button)
		if !ok {
			t.Fatalf("missing Start button for %s", row.Name)
		}
		if err := p.Click(ctx, button.X+button.W/2, button.Y+button.H/2); err != nil {
			t.Fatal(err)
		}
		if !v.Started[row.Button] || v.Selected != row.Name {
			t.Fatalf("%s was not selected", row.Name)
		}
		if err := d.tick(ctx); err != nil {
			t.Fatal(err)
		}
		sample := d.samples[row.SampleKey]
		if sample == nil || len(sample.plan.pairs) == 0 {
			t.Fatalf("%s did not start its own preview", row.Name)
		}
		started++
	}
	if started != 43 || len(d.samples) != 43 {
		t.Fatalf("started previews = %d, want 43", len(d.samples))
	}
}
