package chadcn_test

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// Day buttons carry positional ids (calendar ID + cell index 0..41), so the
// 42-cell grid keeps its ids across months and a month change lands as text
// and attribute edits instead of a rename-driven full render. No callbacks
// are set here: outside days carry no handler by design, so a month change
// with OnSelect would flip handler presence on boundary cells and stay a
// full render.

// TestProbePatchFix2CalendarSameMonth locks in the same-month day change
// patch: only the two day buttons' attributes change.
func TestProbePatchFix2CalendarSameMonth(t *testing.T) {
	t.Parallel()
	month := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
	v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(
		chadcn.Calendar(chadcn.CalendarOptions{
			ID:       "cal-fix2",
			Month:    month,
			Selected: time.Date(2026, time.November, 4, 0, 0, 0, 0, time.UTC),
		}),
	))
	expectProbePatch(t, v, ctx, ui.Div().ID("probe-root").Children(
		chadcn.Calendar(chadcn.CalendarOptions{
			ID:       "cal-fix2",
			Month:    month,
			Selected: time.Date(2026, time.November, 5, 0, 0, 0, 0, time.UTC),
		}),
	))
}

// TestProbePatchFix2CalendarMonthChange proves a month change patches: the
// title, grid label, and day buttons keep their ids and only their text and
// attributes change.
func TestProbePatchFix2CalendarMonthChange(t *testing.T) {
	t.Parallel()
	v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(
		chadcn.Calendar(chadcn.CalendarOptions{
			ID:    "cal-fix2",
			Month: time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC),
		}),
	))
	expectProbePatch(t, v, ctx, ui.Div().ID("probe-root").Children(
		chadcn.Calendar(chadcn.CalendarOptions{
			ID:    "cal-fix2",
			Month: time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC),
		}),
	))
}

// TestProbePatchFix2CalendarNoop proves an identical rebuild does no work.
func TestProbePatchFix2CalendarNoop(t *testing.T) {
	t.Parallel()
	month := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
	v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(
		chadcn.Calendar(chadcn.CalendarOptions{
			ID:    "cal-fix2",
			Month: month,
		}),
	))
	expectProbeNoop(t, v, ctx, ui.Div().ID("probe-root").Children(
		chadcn.Calendar(chadcn.CalendarOptions{
			ID:    "cal-fix2",
			Month: month,
		}),
	))
}
