package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestListShowsRows(t *testing.T) {
	t.Parallel()
	a, err := newListApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	html := a.Page().HTML()
	for _, want := range []string{"Compare Table", listSubtitle, "Item 1", "Item 1000", "compare-row-500", "Description", "Num"} {
		if !strings.Contains(html, want) {
			t.Errorf("list HTML missing %q", want)
		}
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
	if err := pagetest.ClickAction(ctx, a.Page(), "row:500"); err != nil {
		t.Fatal(err)
	}
	if a.selected != 500 {
		t.Fatalf("selected = %d, want 500", a.selected)
	}
	if html := a.Page().HTML(); !strings.Contains(html, "Selected: Item 501") {
		t.Errorf("list HTML missing selection readout")
	} else if !strings.Contains(html, "#dbeafe") {
		t.Errorf("list HTML missing selected-row highlight")
	}
}

// BenchmarkListRedraw measures a warm Redraw of the thousand-row table with
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
// row through the page dispatcher, which selects the row and rebuilds the
// whole table view with the new data.
func BenchmarkListClickUpdate(b *testing.B) {
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
		if err := pagetest.ClickAction(ctx, a.Page(), "row:500"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkListUpdate measures rebuilding the whole thousand-row builder
// tree with unchanged state through View.Update.
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
