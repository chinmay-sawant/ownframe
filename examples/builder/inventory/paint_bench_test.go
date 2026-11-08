package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// TestInventoryHoverStaysPatched pins the committed hover guarantee: sweeping
// the bin grid patches bin-detail in place without a reparse.
func TestInventoryHoverStaysPatched(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := a.page.Probe().Stats().Parses
	prev := ownframe.Box{}
	for _, row := range sampleBins {
		for aisle := range row.Cells {
			next := ownframe.Box{ID: "bin-" + aisleLetter(aisle) + "-" + row.Level}
			if _, err := a.handleHover(ctx, prev, next); err != nil {
				t.Fatal(err)
			}
			prev = next
		}
	}
	// Hovering off the grid restores the summary line, still via patch.
	if _, err := a.handleHover(ctx, prev, ownframe.Box{}); err != nil {
		t.Fatal(err)
	}
	// Workstream A (patch-friendly tree: stable ids, symmetric styles) moves
	// this delta; workstream B (fingerprint-gated refresh) must leave it at
	// 0 because the hover path never reaches refresh on a patch hit.
	if after := a.page.Probe().Stats().Parses; after != before {
		t.Fatalf("parses = %d, want %d: a hover sweep reparsed", after, before)
	}
}

// TestInventoryRepeatClickIsFree pins the repeat-click no-op at the app
// level: re-clicking the active status pill changes no observable HTML.
func TestInventoryRepeatClickIsFree(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "status:Stockout"}); err != nil {
		t.Fatal(err)
	}
	before := a.page.HTML()
	if err := a.handleClick(ctx, ownframe.Box{Action: "status:Stockout"}); err != nil {
		t.Fatal(err)
	}
	// Workstream B (fingerprint-gated selective refresh) moves the cost of
	// this repeat from a full refresh to an early return or fingerprint hit;
	// this test asserts the outcome (identical HTML) so it passes under both
	// paths without depending on the exact no-op mechanism.
	if after := a.page.HTML(); after != before {
		t.Fatalf("repeat status click changed the page HTML (%d vs %d bytes)", len(after), len(before))
	}
}

// BenchmarkInventoryPillSwitch cycles the four status filters, one switch per
// iteration, to measure the structural filter cost.
//
// Expected shape: a filter switch changes the table rows, so each switch is
// a full render by design. This bench measures the cost to beat, not a
// budget to fail: there is deliberately no byte assertion, so legitimate
// structural renders never fail it. Workstream A (stable row ids) is what a
// future incremental path would patch; workstream B skips only identical
// repeats, so every iteration here stays a real switch.
func BenchmarkInventoryPillSwitch(b *testing.B) {
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		b.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	filters := []string{"status:", "status:Stockout", "status:Reorder", "status:Overstock"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := a.handleClick(ctx, ownframe.Box{Action: filters[i%len(filters)]}); err != nil {
			b.Fatal(err)
		}
	}
}
