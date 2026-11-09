package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/pagetest"
)

// BenchmarkRepeatSelectionClick pins the per-repeat cost of re-clicking the
// already-selected row through the app handler.
//
// Expected shape: handleClick sets selected to the value it already holds and
// calls View.Update. Update rebuilds the 1000-row tree and diffs it clean, so
// it adopts the fresh callback registry and returns nil before a page load or
// patch. B/op reflects that rebuild and diff, not page work or a snapshot; the
// zero-page-touch guarantee is asserted by TestRepeatedSelectionIsNoop.
func BenchmarkRepeatSelectionClick(b *testing.B) {
	ctx := context.Background()
	a, err := newListApp()
	if err != nil {
		b.Fatal(err)
	}
	if err := a.Page().Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	// Prime the selection to row:100 so every timed repeat below is an
	// unchanged selection.
	if err := pagetest.ClickAction(ctx, a.Page(), "row:100"); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := pagetest.ClickAction(ctx, a.Page(), "row:100"); err != nil {
			b.Fatal(err)
		}
	}
}

// TestRepeatedSelectionIsNoop pins the zero-page-touch guarantee for a repeated
// identical selection: one repeat click returns nil and moves neither Parses
// nor Layouts, proving the page is untouched (the clean diff returns before a
// load, so no snapshot is serialized and nothing reparses or relayouts).
func TestRepeatedSelectionIsNoop(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newListApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.Page(), "row:100"); err != nil {
		t.Fatal(err)
	}
	before := a.page.Probe().Stats()
	if err := pagetest.ClickAction(ctx, a.Page(), "row:100"); err != nil {
		t.Fatalf("repeat ClickAction with unchanged selection = %v, want nil", err)
	}
	after := a.page.Probe().Stats()
	if after.Parses != before.Parses {
		t.Fatalf("parses = %d, want %d: a repeat click reparsed", after.Parses, before.Parses)
	}
	if after.Layouts != before.Layouts {
		t.Fatalf("layouts = %d, want %d: a repeat click relaid out", after.Layouts, before.Layouts)
	}
}
