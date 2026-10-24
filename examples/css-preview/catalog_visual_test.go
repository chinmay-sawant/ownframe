package main

import (
	"fmt"
	"testing"
	"time"
)

// TestCatalogCardsRenderDifferentMidframes starts every card, lets the
// automatic driver and the manual transition path take one sample, and checks
// that each card paints its own frame.
func TestCatalogCardsRenderDifferentMidframes(t *testing.T) {
	p, _, v := fixturePage(t)
	// Settle the paused cards before the comparison.
	tickPage(t, p)
	base := map[string][]paint{}
	for _, item := range targets() {
		base[item.key] = targetPaints(t, p, item.id)
	}
	for _, row := range v.Catalog {
		clickID(t, p, row.Button)
	}
	time.Sleep(600 * time.Millisecond)
	tickPage(t, p)
	seen := map[string]string{}
	for _, row := range v.Catalog {
		paints := targetPaints(t, p, row.TargetID)
		signature := ""
		for _, paint := range paints {
			signature += fmt.Sprintf("%+v;", paint)
		}
		if other, exists := seen[signature]; exists {
			t.Errorf("%s and %s have the same visible midpoint", row.Name, other)
		}
		seen[signature] = row.Name
		if samePaint(base[row.SampleKey], paints) {
			t.Errorf("%s did not move off its declared paint", row.Name)
		}
	}
	if len(seen) != 43 {
		t.Fatalf("unique catalog midpoints = %d, want 43", len(seen))
	}
}
