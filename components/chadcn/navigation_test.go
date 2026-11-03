package chadcn_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
)

func navigationSample(open string) *ui.Element {
	return chadcn.NavigationMenu(chadcn.NavigationMenuOptions{
		ID:    "primary",
		Label: "Main",
		Open:  open,
		Items: []chadcn.NavigationItem{
			{
				ID:    "getting-started",
				Label: "Getting started",
				Children: []chadcn.NavigationItem{
					{Label: "Introduction", Description: "Re-usable components built with Tailwind CSS."},
					{Label: "Installation", Description: "How to install dependencies and structure your app."},
					{Label: "Typography", Description: "Styles for headings, paragraphs, lists...etc"},
				},
			},
			{
				ID:    "components",
				Label: "Components",
				Children: []chadcn.NavigationItem{
					{Label: "Alert", Description: "A short message for the reader."},
				},
			},
			{ID: "docs", Label: "Docs"},
		},
	})
}

func TestNavigationMenuClosedRow(t *testing.T) {
	t.Parallel()
	html, err := ui.Render(navigationSample(""))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`<nav`,
		`role="navigation"`,
		`aria-label="Main"`,
		`class="of-navigation-menu"`,
		`class="of-navigation-menu-list"`,
		`class="of-navigation-menu-trigger"`,
		`class="of-navigation-menu-link"`,
		`id="getting-started-trigger"`,
		`id="docs-trigger"`,
		`class="of-navigation-menu-chevron"`,
		`aria-expanded="false"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("closed menu missing %q in %s", want, html)
		}
	}
	if got := strings.Count(html, `aria-expanded="false"`); got != 2 {
		t.Fatalf("closed triggers=%d want 2", got)
	}
	if strings.Contains(html, "of-navigation-menu-content") {
		t.Fatalf("closed menu rendered a panel: %s", html)
	}
}

func TestNavigationMenuOpenPanel(t *testing.T) {
	t.Parallel()
	html, err := ui.Render(navigationSample("getting-started"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`aria-expanded="true"`,
		`aria-controls="getting-started-content"`,
		`id="getting-started-content"`,
		`aria-labelledby="getting-started-trigger"`,
		`role="region"`,
		`class="of-navigation-menu-content"`,
		`class="of-navigation-menu-entries"`,
		`class="of-navigation-menu-entry-title"`,
		`Introduction`,
		`Re-usable components built with Tailwind CSS.`,
		`Typography`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("open menu missing %q in %s", want, html)
		}
	}
	if got := strings.Count(html, "of-navigation-menu-content"); got != 1 {
		t.Fatalf("content class appears %d times, want 1", got)
	}
	if strings.Count(html, `role="region"`) != 1 {
		t.Fatal("expected exactly one open panel")
	}
	if strings.Contains(html, "A short message for the reader.") {
		t.Fatal("closed item leaked into the open panel")
	}
	if got := strings.Count(html, `aria-expanded="true"`); got != 1 {
		t.Fatalf("open triggers=%d want 1", got)
	}
	if got := strings.Count(html, `aria-expanded="false"`); got != 1 {
		t.Fatalf("closed triggers=%d want 1", got)
	}
}

func TestNavigationMenuReportsTrigger(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var reported []string
	tree := chadcn.NavigationMenu(chadcn.NavigationMenuOptions{
		ID: "primary",
		Items: []chadcn.NavigationItem{
			{ID: "getting-started", Label: "Getting started", Children: []chadcn.NavigationItem{{Label: "Introduction"}}},
			{ID: "docs", Label: "Docs"},
		},
		OnOpenChange: func(_ context.Context, id string) error {
			reported = append(reported, id)
			return nil
		},
	})
	p, err := ui.NewPage(ownframe.Config{Width: 640, Height: 240}, tree)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, p, "getting-started-trigger"); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, p, "docs-trigger"); err != nil {
		t.Fatal(err)
	}
	if len(reported) != 2 || reported[0] != "getting-started" || reported[1] != "docs" {
		t.Fatalf("reported=%v want [getting-started docs]", reported)
	}
}

func commandSample() *ui.Element {
	return chadcn.Command(chadcn.CommandOptions{
		ID:     "palette",
		Query:  "cal",
		Active: "calendar",
		Groups: []chadcn.CommandGroup{
			{Label: "Suggestions", Items: []chadcn.CommandItem{
				{Value: "calendar", Label: "Calendar", Icon: ui.Span().Text("C")},
				{Value: "emoji", Label: "Search Emoji"},
				{Value: "calculator", Label: "Calculator", Disabled: true},
			}},
			{Label: "Settings", Items: []chadcn.CommandItem{
				{Value: "profile", Label: "Profile", Shortcut: "⌘P"},
				{Value: "billing", Label: "Billing", Shortcut: "⌘B"},
				{Value: "settings", Label: "Settings", Shortcut: "⌘S"},
			}},
		},
	})
}

func TestCommandPaletteGroupsAndActiveRow(t *testing.T) {
	t.Parallel()
	html, err := ui.Render(commandSample())
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`class="of-command"`,
		`id="palette"`,
		`class="of-command-search"`,
		`class="of-command-search-icon"`,
		`class="of-command-input"`,
		`id="palette-input"`,
		`value="cal"`,
		`placeholder="Type a command or search..."`,
		`role="combobox"`,
		`aria-controls="palette-list"`,
		`role="listbox"`,
		`id="palette-list"`,
		`class="of-command-group"`,
		`role="group"`,
		`aria-label="Suggestions"`,
		`aria-label="Settings"`,
		`class="of-command-group-label"`,
		`class="of-command-separator"`,
		`class="of-command-item"`,
		`role="option"`,
		`aria-selected="true"`,
		`class="of-command-item-label"`,
		`Search Emoji`,
		`class="of-command-icon"`,
		`class="of-command-shortcut"`,
		`⌘P`,
		`disabled="disabled"`,
		`aria-disabled="true"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("command missing %q in %s", want, html)
		}
	}
	if got := strings.Count(html, `role="option"`); got != 6 {
		t.Fatalf("options=%d want 6", got)
	}
	if got := strings.Count(html, `aria-selected="true"`); got != 1 {
		t.Fatalf("selected rows=%d want 1", got)
	}
	if got := strings.Count(html, `aria-selected="false"`); got != 5 {
		t.Fatalf("unselected rows=%d want 5", got)
	}
	if strings.Contains(html, "No results") {
		t.Fatal("non-empty palette rendered the empty state")
	}
}

