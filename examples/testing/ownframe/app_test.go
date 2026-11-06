package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestInvoiceShowsRows(t *testing.T) {
	t.Parallel()
	a, err := newInvoiceApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	html := a.Page().HTML()
	for _, want := range []string{"Invoice Table", invoiceSubtitle, "INV-1001", "INV-2000", "invoice-row-500", "Grand total", "Num", "Invoice", "Client", "Status", "invoice-theme", "inv-alt"} {
		if !strings.Contains(html, want) {
			t.Errorf("invoice HTML missing %q", want)
		}
	}
}

func TestInvoiceSelectRow(t *testing.T) {
	t.Parallel()
	a, err := newInvoiceApp()
	if err != nil {
		t.Fatal(err)
	}
	a.patch = false
	ctx := context.Background()
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.Page(), "row:500"); err != nil {
		t.Fatal(err)
	}
	if a.selected != 500 {
		t.Fatalf("selected = %d, want 500", a.selected)
	}
	html := a.Page().HTML()
	if !strings.Contains(html, "Selected: INV-1501") {
		t.Errorf("invoice HTML missing selection readout")
	} else if !strings.Contains(html, "inv-sel") {
		t.Errorf("invoice HTML missing selected-row class")
	}
	// A clean refresh after the committed selection must skip the rebuild,
	// and a real change after that skip must still rebuild.
	committed := a.Page().Probe().Stats()
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	skipped := a.Page().Probe().Stats()
	if skipped.Parses != committed.Parses || skipped.Cascades != committed.Cascades || skipped.Layouts != committed.Layouts {
		t.Errorf("clean refresh after select rebuilt: parses %d->%d cascades %d->%d layouts %d->%d, want flat",
			committed.Parses, skipped.Parses, committed.Cascades, skipped.Cascades, committed.Layouts, skipped.Layouts)
	}
	if err := pagetest.ClickAction(ctx, a.Page(), "row:501"); err != nil {
		t.Fatal(err)
	}
	if a.selected != 501 {
		t.Fatalf("selected = %d, want 501", a.selected)
	}
	rebuilt := a.Page().Probe().Stats()
	if rebuilt.Layouts <= skipped.Layouts {
		t.Errorf("select after skipped refresh laid nothing out: layouts %d->%d, want a rebuild",
			skipped.Layouts, rebuilt.Layouts)
	}
}

func TestInvoiceThemeToggle(t *testing.T) {
	t.Parallel()
	a, err := newInvoiceApp()
	if err != nil {
		t.Fatal(err)
	}
	a.patch = false
	ctx := context.Background()
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if a.theme != invoiceThemeGray {
		t.Fatalf("theme = %q, want gray", a.theme)
	}
	if err := pagetest.ClickAction(ctx, a.Page(), "theme"); err != nil {
		t.Fatal(err)
	}
	if a.theme != invoiceThemeLight {
		t.Fatalf("theme = %q, want light", a.theme)
	}
	html := a.Page().HTML()
	if !strings.Contains(html, "Theme: light") {
		t.Errorf("invoice HTML missing light theme label")
	}
	if strings.Contains(html, "inv-alt") {
		t.Errorf("light theme should have no inv-alt rows")
	}
}

// TestInvoicePatchDefault checks the patch path is on by default: a fresh
// app patches instead of rebuilding, so clicks never rewrite HTML source.
func TestInvoicePatchDefault(t *testing.T) {
	t.Parallel()
	a, err := newInvoiceApp()
	if err != nil {
		t.Fatal(err)
	}
	if !a.patch {
		t.Errorf("patch = false, want true by default")
	}
}

// rowFill returns the fill-rect color the row's own subtree painted, matched
// through the box's operation IDs. It proves the class edit changed the
// paint, not just the tree.
func rowFill(t *testing.T, a *invoiceApp, action string) (float64, float64, float64) {
	t.Helper()
	box, ok := pagetest.BoxByAction(a.Page(), action)
	if !ok {
		t.Fatalf("no box for action %q", action)
	}
	owned := make(map[uint64]bool, len(box.OperationIDs))
	for _, id := range box.OperationIDs {
		owned[id] = true
	}
	d := a.Page().Probe().Display()
	if d == nil {
		t.Fatal("no display list")
	}
	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind == ownframe.DisplayOpFillRect && owned[op.ID] {
			return op.R, op.G, op.B
		}
	}
	t.Fatalf("no fill op for action %q", action)
	return 0, 0, 0
}

