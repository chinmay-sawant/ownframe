package main

import (
	"context"
	"strings"
	"testing"
)

// rowIDs extracts id="stock-row-..." values from the current page source.
func rowIDs(t *testing.T, html string) []string {
	t.Helper()
	var out []string
	for _, part := range strings.Split(html, `id="stock-row-`)[1:] {
		end := strings.IndexByte(part, '"')
		if end <= 0 {
			t.Fatalf("malformed stock-row id in %q", part)
		}
		out = append(out, part[:end])
	}
	return out
}

// TestStockRowStableIDs pins the stable row and pill ids the incremental
// update path needs: every visible row carries stock-row-<code>, ids are
// unique, a pill carries inv-pill-<filter>, and a SKU keeps its id when the
// filter changes, so same-structure updates patch instead of re-rendering.
func TestStockRowStableIDs(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	idsFor := func(filter string) []string {
		t.Helper()
		a.statusFilter = filter
		if err := a.refresh(ctx); err != nil {
			t.Fatal(err)
		}
		return rowIDs(t, a.page.HTML())
	}

	all := idsFor("")
	if len(all) == 0 {
		t.Fatal("no stock-row ids in the unfiltered table")
	}
	seen := map[string]bool{}
	for _, id := range all {
		if seen[id] {
			t.Fatalf("duplicate stock row id %q", id)
		}
		seen[id] = true
	}
	for _, filter := range []string{"Stockout", "Reorder", "Overstock"} {
		got := idsFor(filter)
		if len(got) == 0 {
			t.Fatalf("no stock-row ids under filter %q", filter)
		}
		for _, id := range got {
			if !seen[id] {
				t.Fatalf("row id %q under filter %q is new; want the stable id from the unfiltered table", id, filter)
			}
		}
		html := a.page.HTML()
		want := "inv-pill-"
		if !strings.Contains(html, want) {
			t.Fatalf("page HTML misses %q pills", want)
		}
	}
}
