package main

import (
	"context"
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
