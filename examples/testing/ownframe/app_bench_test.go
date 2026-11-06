package main

import (
	"context"
	"runtime"
	"testing"

	"github.com/chinmay-sawant/ownframe/pagetest"
)

// BenchmarkInvoiceClickBatch measures alternating selection clicks on the
// batched-select path: with patch enabled, clicks route to
// patchSelectBatched, which commits all edits in one PatchElements call.
func BenchmarkInvoiceClickBatch(b *testing.B) {
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

// BenchmarkInvoiceViewUpdate measures the builder's keyed update path with
// patch disabled at the app level: View.Update diffs the rebuilt tree and
// commits the changed elements instead of rendering the whole document.
func BenchmarkInvoiceViewUpdate(b *testing.B) {
	ctx := context.Background()
	a, err := newInvoiceApp()
	if err != nil {
		b.Fatal(err)
	}
	a.patch = false
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

// BenchmarkInvoiceWindowedRebuild measures a windowed-app refresh: only the
// viewport slice rebuilds, so build cost scales with viewport, not table.
func BenchmarkInvoiceWindowedRebuild(b *testing.B) {
	ctx := context.Background()
	a, err := newInvoiceApp()
	if err != nil {
		b.Fatal(err)
	}
	if err := a.Page().Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	a.setWindowed(0)
	if err := a.refresh(ctx); err != nil {
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

// BenchmarkInvoiceThemeLoopRegression measures the per-row patchTheme loop on
// a small-count app.
//
// Forbid rule: the per-row patchTheme loop must stay far slower than the
// stylesheet swap (styleThemeSwap) or a full rebuild (refresh). Never use
// patchTheme for theme changes; prefer styleThemeSwap. This benchmark pins
// the slow path so a regression (it getting close to swap/rebuild cost)
// is visible.
func BenchmarkInvoiceThemeLoopRegression(b *testing.B) {
	ctx := context.Background()
	a, err := newInvoiceAppWithCount(20)
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

// BenchmarkInvoiceColdBuild measures a fresh page plus its first Redraw:
// the closest headless proxy for first paint. Engine-level caches (fonts,
// styles) persist across iterations in-process, so treat this as a cold
// page on a warm engine.
func BenchmarkInvoiceColdBuild(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		a, err := newInvoiceApp()
		if err != nil {
			b.Fatal(err)
		}
		if err := a.Page().Redraw(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkInvoiceResize measures alternating resizes with a redraw each:
// the cost of a window resize or orientation change.
func BenchmarkInvoiceResize(b *testing.B) {
	ctx := context.Background()
	a, err := newInvoiceApp()
	if err != nil {
		b.Fatal(err)
	}
	a.patch = false
	if err := a.Page().Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	sizes := [][2]int{{invoiceWidth, invoiceHeight}, {1200, 800}}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for _, size := range sizes {
			a.Page().SetSize(size[0], size[1])
			if err := a.Page().Redraw(ctx); err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkInvoiceTyping measures one keystroke into the filter box plus the
// full select-rebuild it triggers, alternating a typed char and backspace so
// every iteration does constant work: the headless proxy for typing latency.
func BenchmarkInvoiceTyping(b *testing.B) {
	ctx := context.Background()
	a, err := newInvoiceApp()
	if err != nil {
		b.Fatal(err)
	}
	if err := a.Page().Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	if err := pagetest.Click(ctx, a.Page(), "invoice-filter"); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		if i%2 == 0 {
			if err := pagetest.Type(ctx, a.Page(), "a"); err != nil {
				b.Fatal(err)
			}
			continue
		}
		if err := a.Page().Edit().Backspace(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkInvoiceHeapGrowth runs full rebuilds, collects, and reports the
// retained heap. Run with -benchtime=1x: retention is a steady-state
// property, one measurement suffices.
func BenchmarkInvoiceHeapGrowth(b *testing.B) {
	ctx := context.Background()
	a, err := newInvoiceApp()
	if err != nil {
		b.Fatal(err)
	}
	a.patch = false
	if err := a.Page().Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		runtime.GC()
		var before runtime.MemStats
		runtime.ReadMemStats(&before)
		for i := 0; i < 5; i++ {
			if err := pagetest.ClickAction(ctx, a.Page(), "row:500"); err != nil {
				b.Fatal(err)
			}
			if err := pagetest.ClickAction(ctx, a.Page(), "row:501"); err != nil {
				b.Fatal(err)
			}
		}
		var mem runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&mem)
		b.ReportMetric(float64(mem.HeapAlloc), "heap-retained-bytes")
		b.ReportMetric(float64(mem.HeapObjects), "heap-retained-objects")
		b.ReportMetric(float64(int64(mem.HeapAlloc)-int64(before.HeapAlloc)), "heap-growth-bytes")
		b.ReportMetric(float64(int64(mem.HeapObjects)-int64(before.HeapObjects)), "heap-growth-objects")
	}
}
