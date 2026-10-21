package main

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// BenchmarkPatchCost measures a full redraw and an element patch of the same
// change on the real CSS preview fixture at the desktop window size.
//
// The page API has no other way to change the tree, so both paths make the
// same tree edit through PatchElement. The full path redraws the page after
// it, and the patch path stops there. The difference between the two is the
// cost of the redraw.
func BenchmarkPatchCost(b *testing.B) {
	b.Run("FullRedrawCardClass", benchmarkFullRedrawCardClass)
	b.Run("ElementPatchCardClass", benchmarkElementPatchCardClass)
	b.Run("FullRedrawStateClass", benchmarkFullRedrawStateClass)
	b.Run("ElementPatchStateClass", benchmarkElementPatchStateClass)
	b.Run("FullRedrawSameSource", benchmarkFullRedrawSameSource)
}

// benchmarkFullRedrawCardClass measures the full path for the class change on
// a card row: the tree edit, then a redraw of the whole page.
func benchmarkFullRedrawCardClass(b *testing.B) {
	ctx, p, v := patchCostPage(b)
	card := patchCostCardOf(v)
	b.ReportAllocs()
	b.ResetTimer()
	var template, cascade, layout time.Duration
	for i := 0; i < b.N; i++ {
		class := card.cardBase
		if i%2 == 0 {
			class = card.cardOn
		}
		p.PatchElement(ctx, card.id, page.ElementEdit{Attribute: "class", Value: class})
		p.SetData(v)
		if err := p.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
		template, cascade, layout = addPatchStages(p, template, cascade, layout)
	}
	b.StopTimer()
	reportPatchStages(b, template, cascade, layout)
}
