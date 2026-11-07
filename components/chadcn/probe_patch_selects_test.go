package chadcn_test

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// The TestProbePatchSelects* tests prove one View.Update each takes the
// ui.View.Update patch path (parses+0) instead of the full-render fallback.
//
// What stays a full render by design, and why:
//   - Select closed->open and DatePicker closed->open add or remove the
//     panel child, so the child counts differ and the diff bails out.
//   - Combobox query-filter changes add or remove option children for the
//     same reason.
//   - Calendar month changes rewrite the title and all 42 day cells; day
//     buttons carry positional ids (see probe_patch_fix2_calendar_test.go),
//     so plain month switches patch, while month switches that flip
//     OnSelect handler presence on boundary cells stay full.
func probeSelectItems() []chadcn.SelectItem {
	return []chadcn.SelectItem{
		{Value: "a", Label: "Alpha"},
		{Value: "b", Label: "Beta"},
	}
}

func probeComboboxItems() []chadcn.ComboboxItem {
	return []chadcn.ComboboxItem{
		{Value: "next", Label: "Next.js"},
		{Value: "svelte", Label: "SvelteKit"},
	}
}

// TestProbePatchSelectsClosedValueChange proves a closed Select value change
// patches the trigger label's text instead of reparsing.
func TestProbePatchSelectsClosedValueChange(t *testing.T) {
	t.Parallel()
	v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(
		chadcn.Select(chadcn.SelectOptions{
			ID: "sel-1", Open: false, Value: "a", Items: probeSelectItems(),
		}),
	))
	expectProbePatch(t, v, ctx, ui.Div().ID("probe-root").Children(
		chadcn.Select(chadcn.SelectOptions{
			ID: "sel-1", Open: false, Value: "b", Items: probeSelectItems(),
		}),
	))
}

// TestProbePatchSelectsComboboxSameList proves moving the selection marker
// within one unchanged filtered list patches the two options instead of
// reparsing. Query-filter list changes stay a full render by design.
func TestProbePatchSelectsComboboxSameList(t *testing.T) {
	t.Parallel()
	v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(
		chadcn.Combobox(chadcn.ComboboxOptions{
			ID: "cb-1", Open: true, Query: "", Value: "next", Items: probeComboboxItems(),
		}),
	))
	expectProbePatch(t, v, ctx, ui.Div().ID("probe-root").Children(
		chadcn.Combobox(chadcn.ComboboxOptions{
			ID: "cb-1", Open: true, Query: "", Value: "svelte", Items: probeComboboxItems(),
		}),
	))
}

// TestProbePatchSelectsCalendarSameMonth proves a same-month day change
// patches the two day buttons' attributes instead of reparsing. A month
// change renames every day id and stays a full render by design.
func TestProbePatchSelectsCalendarSameMonth(t *testing.T) {
	t.Parallel()
	month := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
	v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(
		chadcn.Calendar(chadcn.CalendarOptions{
			ID:       "cal-1",
			Month:    month,
			Selected: time.Date(2026, time.November, 4, 0, 0, 0, 0, time.UTC),
		}),
	))
	expectProbePatch(t, v, ctx, ui.Div().ID("probe-root").Children(
		chadcn.Calendar(chadcn.CalendarOptions{
			ID:       "cal-1",
			Month:    month,
			Selected: time.Date(2026, time.November, 5, 0, 0, 0, 0, time.UTC),
		}),
	))
}

// TestProbePatchSelectsDatePickerClosedLabel proves a closed date change
// patches the trigger label's text instead of reparsing. Closed->open adds
// the panel child and stays a full render by design.
func TestProbePatchSelectsDatePickerClosedLabel(t *testing.T) {
	t.Parallel()
	month := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
	v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(
		chadcn.DatePicker(chadcn.DatePickerOptions{
			ID:       "dp-1",
			Open:     false,
			Month:    month,
			Selected: time.Date(2026, time.November, 4, 0, 0, 0, 0, time.UTC),
		}),
	))
	expectProbePatch(t, v, ctx, ui.Div().ID("probe-root").Children(
		chadcn.DatePicker(chadcn.DatePickerOptions{
			ID:       "dp-1",
			Open:     false,
			Month:    month,
			Selected: time.Date(2026, time.November, 5, 0, 0, 0, 0, time.UTC),
		}),
	))
}
