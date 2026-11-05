package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestWidgetsSearchOpensFiltersAndCloses(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	// Closed by default: no backdrop card.
	if html := a.page.HTML(); strings.Contains(html, "widgets-search-backdrop") {
		t.Fatal("search backdrop renders while closed")
	}
	// Left field position before opening.
	before, ok := pagetest.BoxByID(a.page, "widgets-input")
	if !ok {
		t.Fatal("missing left input field")
	}
	// Click the search field: the card with prefilled rows opens over a backdrop.
	if err := pagetest.Click(ctx, a.page, "widgets-search-input"); err != nil {
		t.Fatalf("click search field: %v", err)
	}
	if !a.widgets.searchOpen {
		t.Fatal("click did not open the search card")
	}
	html := a.page.HTML()
	if !strings.Contains(html, "widgets-search-backdrop") || !strings.Contains(html, "widgets-search-panel") {
		t.Fatal("open search misses its backdrop card")
	}
	if !strings.Contains(html, "Buttons") {
		t.Fatal("open search misses prefilled rows")
	}
	// Absolute panel must not move the left-hand field.
	after, ok := pagetest.BoxByID(a.page, "widgets-input")
	if !ok {
		t.Fatal("missing left input field after open")
	}
	if before.X != after.X || before.Y != after.Y {
		t.Fatalf("opening search moved the left field: before=%+v after=%+v", before, after)
	}
	// Typing filters live: only Buttons survives "but".
	a.page.Form().SetFormValue("widgets-search-input", "but")
	if err := a.handleChange(ctx, ownframe.Box{ID: "widgets-search-input"}); err != nil {
		t.Fatal(err)
	}
	html = a.page.HTML()
	if !strings.Contains(html, "Buttons") {
		t.Fatal("typing did not keep the Buttons row")
	}
	if got := strings.Count(html, `role="option"`); got != 1 {
		t.Fatalf("live filtering kept %d options, want 1", got)
	}
	// No local match names the background fetch.
	a.page.Form().SetFormValue("widgets-search-input", "zzz-no-match")
	if err := a.handleChange(ctx, ownframe.Box{ID: "widgets-search-input"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(a.notice, "background fetch") {
		t.Fatalf("no-match notice should name the background fetch, got %q", a.notice)
	}
	// Backdrop click closes.
	if err := pagetest.Click(ctx, a.page, "widgets-search-backdrop"); err != nil {
		t.Fatalf("click backdrop: %v", err)
	}
	if a.widgets.searchOpen || strings.Contains(a.page.HTML(), "widgets-search-backdrop") {
		t.Fatal("backdrop click did not close the search card")
	}
}

func TestCatalogSearchFiltersLive(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	a.tab = "catalog"
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	a.page.Form().SetFormValue("catalog-search-input", "1001")
	if err := a.handleChange(ctx, ownframe.Box{ID: "catalog-search-input"}); err != nil {
		t.Fatal(err)
	}
	html := a.page.HTML()
	if !strings.Contains(html, "Order #1001") {
		t.Fatal("catalog search missed Order #1001")
	}
	if strings.Contains(html, "Order #1002") {
		t.Fatal("catalog search did not filter live")
	}
}