// wantFill fails when the painted fill differs from the want triple (0..1)
// by more than a rounding tolerance on any channel.
func wantFill(t *testing.T, what string, r, g, b float64, want [3]float64) {
	t.Helper()
	const tolerance = 0.02
	diff := func(v, w float64) float64 {
		if v > w {
			return v - w
		}
		return w - v
	}
	if diff(r, want[0]) > tolerance || diff(g, want[1]) > tolerance || diff(b, want[2]) > tolerance {
		t.Errorf("%s fill = (%.3f,%.3f,%.3f), want %v", what, r, g, b, want)
	}
}

var (
	fillWhite = [3]float64{1, 1, 1}
	fillGray  = [3]float64{0.953, 0.957, 0.965}
	fillBlue  = [3]float64{0.859, 0.918, 0.996}
)

// TestInvoiceSelectPatch checks the fine-grained select path: two class
// edits plus a text edit, with no reparse or recascade, and the blue
// highlight painted on the newly selected visible row.
func TestInvoiceSelectPatch(t *testing.T) {
	t.Parallel()
	a, err := newInvoiceApp()
	if err != nil {
		t.Fatal(err)
	}
	a.patch = true
	ctx := context.Background()
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	r, g, b := rowFill(t, a, "row:5")
	wantFill(t, "row 5 before select", r, g, b, fillGray)
	before := a.Page().Probe().Stats()
	if err := pagetest.ClickAction(ctx, a.Page(), "row:5"); err != nil {
		t.Fatal(err)
	}
	if a.selected != 5 {
		t.Fatalf("selected = %d, want 5", a.selected)
	}
	after := a.Page().Probe().Stats()
	if after.Parses != before.Parses || after.Cascades != before.Cascades {
		t.Errorf("patch select reparsed: parses %d->%d cascades %d->%d",
			before.Parses, after.Parses, before.Cascades, after.Cascades)
	}
	if after.Layouts <= before.Layouts {
		t.Errorf("patch select laid nothing out")
	}
	r, g, b = rowFill(t, a, "row:5")
	wantFill(t, "row 5 selected", r, g, b, fillBlue)
	r, g, b = rowFill(t, a, "row:6")
	wantFill(t, "row 6 untouched plain", r, g, b, fillWhite)
	if err := pagetest.ClickAction(ctx, a.Page(), "row:6"); err != nil {
		t.Fatal(err)
	}
	r, g, b = rowFill(t, a, "row:5")
	wantFill(t, "row 5 deselected", r, g, b, fillGray)
	r, g, b = rowFill(t, a, "row:6")
	wantFill(t, "row 6 selected", r, g, b, fillBlue)
}

// TestInvoiceThemePatch checks the batched theme path on a small table: one
// class edit per row in a single batch, no reparse or recascade, all rows
// painted white and the zebra restored on toggle back.
func TestInvoiceThemePatch(t *testing.T) {
	t.Parallel()
	a, err := newInvoiceAppWithCount(20)
	if err != nil {
		t.Fatal(err)
	}
	a.patch = true
	ctx := context.Background()
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := a.Page().Probe().Stats()
	if err := pagetest.ClickAction(ctx, a.Page(), "theme"); err != nil {
		t.Fatal(err)
	}
	if a.theme != invoiceThemeLight {
		t.Fatalf("theme = %q, want light", a.theme)
	}
	after := a.Page().Probe().Stats()
	if after.Parses != before.Parses || after.Cascades != before.Cascades {
		t.Errorf("patch theme reparsed: parses %d->%d cascades %d->%d",
			before.Parses, after.Parses, before.Cascades, after.Cascades)
	}
	r, g, b := rowFill(t, a, "row:7")
	wantFill(t, "row 7 light theme", r, g, b, fillWhite)
	if err := pagetest.ClickAction(ctx, a.Page(), "theme"); err != nil {
		t.Fatal(err)
	}
	r, g, b = rowFill(t, a, "row:7")
	wantFill(t, "row 7 gray theme back", r, g, b, fillGray)
}

// TestInvoiceThemeStyle checks the stylesheet-swap path: flipping the theme
// restyles every row with no reparse, and the paint follows.
func TestInvoiceThemeStyle(t *testing.T) {
	t.Parallel()
	a, err := newInvoiceApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := a.Page().Probe().Stats()
	if err := a.styleThemeSwap(ctx); err != nil {
		t.Fatal(err)
	}
	if a.theme != invoiceThemeLight {
		t.Fatalf("theme = %q, want light", a.theme)
	}
	after := a.Page().Probe().Stats()
	if after.Cascades <= before.Cascades {
		t.Errorf("style swap recascaded nothing")
	}
	r, g, b := rowFill(t, a, "row:7")
	wantFill(t, "row 7 light theme", r, g, b, fillWhite)
}

