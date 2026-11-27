package main

import (
	"fmt"
	"testing"
	"time"
)

// TestCatalogCardsRenderDifferentMidframes starts every card, lets the
// automatic driver and the manual transition path run, and checks that each
// card paints its own frame.
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
	remaining := make(map[string]struct{}, len(v.Catalog))
	for _, row := range v.Catalog {
		remaining[row.Button] = struct{}{}
	}
	seen := map[string]string{}
	// Separate redraws give each card a different animation age. A looping
	// animation can be at its base pose at one sample, so keep sampling through
	// one second and retain each card's first changed frame.
	deadline := time.Now().Add(time.Second)
	for len(remaining) > 0 && time.Now().Before(deadline) {
		tickPage(t, p)
		for _, row := range v.Catalog {
			if _, pending := remaining[row.Button]; !pending {
				continue
			}
			paints := targetPaints(t, p, row.TargetID)
			if samePaint(base[row.SampleKey], paints) {
				continue
			}
			signature := ""
			for _, paint := range paints {
				signature += fmt.Sprintf("%+v;", paint)
			}
			if other, exists := seen[signature]; exists {
				t.Errorf("%s and %s have the same visible sample", row.Name, other)
			}
			seen[signature] = row.Name
			delete(remaining, row.Button)
		}
		if len(remaining) > 0 {
			time.Sleep(20 * time.Millisecond)
		}
	}
	for _, row := range v.Catalog {
		if _, pending := remaining[row.Button]; pending {
			t.Errorf("%s did not move off its declared paint", row.Name)
		}
	}
	if len(seen) != 43 {
		t.Fatalf("unique catalog motion samples = %d, want 43", len(seen))
	}
}
