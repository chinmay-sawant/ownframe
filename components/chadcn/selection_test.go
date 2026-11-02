package chadcn_test

import (
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// renderSelectionHTML renders one selection component for markup assertions.
func renderSelectionHTML(t *testing.T, root *ui.Element) string {
	t.Helper()
	rendered, err := ui.Render(root)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return rendered
}

func requireSelectionMarkup(t *testing.T, rendered string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(rendered, fragment) {
			t.Fatalf("rendered markup is missing %q:\n%s", fragment, rendered)
		}
	}
}

func TestSelectOpenRendersListbox(t *testing.T) {
	t.Parallel()
	rendered := renderSelectionHTML(t, chadcn.Select(chadcn.SelectOptions{
		ID:          "fruit",
		Open:        true,
		Value:       "banana",
		Placeholder: "Select a fruit",
		Items: []chadcn.SelectItem{
			{Value: "apple", Label: "Apple"},
			{Value: "banana", Label: "Banana"},
			{Value: "cherry", Label: "Cherry", Disabled: true},
		},
	}))
	requireSelectionMarkup(t, rendered,
		`class="of-scope of-select"`,
		`id="fruit"`,
		`class="of-select-trigger of-transition"`,
		`id="fruit-trigger"`,
		`aria-controls="fruit-panel"`,
		`aria-expanded="true"`,
		`id="fruit-panel"`,
		`role="listbox"`,
		`class="of-select-label of-select-value"`,
		`>Banana</span>`,
		`id="fruit-option-0"`,
		`role="option"`,
		`aria-selected="true"`,
		`aria-selected="false"`,
		`class="of-select-check"`,
		`disabled="disabled"`,
	)
	if got := strings.Count(rendered, `role="option"`); got != 3 {
		t.Fatalf("options=%d want 3", got)
	}
	if got := strings.Count(rendered, `aria-selected="true"`); got != 1 {
		t.Fatalf("selected options=%d want 1", got)
	}
}

func TestSelectClosedRendersTriggerOnly(t *testing.T) {
	t.Parallel()
	rendered := renderSelectionHTML(t, chadcn.Select(chadcn.SelectOptions{
		ID:          "fruit",
		Placeholder: "Select a fruit",
		Items:       []chadcn.SelectItem{{Value: "apple", Label: "Apple"}},
	}))
	requireSelectionMarkup(t, rendered,
		`class="of-select-trigger of-transition"`,
		`aria-expanded="false"`,
		`class="of-select-label of-select-placeholder"`,
		`>Select a fruit</span>`,
		`>⌄</span>`,
	)
	if strings.Contains(rendered, `role="listbox"`) {
		t.Fatalf("closed select rendered a listbox:\n%s", rendered)
	}
}

func TestComboboxOpenFiltersItems(t *testing.T) {
	t.Parallel()
	rendered := renderSelectionHTML(t, chadcn.Combobox(chadcn.ComboboxOptions{
		ID:          "framework",
		Open:        true,
		Query:       "svelte",
		Placeholder: "Select a framework",
		Items: []chadcn.ComboboxItem{
			{Value: "next", Label: "Next.js"},
			{Value: "svelte", Label: "SvelteKit"},
			{Value: "astro", Label: "Astro"},
		},
	}))
	requireSelectionMarkup(t, rendered,
		`class="of-scope of-combobox"`,
		`id="framework"`,
		`class="of-combobox-input of-transition"`,
		`id="framework-input"`,
		`role="combobox"`,
		`aria-controls="framework-panel"`,
		`aria-expanded="true"`,
		`placeholder="Select a framework"`,
		`value="svelte"`,
		`id="framework-trigger"`,
		`id="framework-panel"`,
		`role="listbox"`,
		`>SvelteKit</span>`,
	)
	if strings.Contains(rendered, "Next.js") || strings.Contains(rendered, "Astro") {
		t.Fatalf("combobox did not filter the list:\n%s", rendered)
	}
	if got := strings.Count(rendered, `role="option"`); got != 1 {
		t.Fatalf("filtered options=%d want 1", got)
	}
}

func TestComboboxEmptyState(t *testing.T) {
	t.Parallel()
	rendered := renderSelectionHTML(t, chadcn.Combobox(chadcn.ComboboxOptions{
		ID:    "framework",
		Open:  true,
		Query: "zzz",
		Items: []chadcn.ComboboxItem{{Value: "next", Label: "Next.js"}},
	}))
	requireSelectionMarkup(t, rendered,
		`class="of-combobox-empty"`,
		`>No results</div>`,
		`role="listbox"`,
	)
	if strings.Contains(rendered, `role="option"`) {
		t.Fatalf("empty combobox rendered options:\n%s", rendered)
	}
}