func TestCommandEmptyState(t *testing.T) {
	t.Parallel()
	html, err := ui.Render(chadcn.Command(chadcn.CommandOptions{
		ID:     "empty",
		Query:  "zzz",
		Active: "zzz",
		Groups: []chadcn.CommandGroup{{Label: "Suggestions"}},
	}))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`class="of-command-empty"`,
		`No results`,
		`role="listbox"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("empty palette missing %q in %s", want, html)
		}
	}
	if strings.Contains(html, `role="option"`) {
		t.Fatal("empty palette rendered options")
	}
}

func TestCommandCallbacks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var selected []string
	tree := chadcn.Command(chadcn.CommandOptions{
		ID:     "cmd",
		Query:  "cal",
		Active: "calendar",
		Groups: []chadcn.CommandGroup{{Label: "Suggestions", Items: []chadcn.CommandItem{
			{Value: "calendar", Label: "Calendar"},
			{Value: "calculator", Label: "Calculator", Disabled: true},
		}}},
		OnSelect: func(_ context.Context, value string) error {
			selected = append(selected, value)
			return nil
		},
	})
	p, err := ui.NewPage(ownframe.Config{Width: 560, Height: 400, Theme: chadcn.Stylesheet}, tree)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	clickLabel := func(label string) {
		t.Helper()
		for _, box := range p.Boxes() {
			if box.Tag == "span" && box.Text == label {
				if err := p.Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
					t.Fatal(err)
				}
				return
			}
		}
		t.Fatalf("missing %q row", label)
	}
	clickLabel("Calendar")
	if len(selected) != 1 || selected[0] != "calendar" {
		t.Fatalf("selected=%v want [calendar]", selected)
	}
	clickLabel("Calculator")
	if len(selected) != 1 {
		t.Fatalf("disabled row dispatched %v", selected)
	}
}
