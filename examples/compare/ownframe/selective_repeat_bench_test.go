package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

// BenchmarkSelectiveRepeatClick pins the per-repeat cost of re-clicking the
// already-selected row on the selective path.
//
// Expected shape: the first ClickAction records fingerprint "sel=100" in the
// applied-fingerprint table; every measured repeat re-sets selected=100, so
// View.UpdateChanged sees a fingerprint map hit and returns nil before build,
// render, diff, or page work. B/op should be single-digit KB or less,
// dominated by the test loop itself (context + ClickAction dispatch), not by
// the view or page.
func BenchmarkSelectiveRepeatClick(b *testing.B) {
	ctx := context.Background()
	a, err := newListApp()
	if err != nil {
		b.Fatal(err)
	}
	a.view.Handle(ownframe.Handlers{Click: a.handleClickSelective})
	if err := a.Page().Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	// Prime the selection to row:100 via the selective handler so the
	// fingerprint table holds "sel=100" before the clock starts; every timed
	// repeat below must hit that fast path.
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

// TestSelectiveRepeatIsNoop pins the zero-page-touch guarantee for a repeated
// identical selection: one repeat click returns nil and moves neither Parses
// nor Layouts, proving the page is untouched (fingerprint hit returns before
// build, so no snapshot is serialized and nothing reparses or relayouts).
func TestSelectiveRepeatIsNoop(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newListApp()
	if err != nil {
		t.Fatal(err)
	}
	a.view.Handle(ownframe.Handlers{Click: a.handleClickSelective})
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
		t.Fatalf("parses = %d, want %d: a repeat selective click reparsed", after.Parses, before.Parses)
	}
	if after.Layouts != before.Layouts {
		t.Fatalf("layouts = %d, want %d: a repeat selective click relaid out", after.Layouts, before.Layouts)
	}
}
