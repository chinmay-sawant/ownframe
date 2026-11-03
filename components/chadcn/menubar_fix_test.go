package chadcn_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
)

func TestMenubarFixFlatRendersTriggerRow(t *testing.T) {
	t.Parallel()
	markup := renderMenuMarkup(t, chadcn.Menubar(chadcn.MenubarOptions{
		ID:    "top-menubar",
		Label: "Application",
		Open:  "file",
		Flat:  true,
		Menus: []chadcn.MenubarMenu{
			{ID: "file", Label: "File", Items: []chadcn.MenuItem{{Label: "Refresh", Shortcut: "⌘R"}}},
			{ID: "view", Label: "View", Items: []chadcn.MenuItem{{Label: "Overview"}}},
		},
	}))
	requireMenuMarkup(t, markup,
		`class="of-menubar of-menubar--flat"`,
		`id="top-menubar"`,
		`class="of-menubar-bar"`,
		`role="menubar"`,
		`aria-label="Application"`,
		`class="of-menubar-slot"`,
		`class="of-menubar-trigger of-focus of-transition"`,
		`id="file-trigger"`,
		`id="view-trigger"`,
		`aria-expanded="true"`,
		`aria-expanded="false"`,
		`aria-controls="file-menu"`,
		`class="of-menubar-menu"`,
		`Refresh`,
	)
}

func TestMenubarFixDefaultKeepsBoxedRoot(t *testing.T) {
	t.Parallel()
	markup := renderMenuMarkup(t, chadcn.Menubar(chadcn.MenubarOptions{
		Menus: []chadcn.MenubarMenu{
			{ID: "file", Label: "File", Items: []chadcn.MenuItem{{Label: "New Tab"}}},
		},
	}))
	requireMenuMarkup(t, markup, `class="of-menubar"`, `class="of-menubar-bar"`)
	requireMenuEmpty(t, markup, "of-menubar--flat")
}

func TestMenubarFixStylesheetFlattensBar(t *testing.T) {
	t.Parallel()
	bar := menubarFixCSSRule(chadcn.Stylesheet, ".of-menubar--flat .of-menubar-bar")
	if bar == "" {
		t.Fatal("stylesheet missing flat menubar bar rule")
	}
	for _, want := range []string{"padding: 0;", "border: 0;", "background-color: transparent;"} {
		if !strings.Contains(bar, want) {
			t.Errorf("flat menubar bar rule missing %q:\n%s", want, bar)
		}
	}
	trigger := menubarFixCSSRule(chadcn.Stylesheet, ".of-menubar--flat .of-menubar-trigger")
	if trigger == "" {
		t.Fatal("stylesheet missing flat menubar trigger rule")
	}
	if !strings.Contains(trigger, "height: 28px;") {
		t.Errorf("flat menubar trigger should keep its height:\n%s", trigger)
	}
}

// menubarFixCSSRule returns one stylesheet rule body, from its selector through
// the closing brace, or the empty string when the selector is absent.
func menubarFixCSSRule(sheet, selector string) string {
	start := strings.Index(sheet, selector+" {")
	if start < 0 {
		return ""
	}
	end := strings.Index(sheet[start:], "}")
	if end < 0 {
		return ""
	}
	return sheet[start : start+end]
}
