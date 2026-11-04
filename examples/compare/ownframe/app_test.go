package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestListShowsWindow(t *testing.T) {
	t.Parallel()
	a, err := newListApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	html := a.Page().HTML()
	for _, want := range []string{"Compare Table", listSubtitle, "Item 1", "compare-row-0", "Description", "Num"} {
		if !strings.Contains(html, want) {
			t.Errorf("list HTML missing %q", want)
		}
	}
	// The window holds the viewport slice, not the whole table.
	for _, absent := range []string{"Item 1000", "compare-row-999"} {
		if strings.Contains(html, absent) {
			t.Errorf("list HTML should not hold %q outside the window", absent)
		}
	}
	if !strings.Contains(html, "compare-bot-pad") {
		t.Errorf("list HTML missing the bottom spacer that keeps full height")
	}
}

func TestListSelectRow(t *testing.T) {
	t.Parallel()
	a, err := newListApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.Page(), "row:5"); err != nil {
		t.Fatal(err)
	}
	if a.selected != 5 {
		t.Fatalf("selected = %d, want 5", a.selected)
	}
	if html := a.Page().HTML(); !strings.Contains(html, "Selected: Item 6") {
		t.Errorf("list HTML missing selection readout")
	}
	// A no-op click on the selected row leaves state alone.
	if err := pagetest.ClickAction(ctx, a.Page(), "row:5"); err != nil {
		t.Fatal(err)
	}
	if a.selected != 5 {
		t.Fatalf("selected = %d, want 5", a.selected)
	}
}

func TestListScrollWindowMoves(t *testing.T) {
	t.Parallel()
	a, err := newListApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	first := a.winStart
	// Drive the window callback directly: headless there is no window to
	// own the offset, so ScrollTo would only queue. The callback reads
	// layout boxes, which Redraw above produced.
	if !a.slice(2000, 600) {
		t.Fatal("scroll window did not move for offset 2000")
	}
	if a.winStart == first {
		t.Fatalf("winStart stuck at %d", first)
	}
	if !a.windowPending {
		t.Fatal("scroll move did not arm the tick rebuild")
	}
	// Settle back to the true headless offset: the rebuild must disarm.
	a.slice(0, 600)
	if err := a.tick(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.tick(ctx); err != nil {
		t.Fatal(err)
	}
	if a.windowPending {
		t.Fatal("tick did not disarm after rebuild")
	}
	want := a.rows[a.winStart].title
	if html := a.Page().HTML(); !strings.Contains(html, want) {
		t.Errorf("rebuilt HTML missing window start %q", want)
	}
}

// BenchmarkListRedraw measures a warm Redraw of the windowed table with
// unchanged data.
func BenchmarkListRedraw(b *testing.B) {
	ctx := context.Background()
	a, err := newListApp()
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

// BenchmarkListClickUpdate measures the full data-update path: a click on a
// visible row through the page dispatcher, which selects the row and rebuilds
// the small windowed view.
func BenchmarkListClickUpdate(b *testing.B) {
	ctx := context.Background()
	a, err := newListApp()
	if err != nil {
		b.Fatal(err)
	}
	if err := a.Page().Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	// Alternate two rows so every click changes the selection.
	actions := []string{"row:5", "row:6"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		if err := pagetest.ClickAction(ctx, a.Page(), actions[i%len(actions)]); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkListUpdate measures rebuilding the windowed builder tree with
// unchanged state through View.Update.
func BenchmarkListUpdate(b *testing.B) {
	ctx := context.Background()
	a, err := newListApp()
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
