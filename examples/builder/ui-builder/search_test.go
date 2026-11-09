package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestWidgetsSearchHitsFieldPanelAndBackdrop(t *testing.T) {
	ctx := context.Background()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, a.page, "widgets-search-input"); err != nil {
		t.Fatal(err)
	}
	if !a.widgets.searchOpen {
		t.Fatal("click did not open the search card")
	}
	field, ok := pagetest.BoxByID(a.page, "widgets-search-input")
	if !ok {
		t.Fatal("missing search field")
	}
	if field.W < 80 {
		t.Fatalf("field width %.0f does not cover the control", field.W)
	}
	for _, pt := range []struct {
		name string
		x, y float64
	}{
		{"left", field.X + 4, field.Y + field.H/2},
		{"middle", field.X + field.W/2, field.Y + field.H/2},
		{"right", field.X + field.W - 2, field.Y + field.H/2},
	} {
		if err := a.page.Click(ctx, pt.x, pt.y); err != nil {
			t.Fatal(err)
		}
		if !a.widgets.searchOpen {
			t.Fatalf("click on the %s of the field closed the card", pt.name)
		}
	}
	if a.widgets.searchValue != "" {
		t.Fatalf("field click selected %q", a.widgets.searchValue)
	}
	if err := a.page.Click(ctx, 8, 8); err != nil {
		t.Fatal(err)
	}
	if a.widgets.searchOpen {
		t.Fatal("click outside the field and the panel did not close the card")
	}
	if a.widgets.searchValue != "" {
		t.Fatalf("backdrop click selected %q", a.widgets.searchValue)
	}
	if err := pagetest.Click(ctx, a.page, "widgets-search-input"); err != nil {
		t.Fatal(err)
	}
	option, ok := pagetest.BoxByID(a.page, "widgets-search-option-0")
	if !ok {
		t.Fatal("missing search option")
	}
	if err := a.page.Click(ctx, option.X+option.W/2, option.Y+option.H/2); err != nil {
		t.Fatal(err)
	}
	if a.widgets.searchValue != "buttons" {
		t.Fatalf("panel click selected %q, want buttons", a.widgets.searchValue)
	}
}

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
