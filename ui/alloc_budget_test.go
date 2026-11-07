package ui

import (
	"context"
	"strconv"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// Baseline (Nov 2026, 1000-row invoice page): a no-op Update re-click costs
// ~2.07MB per click, dominated by rebuilding the 1000-row builder tree and
// re-serializing the source snapshot, while a real change costs ~7.5MB
// (rebuild + ~369KB serialize + parse + layout). Goal: <100KB per click.
//
// This fixture is 20 rows, not 1000, so it runs in milliseconds and pins the
// shape of the cost, not the byte count: the no-op path must do no page work
// at all (no parse, no layout). Expected direction as fixes land:
//   - tag interning lowers the per-op tree rebuild (repeated tag and
//     attribute strings stop allocating once per node);
//   - source sync keeps the no-op early-out (the diff finds zero patches, so
//     no snapshot is serialized and the page is untouched).
// The benchmark's B/op should fall toward the tree-build floor; the test's
// Parses/Layouts deltas must stay zero throughout.

const (
	noopBudgetRows = 20
	noopBudgetCols = 3
	// noopBudgetSelected is the one row carrying an OnClick handler, so the
	// fixture covers callback-registry adoption on the no-op path.
	noopBudgetSelected = 3
)

// noopBudgetTree builds the small stable-id table the no-op budget pins: a
// header row plus noopBudgetRows body rows, every cell addressable by id.
func noopBudgetTree(selected int) *Element {
	body := Tag("tbody").ID("noop-body")
	for i := range noopBudgetRows {
		id := "noop-row-" + strconv.Itoa(i)
		row := Tag("tr").ID(id)
		if i == selected {
			row.OnClick(func(context.Context) error { return nil })
		}
		for c := range noopBudgetCols {
			row.Children(Tag("td").ID(id + "-c" + strconv.Itoa(c)).Text("r" + strconv.Itoa(i) + "c" + strconv.Itoa(c)))
		}
		body.Children(row)
	}
	head := Tag("tr").ID("noop-head-row")
	for c := range noopBudgetCols {
		head.Children(Tag("th").ID("noop-h" + strconv.Itoa(c)).Text("col " + strconv.Itoa(c)))
	}
	return Tag("table").ID("noop-table").Children(
		Tag("thead").ID("noop-head").Children(head),
		body,
	)
}

// BenchmarkUpdateNoopBudget pins the steady-state cost of re-clicking with
// unchanged data: each op rebuilds an identical 20-row tree and runs
// View.Update, which must take the no-op early-out (zero patches, page
// untouched). Run with -benchmem: ReportAllocs surfaces allocs/op and B/op.
// B/op today is the tree rebuild plus the diff walk; tag interning is the
// fix expected to move it, since the page itself does no work on this path.
func BenchmarkUpdateNoopBudget(b *testing.B) {
	ctx := context.Background()
	view, err := Mount(ownframe.Config{Width: testWidth, Height: testHeight}, noopBudgetTree(noopBudgetSelected))
	if err != nil {
		b.Fatal(err)
	}
	if err := view.Page.Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if err := view.Update(ctx, noopBudgetTree(noopBudgetSelected)); err != nil {
			b.Fatal(err)
		}
	}
}

// TestUpdateNoopBudget pins the framework no-op guarantee: one Update with
// unchanged data returns nil and grows neither Parses nor Layouts, so a
// re-click never reparses or relayouts the page. Source sync must preserve
// this early-out while it removes the serialize-plus-parse work from the
// real-change path.
func TestUpdateNoopBudget(t *testing.T) {
	ctx := context.Background()
	view := mountView(t, ownframe.Config{}, noopBudgetTree(noopBudgetSelected))
	before := view.Page.Probe().Stats()
	if err := view.Update(ctx, noopBudgetTree(noopBudgetSelected)); err != nil {
		t.Fatalf("Update with unchanged data = %v, want nil", err)
	}
	after := view.Page.Probe().Stats()
	if after.Parses != before.Parses {
		t.Fatalf("parses = %d, want %d: a no-op Update reparsed", after.Parses, before.Parses)
	}
	if after.Layouts != before.Layouts {
		t.Fatalf("layouts = %d, want %d: a no-op Update relaid out", after.Layouts, before.Layouts)
	}
	allocs := testing.AllocsPerRun(10, func() {
		if err := view.Update(ctx, noopBudgetTree(noopBudgetSelected)); err != nil {
			t.Error(err)
		}
	})
	t.Logf("no-op Update allocs/op: %v (tree rebuild plus diff walk; the page does no work)", allocs)
}