func TestComboboxClosedShowsSelection(t *testing.T) {
	t.Parallel()
	rendered := renderSelectionHTML(t, chadcn.Combobox(chadcn.ComboboxOptions{
		ID:    "framework",
		Value: "svelte",
		Items: []chadcn.ComboboxItem{{Value: "svelte", Label: "SvelteKit"}},
	}))
	requireSelectionMarkup(t, rendered,
		`aria-expanded="false"`,
		`value="SvelteKit"`,
	)
	if strings.Contains(rendered, `role="listbox"`) {
		t.Fatalf("closed combobox rendered a listbox:\n%s", rendered)
	}
}

func TestCalendarFixedMonth(t *testing.T) {
	t.Parallel()
	month := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
	selected := time.Date(2026, time.November, 3, 0, 0, 0, 0, time.UTC)
	rendered := renderSelectionHTML(t, chadcn.Calendar(chadcn.CalendarOptions{
		Month:    month,
		Selected: selected,
	}))
	requireSelectionMarkup(t, rendered,
		`class="of-scope of-calendar"`,
		`class="of-calendar-title"`,
		`>November 2026</div>`,
		`class="of-calendar-weekday"`,
		`aria-label="Sunday"`,
		`aria-label="Saturday"`,
		`aria-label="Previous month"`,
		`aria-label="Next month"`,
		`role="grid"`,
		`aria-label="November 2026"`,
		`class="of-calendar-day of-transition of-calendar-day--selected"`,
		`of-calendar-day--outside`,
		`aria-selected="true"`,
		`>30</button>`,
	)
	if got := strings.Count(rendered, `role="gridcell"`); got != 42 {
		t.Fatalf("days=%d want 42", got)
	}
	if got := strings.Count(rendered, `class="of-calendar-week"`); got != 6 {
		t.Fatalf("weeks=%d want 6", got)
	}
	if got := strings.Count(rendered, `aria-selected="true"`); got != 1 {
		t.Fatalf("selected days=%d want 1", got)
	}
	if strings.Contains(rendered, `>31</button>`) {
		t.Fatalf("November rendered a 31st day:\n%s", rendered)
	}
}

func TestCalendarLeapYear(t *testing.T) {
	t.Parallel()
	rendered := renderSelectionHTML(t, chadcn.Calendar(chadcn.CalendarOptions{
		Month: time.Date(2028, time.February, 1, 0, 0, 0, 0, time.UTC),
	}))
	requireSelectionMarkup(t, rendered,
		`>February 2028</div>`,
		`>29</button>`,
	)
	if got := strings.Count(rendered, `role="gridcell"`); got != 42 {
		t.Fatalf("days=%d want 42", got)
	}
}

func TestDatePickerOpenShowsCalendar(t *testing.T) {
	t.Parallel()
	selected := time.Date(2026, time.November, 3, 0, 0, 0, 0, time.UTC)
	rendered := renderSelectionHTML(t, chadcn.DatePicker(chadcn.DatePickerOptions{
		ID:       "due",
		Open:     true,
		Month:    selected,
		Selected: selected,
	}))
	requireSelectionMarkup(t, rendered,
		`class="of-scope of-date-picker"`,
		`id="due"`,
		`class="of-date-picker-trigger of-transition"`,
		`id="due-trigger"`,
		`aria-controls="due-panel"`,
		`aria-expanded="true"`,
		`class="of-date-picker-label of-date-picker-value"`,
		`>November 3, 2026</span>`,
		`id="due-panel"`,
		`role="dialog"`,
		`class="of-scope of-calendar"`,
		`>November 2026</div>`,
	)
	if got := strings.Count(rendered, `role="gridcell"`); got != 42 {
		t.Fatalf("days=%d want 42", got)
	}
}

func TestDatePickerClosedShowsPlaceholder(t *testing.T) {
	t.Parallel()
	rendered := renderSelectionHTML(t, chadcn.DatePicker(chadcn.DatePickerOptions{ID: "due"}))
	requireSelectionMarkup(t, rendered,
		`class="of-date-picker-trigger of-transition"`,
		`aria-expanded="false"`,
		`class="of-date-picker-label of-date-picker-placeholder"`,
		`>Pick a date</span>`,
	)
	if strings.Contains(rendered, `role="dialog"`) || strings.Contains(rendered, "of-calendar") {
		t.Fatalf("closed date picker rendered a panel:\n%s", rendered)
	}
}
