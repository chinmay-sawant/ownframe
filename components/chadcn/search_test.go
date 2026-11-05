package chadcn_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func renderSearchHTML(t *testing.T, root *ui.Element) string {
	t.Helper()
	rendered, err := ui.Render(root)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return rendered
}

func TestSearchClosedRendersFieldOnly(t *testing.T) {
	t.Parallel()
	rendered := renderSearchHTML(t, chadcn.Search(chadcn.SearchOptions{
		ID:    "site",
		Query: "dash",
		Items: []chadcn.SearchItem{{Value: "dashboard", Label: "Dashboard"}},
	}))
	for _, fragment := range []string{
		`class="of-scope of-search"`,
		`id="site"`,
		`id="site-input"`,
		`type="search"`,
		`value="dash"`,
		`role="combobox"`,
		`aria-expanded="false"`,
		`placeholder="Search..."`,
	} {
		if !strings.Contains(rendered, fragment) {
			t.Fatalf("rendered markup is missing %q:\n%s", fragment, rendered)
		}
	}
	if strings.Contains(rendered, "of-search-backdrop") || strings.Contains(rendered, "of-search-panel") {
		t.Fatalf("closed search rendered a card:\n%s", rendered)
	}
}

func TestSearchOpenRendersBackdropCard(t *testing.T) {
	t.Parallel()
	rendered := renderSearchHTML(t, chadcn.Search(chadcn.SearchOptions{
		ID:    "site",
		Query: "",
		Open:  true,
		Items: []chadcn.SearchItem{
			{Value: "dashboard", Label: "Dashboard", Description: "Overview"},
			{Value: "buttons", Label: "Buttons"},
		},
		OnOpenChange: func(context.Context, bool) error { return nil },
		OnSelect:     func(context.Context, string) error { return nil },
	}))
	for _, fragment := range []string{
		`class="of-scope of-search of-search--open"`,
		`class="of-search-backdrop"`,
		`id="site-backdrop"`,
		`class="of-search-panel"`,
		`id="site-panel"`,
		`role="listbox"`,
		`role="option"`,
		`Dashboard`,
		`Overview`,
		`Buttons`,
		`aria-expanded="true"`,
		`aria-controls="site-panel"`,
	} {
		if !strings.Contains(rendered, fragment) {
			t.Fatalf("rendered markup is missing %q:\n%s", fragment, rendered)
		}
	}
	if got := strings.Count(rendered, `role="option"`); got != 2 {
		t.Fatalf("options=%d want 2", got)
	}
}

func TestSearchFiltersAsQueryChanges(t *testing.T) {
	t.Parallel()
	items := []chadcn.SearchItem{
		{Value: "dashboard", Label: "Dashboard"},
		{Value: "buttons", Label: "Buttons"},
		{Value: "badges", Label: "Badges"},
	}
	rendered := renderSearchHTML(t, chadcn.Search(chadcn.SearchOptions{
		ID:    "site",
		Query: "but",
		Open:  true,
		Items: items,
	}))
	if !strings.Contains(rendered, "Buttons") {
		t.Fatalf("filtered search missing Buttons:\n%s", rendered)
	}
	if strings.Contains(rendered, "Dashboard") || strings.Contains(rendered, "Badges") {
		t.Fatalf("search did not filter as the query changed:\n%s", rendered)
	}
	if got := strings.Count(rendered, `role="option"`); got != 1 {
		t.Fatalf("filtered options=%d want 1", got)
	}
}

func TestSearchEmptyStateKeepsListbox(t *testing.T) {
	t.Parallel()
	rendered := renderSearchHTML(t, chadcn.Search(chadcn.SearchOptions{
		ID:    "site",
		Query: "zzz",
		Open:  true,
		Items: []chadcn.SearchItem{{Value: "dashboard", Label: "Dashboard"}},
	}))
	if !strings.Contains(rendered, `class="of-search-empty"`) {
		t.Fatalf("empty search missing empty state:\n%s", rendered)
	}
	if !strings.Contains(rendered, `role="listbox"`) {
		t.Fatalf("empty search dropped the listbox:\n%s", rendered)
	}
	if strings.Contains(rendered, `role="option"`) {
		t.Fatalf("empty search rendered options:\n%s", rendered)
	}
}

func TestSearchStylesheetHasNoShadow(t *testing.T) {
	t.Parallel()
	for _, fragment := range []string{
		".of-search {",
		".of-search-backdrop {",
		".of-search-panel {",
		"background: transparent;",
		"box-shadow: none;",
		"position: absolute;",
		"position: fixed;",
	} {
		if !strings.Contains(chadcn.Stylesheet, fragment) {
			t.Fatalf("stylesheet missing %q", fragment)
		}
	}
}
