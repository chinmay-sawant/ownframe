package main

import (
	"testing"
	"time"
)

// benchmarkFullRedrawStateClass measures the full path for the class change
// on the state span of a card: the tree edit, then a redraw of the whole
// page.
func benchmarkFullRedrawStateClass(b *testing.B) {
	ctx, p, v := patchCostPage(b)
	card := patchCostCardOf(v)
	b.ReportAllocs()
	b.ResetTimer()
	var template, cascade, layout time.Duration
	for i := 0; i < b.N; i++ {
		patchCostStateSpan(ctx, p, card, i%2 == 0)
		p.SetData(v)
		if err := p.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
		template, cascade, layout = addPatchStages(p, template, cascade, layout)
	}
	b.StopTimer()
	reportPatchStages(b, template, cascade, layout)
}

// benchmarkFullRedrawSameSource measures a redraw whose source did not
// change: the page re-cascades and lays the whole page out again, and skips
// the parse.
func benchmarkFullRedrawSameSource(b *testing.B) {
	ctx, p, _ := patchCostPage(b)
	b.ReportAllocs()
	b.ResetTimer()
	var template, cascade, layout time.Duration
	for i := 0; i < b.N; i++ {
		if err := p.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
		template, cascade, layout = addPatchStages(p, template, cascade, layout)
	}
	b.StopTimer()
	reportPatchStages(b, template, cascade, layout)
}

// reportPatchStages reports the mean stage times of the measured redraws.
func reportPatchStages(b *testing.B, template, cascade, layout time.Duration) {
	if b.N == 0 {
		return
	}
	b.ReportMetric(float64(template.Microseconds())/float64(b.N)/1000, "template-ms")
	b.ReportMetric(float64(cascade.Microseconds())/float64(b.N)/1000, "cascade-ms")
	b.ReportMetric(float64(layout.Microseconds())/float64(b.N)/1000, "layout-ms")
}
