package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

// benchmarkElementPatchCardClass measures the element patch of the class
// change on a card row. The engine refuses a card row today, so the run
// reports how many patches applied.
func benchmarkElementPatchCardClass(b *testing.B) {
	ctx, p, v := patchCostPage(b)
	card := patchCostCardOf(v)
	b.ReportAllocs()
	b.ResetTimer()
	applied := 0
	for i := 0; i < b.N; i++ {
		class := card.cardBase
		if i%2 == 0 {
			class = card.cardOn
		}
		if p.PatchElement(ctx, card.id, page.ElementEdit{Attribute: "class", Value: class}) {
			applied++
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(applied), "applied")
}

// benchmarkElementPatchStateClass measures the element patch of the class
// change on the state span of a card.
func benchmarkElementPatchStateClass(b *testing.B) {
	ctx, p, v := patchCostPage(b)
	card := patchCostCardOf(v)
	b.ReportAllocs()
	b.ResetTimer()
	applied := 0
	for i := 0; i < b.N; i++ {
		if patchCostStateSpan(ctx, p, card, i%2 == 0) {
			applied++
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(applied), "applied")
}

// patchCostStateSpan puts the state of a card into the state span of the
// card. The span carries the class and the text the card shows. It reports
// whether the engine spliced the span.
func patchCostStateSpan(ctx context.Context, p *ownframe.Page, card patchCostCard, on bool) bool {
	class, text := card.spanBase, "Ready"
	if on {
		class, text = card.spanOn, "Started / selected"
	}
	return p.PatchElement(ctx, card.state,
		page.ElementEdit{Attribute: "class", Value: class},
		page.ElementEdit{Text: &text})
}
