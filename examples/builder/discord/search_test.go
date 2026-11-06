package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// Search palette filters channels and messages live as the query changes,
// and picking a row closes the palette without keeping a stale filter.
func TestSearchFiltersChannelsAndMessages(t *testing.T) {
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "search:open"}); err != nil {
		t.Fatal(err)
	}
	if !a.searchOpen {
		t.Fatal("search did not open")
	}
	if got := a.page.Form().FocusID(); got != "dc-search" {
		t.Fatalf("focus = %q, want dc-search", got)
	}
	a.page.Form().SetFormValue("dc-search", "dev-log")
	if err := a.handleChange(ctx, ownframe.Box{ID: "dc-search"}); err != nil {
		t.Fatal(err)
	}
	html := a.page.HTML()
	if !strings.Contains(html, "dev-log") {
		t.Fatal("channel match missing after filter")
	}
	if strings.Contains(html, "No results for") {
		t.Fatalf("unexpected empty hint in %q", html[len(html)-500:])
	}
	a.page.Form().SetFormValue("dc-search", "zzz-no-match")
	if err := a.handleChange(ctx, ownframe.Box{ID: "dc-search"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(a.page.HTML(), "No results for") {
		t.Fatal("no-match hint missing")
	}
}

// Selecting a channel row from the palette closes it and clears the query.
func TestSearchSelectChannelClosesPalette(t *testing.T) {
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "search:open"}); err != nil {
		t.Fatal(err)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "chan:select:ownframe-general"}); err != nil {
		t.Fatal(err)
	}
	if a.searchOpen {
		t.Fatal("palette stayed open after channel select")
	}
	if a.searchQuery != "" {
		t.Fatalf("stale query %q after select", a.searchQuery)
	}
}

// Ctrl+K opens the palette, fulfilling the topbar tooltip promise.
func TestSearchCtrlKOpensPalette(t *testing.T) {
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.handleKey(ctx, "ctrl+k"); err != nil {
		t.Fatal(err)
	}
	if !a.searchOpen {
		t.Fatal("ctrl+k did not open search")
	}
}
