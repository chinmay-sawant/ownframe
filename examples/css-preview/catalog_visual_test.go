package main

import (
	"fmt"
	"testing"
)

func TestCatalogCardsRenderDifferentMidframes(t *testing.T) {
	p, d, v := fixturePage(t)
	seen := map[string]string{}
	for _, row := range v.Catalog {
		startCatalog(t, p, d, row.Name)
		plan := d.samples[row.SampleKey].plan
		plan.step(plan.start.Add(plan.delay + plan.duration/2))
		signature := ""
		for _, pair := range plan.pairs {
			signature += fmt.Sprintf("%+v;", read(pair.op))
		}
		if other, exists := seen[signature]; exists {
			t.Errorf("%s and %s have the same visible midpoint", row.Name, other)
		}
		seen[signature] = row.Name
	}
	if len(seen) != 43 {
		t.Fatalf("unique catalog midpoints = %d, want 43", len(seen))
	}
}