// TestInvoiceRefreshSkipsCleanState checks the render-skip: refresh() no-ops
// when the state version matches the last committed render, leaving
// parse/cascade/layout counters flat, and a real change after that skip
// still rebuilds.
func TestInvoiceRefreshSkipsCleanState(t *testing.T) {
	t.Parallel()
	a, err := newInvoiceAppWithCount(20)
	if err != nil {
		t.Fatal(err)
	}
	a.patch = false
	ctx := context.Background()
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	// Commit a real change so the skip has a rendered version to match.
	a.selected = 3
	a.version++
	preCommit := a.Page().Probe().Stats()
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	committed := a.Page().Probe().Stats()
	if committed.Layouts <= preCommit.Layouts {
		t.Fatalf("committing refresh laid nothing out: layouts %d->%d", preCommit.Layouts, committed.Layouts)
	}
	// Second refresh with no mutation must skip the rebuild.
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	skipped := a.Page().Probe().Stats()
	if skipped.Parses != committed.Parses || skipped.Cascades != committed.Cascades || skipped.Layouts != committed.Layouts {
		t.Errorf("clean refresh rebuilt: parses %d->%d cascades %d->%d layouts %d->%d, want flat",
			committed.Parses, skipped.Parses, committed.Cascades, skipped.Cascades, committed.Layouts, skipped.Layouts)
	}
	// A real change after the skip must still rebuild.
	a.selected = 5
	a.version++
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	rebuilt := a.Page().Probe().Stats()
	if rebuilt.Layouts <= skipped.Layouts {
		t.Errorf("change after skipped refresh laid nothing out: layouts %d->%d, want a rebuild",
			skipped.Layouts, rebuilt.Layouts)
	}
	if want := a.selectedText(); !strings.Contains(a.Page().HTML(), want) {
		t.Errorf("rebuilt HTML missing selection readout %q", want)
	}
}

// BenchmarkInvoiceRedraw measures a warm Redraw with unchanged data. The page
// skips layout on an identical frame, so this is the cache-hit floor.
func BenchmarkInvoiceRedraw(b *testing.B) {
	ctx := context.Background()
	a, err := newInvoiceApp()
	if err != nil {
		b.Fatal(err)
	}
	if err := a.Page().Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := a.Page().Redraw(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkInvoiceRebuild measures rebuilding the whole thousand-row builder
// tree with unchanged state: HTML render plus parse, cascade, and layout.
func BenchmarkInvoiceRebuild(b *testing.B) {
	ctx := context.Background()
	a, err := newInvoiceApp()
	if err != nil {
		b.Fatal(err)
	}
	if err := a.Page().Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := a.refresh(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkInvoiceClickUpdate measures the full data-update path: alternating
// clicks on two rows through the page dispatcher, forcing a rebuild each time.
func BenchmarkInvoiceClickUpdate(b *testing.B) {
	ctx := context.Background()
	a, err := newInvoiceApp()
	if err != nil {
		b.Fatal(err)
	}
	if err := a.Page().Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	targets := []string{"row:500", "row:501"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		if err := pagetest.ClickAction(ctx, a.Page(), targets[i%len(targets)]); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkInvoiceClickPatch measures the fine-grained select path:
// alternating clicks through the page dispatcher, each committing two class
// edits plus a text edit instead of a rebuild.
func BenchmarkInvoiceClickPatch(b *testing.B) {
	ctx := context.Background()
	a, err := newInvoiceApp()
	if err != nil {
		b.Fatal(err)
	}
	a.patch = true
	if err := a.Page().Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	targets := []string{"row:500", "row:501"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		if err := pagetest.ClickAction(ctx, a.Page(), targets[i%len(targets)]); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkInvoiceThemePatch measures the batched theme path: each click
// flips all thousand row classes inside one redraw batch.
func BenchmarkInvoiceThemePatch(b *testing.B) {
	ctx := context.Background()
	a, err := newInvoiceApp()
	if err != nil {
		b.Fatal(err)
	}
	a.patch = true
	if err := a.Page().Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := pagetest.ClickAction(ctx, a.Page(), "theme"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkInvoiceThemeStyle measures the stylesheet-swap path: each toggle
// restyles every row with one cascade, layout, and paint, and no reparse.
func BenchmarkInvoiceThemeStyle(b *testing.B) {
	ctx := context.Background()
	a, err := newInvoiceApp()
	if err != nil {
		b.Fatal(err)
	}
	if err := a.Page().Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := a.styleThemeSwap(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkInvoiceFrameThroughput converts rebuild cost into FPS so the report
// can compare frames per second directly with mygo.
func BenchmarkInvoiceFrameThroughput(b *testing.B) {
	ctx := context.Background()
	a, err := newInvoiceApp()
	if err != nil {
		b.Fatal(err)
	}
	if err := a.Page().Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := a.refresh(ctx); err != nil {
			b.Fatal(err)
		}
	}
}
