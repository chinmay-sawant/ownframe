package chadcn_test

import (
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
)

func TestPickersFixComboboxOpenClass(t *testing.T) {
	t.Parallel()
	rendered := renderSelectionHTML(t, chadcn.Combobox(chadcn.ComboboxOptions{
		ID:          "service",
		Open:        true,
		Placeholder: "Select a service",
		Items:       []chadcn.ComboboxItem{{Value: "auth", Label: "Auth"}},
	}))
	requireSelectionMarkup(t, rendered,
		`class="of-scope of-combobox of-combobox--open"`,
		`class="of-combobox-input of-transition"`,
		`class="of-combobox-panel"`,
		`role="listbox"`,
	)
	if got := strings.Count(rendered, "of-combobox--open"); got != 1 {
		t.Fatalf("open class count=%d want 1", got)
	}
}

func TestPickersFixComboboxClosedClass(t *testing.T) {
	t.Parallel()
	rendered := renderSelectionHTML(t, chadcn.Combobox(chadcn.ComboboxOptions{
		ID:    "service",
		Items: []chadcn.ComboboxItem{{Value: "auth", Label: "Auth"}},
	}))
	requireSelectionMarkup(t, rendered, `class="of-scope of-combobox"`)
	if strings.Contains(rendered, "of-combobox--open") {
		t.Fatalf("closed combobox rendered the open class:\n%s", rendered)
	}
	if strings.Contains(rendered, "of-combobox-panel") {
		t.Fatalf("closed combobox rendered a panel:\n%s", rendered)
	}
}

func TestPickersFixDatePickerOpenClass(t *testing.T) {
	t.Parallel()
	selected := time.Date(2026, time.November, 3, 0, 0, 0, 0, time.UTC)
	rendered := renderSelectionHTML(t, chadcn.DatePicker(chadcn.DatePickerOptions{
		ID:       "from-date",
		Open:     true,
		Month:    selected,
		Selected: selected,
	}))
	requireSelectionMarkup(t, rendered,
		`class="of-scope of-date-picker of-date-picker--open"`,
		`class="of-date-picker-trigger of-transition"`,
		`class="of-date-picker-panel"`,
		`role="dialog"`,
		`class="of-scope of-calendar"`,
	)
	if got := strings.Count(rendered, "of-date-picker--open"); got != 1 {
		t.Fatalf("open class count=%d want 1", got)
	}
}

func TestPickersFixDatePickerClosedClass(t *testing.T) {
	t.Parallel()
	rendered := renderSelectionHTML(t, chadcn.DatePicker(chadcn.DatePickerOptions{ID: "from-date"}))
	requireSelectionMarkup(t, rendered, `class="of-scope of-date-picker"`)
	if strings.Contains(rendered, "of-date-picker--open") {
		t.Fatalf("closed date picker rendered the open class:\n%s", rendered)
	}
}

func TestPickersFixStylesheetKeepsOpenPanelsOverlaid(t *testing.T) {
	t.Parallel()
	for _, selector := range []string{
		".of-combobox.of-combobox--open",
		".of-date-picker.of-date-picker--open",
	} {
		block := pickersFixRule(t, chadcn.Stylesheet, selector)
		if !strings.Contains(block, "z-index:") {
			t.Fatalf("rule %q does not raise the open root:\n%s", selector, block)
		}
	}
	for _, selector := range []string{
		".of-combobox.of-combobox--open .of-combobox-panel",
		".of-date-picker.of-date-picker--open .of-date-picker-panel",
	} {
		block := pickersFixRule(t, chadcn.Stylesheet, selector)
		for _, declaration := range []string{"position: absolute;", "top: 40px;", "width:"} {
			if !strings.Contains(block, declaration) {
				t.Fatalf("rule %q is missing %q:\n%s", selector, declaration, block)
			}
		}
	}
}

// pickersFixRule returns the declaration block of one stylesheet rule.
func pickersFixRule(t *testing.T, sheet, selector string) string {
	t.Helper()
	start := strings.Index(sheet, selector+" {")
	if start < 0 {
		t.Fatalf("stylesheet is missing rule %q", selector)
	}
	start += len(selector)
	end := strings.Index(sheet[start:], "}")
	if end < 0 {
		t.Fatalf("stylesheet rule %q has no closing brace", selector)
	}
	return sheet[start : start+end]
}
