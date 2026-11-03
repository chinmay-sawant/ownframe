package chadcn_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// commandFixSample spans three groups, including the Services group that used
// to sit below the clipped list.
func commandFixSample() *ui.Element {
	return chadcn.Command(chadcn.CommandOptions{
		ID:     "palette",
		Active: "overview",
		Groups: []chadcn.CommandGroup{
			{Label: "Navigate", Items: []chadcn.CommandItem{
				{Value: "overview", Label: "Overview", Shortcut: "⌘1"},
				{Value: "logs", Label: "Logs", Shortcut: "⌘2"},
				{Value: "traces", Label: "Traces", Shortcut: "⌘3"},
			}},
			{Label: "Actions", Items: []chadcn.CommandItem{
				{Value: "alert:new", Label: "New alert rule", Shortcut: "⇧A"},
				{Value: "refresh", Label: "Refresh dashboard", Shortcut: "R"},
			}},
			{Label: "Services", Items: []chadcn.CommandItem{
				{Value: "service:checkout", Label: "Filter logs to Checkout"},
				{Value: "service:billing", Label: "Filter logs to Billing"},
			}},
		},
	})
}

// TestCommandFixListIsScrollRegion checks the listbox that owns the bounded
// scroll region and the groups it holds.
func TestCommandFixListIsScrollRegion(t *testing.T) {
	t.Parallel()
	html, err := ui.Render(commandFixSample())
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`class="of-command"`,
		`class="of-command-search"`,
		`class="of-command-list"`,
		`id="palette-list"`,
		`role="listbox"`,
		`aria-orientation="vertical"`,
		`tabindex="0"`,
		`class="of-command-group"`,
		`class="of-command-group-label"`,
		`class="of-command-item"`,
		`class="of-command-separator"`,
		`Services`,
		`Filter logs to Checkout`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("command fix markup missing %q in %s", want, html)
		}
	}
	if got := strings.Count(html, `class="of-command-group"`); got != 3 {
		t.Fatalf("groups=%d want 3", got)
	}
	if got := strings.Count(html, `class="of-command-group-label"`); got != 3 {
		t.Fatalf("group labels=%d want 3", got)
	}
	if got := strings.Count(html, `class="of-command-separator"`); got != 2 {
		t.Fatalf("separators=%d want 2", got)
	}
	if got := strings.Count(html, `role="group"`); got != 3 {
		t.Fatalf("group roles=%d want 3", got)
	}
	if got := strings.Count(html, `role="option"`); got != 7 {
		t.Fatalf("options=%d want 7", got)
	}
}

// TestCommandFixStylesheetBoundsList bounds the list so the labels and rows
// never shrink instead of scrolling. The panel itself hugs the list, so a
// short palette does not paint an empty surface.
func TestCommandFixStylesheetBoundsList(t *testing.T) {
	t.Parallel()
	panel := commandFixRule(t, chadcn.Stylesheet, ".of-command")
	for _, want := range []string{"box-sizing: border-box;", "padding: 6px;"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("command panel rule missing %q:\n%s", want, panel)
		}
	}
	if strings.Contains(panel, "max-height:") {
		t.Fatalf("command panel must hug the list, not take a fixed height:\n%s", panel)
	}
	list := commandFixRule(t, chadcn.Stylesheet, ".of-command-list")
	for _, want := range []string{
		"flex: 1 1 auto;",
		"min-height: 0;",
		"max-height: var(--of-command-max-height, 420px);",
		"overflow-x: hidden;",
		"overflow-y: auto;",
	} {
		if !strings.Contains(list, want) {
			t.Fatalf("command list rule missing %q:\n%s", want, list)
		}
	}
	group := commandFixRule(t, chadcn.Stylesheet, ".of-command-group")
	if !strings.Contains(group, "flex-shrink: 0;") {
		t.Fatalf("command group can shrink inside the list:\n%s", group)
	}
	item := commandFixRule(t, chadcn.Stylesheet, ".of-command-item")
	for _, want := range []string{"flex-shrink: 0;", "padding: 6px 8px;"} {
		if !strings.Contains(item, want) {
			t.Fatalf("command item rule missing %q:\n%s", want, item)
		}
	}
	label := commandFixRule(t, chadcn.Stylesheet, ".of-command-group-label")
	for _, want := range []string{"padding: 4px 8px 2px;", "font-size: 11px;"} {
		if !strings.Contains(label, want) {
			t.Fatalf("command group label rule missing %q:\n%s", want, label)
		}
	}
	search := commandFixRule(t, chadcn.Stylesheet, ".of-command-search")
	if !strings.Contains(search, "padding: 8px 10px;") {
		t.Fatalf("command search rule is not tightened:\n%s", search)
	}
}

// commandFixRule returns one stylesheet rule body, from its selector through
// the closing brace, or fails when the rule is absent.
func commandFixRule(t *testing.T, sheet, selector string) string {
	t.Helper()
	start := strings.Index(sheet, selector+" {")
	if start < 0 {
		t.Fatalf("stylesheet is missing %s", selector)
	}
	end := strings.Index(sheet[start:], "}")
	if end < 0 {
		t.Fatalf("stylesheet rule %s is unterminated", selector)
	}
	return sheet[start : start+end]
}
