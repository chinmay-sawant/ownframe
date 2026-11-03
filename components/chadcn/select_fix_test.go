package chadcn_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
)

// TestSelectFixOpenRootClass covers the open modifier on the root and the
// placement class on the panel it holds.
func TestSelectFixOpenRootClass(t *testing.T) {
	t.Parallel()
	items := []chadcn.SelectItem{
		{Value: "error", Label: "Error"},
		{Value: "warn", Label: "Warn"},
	}
	open := renderSelectionHTML(t, chadcn.Select(chadcn.SelectOptions{
		ID:    "severity",
		Open:  true,
		Value: "error",
		Items: items,
	}))
	requireSelectionMarkup(t, open,
		`class="of-scope of-select of-select--open"`,
		`class="of-select-panel"`,
	)
	closed := renderSelectionHTML(t, chadcn.Select(chadcn.SelectOptions{
		ID:    "severity",
		Value: "error",
		Items: items,
	}))
	requireSelectionMarkup(t, closed, `class="of-scope of-select"`)
	if strings.Contains(closed, "of-select--open") || strings.Contains(closed, "of-select-panel") {
		t.Fatalf("closed select leaked open state:\n%s", closed)
	}
}

// TestSelectFixPanelPlacement keeps the panel absolutely placed under the
// trigger and makes the open root the stacking context that lifts it.
func TestSelectFixPanelPlacement(t *testing.T) {
	t.Parallel()
	panel := selectFixRule(t, chadcn.Stylesheet, ".of-select-panel")
	for _, want := range []string{"position: absolute;", "top: 40px;", "width: 100%;"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("panel rule is missing %q:\n%s", want, panel)
		}
	}
	trigger := selectFixRule(t, chadcn.Stylesheet, ".of-select-trigger")
	if !strings.Contains(trigger, "width: 100%;") {
		t.Fatalf("trigger rule does not match the panel width:\n%s", trigger)
	}
	open := selectFixRule(t, chadcn.Stylesheet, ".of-select--open")
	if !strings.Contains(open, "z-index: 50;") {
		t.Fatalf("open root does not establish a stacking context:\n%s", open)
	}
}

// selectFixRule returns one rule body from the stylesheet, without its braces.
func selectFixRule(t *testing.T, sheet, selector string) string {
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
