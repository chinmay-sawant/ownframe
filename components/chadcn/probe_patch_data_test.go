// Probe-patch tests for dashboard data changes (TestProbePatchData*).
//
// What patches and what stays FULL (see ui/diff.go: same tag, same child
// count, same order at every node; every changed node needs a stable id;
// a removed attribute forces a rebuild):
//
//   - Chart scaled values ([10,20] -> [20,40]) patch: bar percentages are
//     unchanged, so only the summary value is a text edit on its stable id.
//     Legend labels, axis labels, swatches, and stat spans all carry
//     chart-ID-derived ids; without them even this case falls back.
//   - Chart probe values ([10,20] -> [30,40]) stay FULL: bar0 grows 50% ->
//     75%. The engine cannot re-lay out an element that carries a percentage
//     height in isolation (even a custom-property-only touch of such a bar
//     falls back), so any bar resize needs engine support. Same class of
//     block as the slider/progress geometry skips in probe_patch_checks_test.
//   - AreaChart recolor patches: dot color is a custom-property edit and no
//     percentage height changes, so the legend and point dots patch on their
//     stable ids. AreaChart takes an ID option for this; without an ID it
//     invents nothing.
//   - AreaChart probe values stay FULL: every spacer and gap band height
//     resolves as a percentage, so any value change resizes them.
//   - StatCard value + delta patch: text edits on the stable spans.
//     StatCard takes an ID option for this; without an ID it is unchanged.
//   - FilterBar pill flip patches: class plus aria-pressed edits on the
//     stable pill id. The active tone is color-only.
//   - StatusSegments value + count changes patch: flex-grow edits on the
//     stable bars plus text edits on the stable counts.
//   - HexGrid tone change patches: a data-tone edit on the stable cell id.
//   - DataTable cell-text edit stays FULL: the engine cannot re-lay out
//     table-internal content in isolation, not even a class-only row
//     selection flip, so every same-shape content change needs engine
//     support. Row, cell, header, select, and empty ids are still stable and
//     load-bearing for hover matching. The caption keeps no id: its exact
//     markup is pinned by TestDataTableFixCaptionHeaderCellClasses.
//   - DataTable add-row stays FULL by design: one more row changes the tbody
//     child count, so no id scheme can address it. Added/removed series,
//     labels, cells, columns, pills, and segments, and an AreaChart gap-band
//     show/hide, behave the same way.
package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func probeDataRoot(el *ui.Element) *ui.Element {
	return ui.Div().ID("probe-root").Children(el)
}

func probeChart(values []float64) *ui.Element {
	return probeDataRoot(chadcn.Chart(chadcn.ChartOptions{
		ID: "ch-1", Title: "Traffic", Labels: []string{"Mon", "Tue"},
		Series: []chadcn.ChartSeries{{Label: "Hits", Values: values}},
	}))
}

// TestProbePatchData proves the dashboard data transitions that can patch do,
// and records the engine-blocked ones as skips in the style of the
// slider/progress geometry skips.
func TestProbePatchData(t *testing.T) {
	t.Parallel()

	t.Run("ChartValues", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, probeChart([]float64{10, 20}))
		s0 := v.Page.Probe().Stats()
		if err := v.Update(ctx, probeChart([]float64{30, 40})); err != nil {
			t.Fatal(err)
		}
		s1 := v.Page.Probe().Stats()
		t.Skipf("blocked: bar height 50%%->75%% resizes a percentage-height element and the engine "+
			"refuses such subtree patches (parses %d->%d); needs engine support", s0.Parses, s1.Parses)
	})

	t.Run("ChartScaledValues", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, probeChart([]float64{10, 20}))
		expectProbePatch(t, v, ctx, probeChart([]float64{20, 40}))
	})

	t.Run("AreaChartValues", func(t *testing.T) {
		t.Parallel()
		first := probeDataRoot(chadcn.AreaChart(chadcn.AreaChartOptions{
			ID: "ac-1", Title: "Backlog", Labels: []string{"Mon", "Tue"},
			Series: []chadcn.ChartSeries{
				{Label: "Open", Values: []float64{8, 6}},
				{Label: "Shut", Values: []float64{3, 5}},
			},
		}))
		v, ctx := mountProbeView(t, first)
		s0 := v.Page.Probe().Stats()
		next := probeDataRoot(chadcn.AreaChart(chadcn.AreaChartOptions{
			ID: "ac-1", Title: "Backlog", Labels: []string{"Mon", "Tue"},
			Series: []chadcn.ChartSeries{
				{Label: "Open", Values: []float64{9, 2}},
				{Label: "Shut", Values: []float64{4, 7}},
			},
		}))
		if err := v.Update(ctx, next); err != nil {
			t.Fatal(err)
		}
		s1 := v.Page.Probe().Stats()
		t.Skipf("blocked: spacer and band percentage heights resize and the engine "+
			"refuses such subtree patches (parses %d->%d); needs engine support", s0.Parses, s1.Parses)
	})

	t.Run("AreaChartRecolor", func(t *testing.T) {
		t.Parallel()
		mk := func(color string) *ui.Element {
			return probeDataRoot(chadcn.AreaChart(chadcn.AreaChartOptions{
				ID: "ac-1", Title: "Backlog", Labels: []string{"Mon", "Tue"},
				Series: []chadcn.ChartSeries{
					{Label: "Open", Values: []float64{8, 6}, Color: color},
					{Label: "Shut", Values: []float64{3, 5}},
				},
			}))
		}
		v, ctx := mountProbeView(t, mk(""))
		expectProbePatch(t, v, ctx, mk("#8b5cf6"))
	})

	t.Run("StatCardValue", func(t *testing.T) {
		t.Parallel()
		first := probeDataRoot(chadcn.StatCard(chadcn.StatCardOptions{
			ID: "sc-1", Label: "MRR", Value: "$9k", Delta: "+2%",
		}))
		v, ctx := mountProbeView(t, first)
		expectProbePatch(t, v, ctx, probeDataRoot(chadcn.StatCard(chadcn.StatCardOptions{
			ID: "sc-1", Label: "MRR", Value: "$11k", Delta: "+4%",
		})))
	})

	t.Run("FilterBarPill", func(t *testing.T) {
		t.Parallel()
		first := probeDataRoot(chadcn.FilterBar(chadcn.FilterBarOptions{
			ID: "fb-1",
			Options: []chadcn.FilterBarOption{
				{Label: "Open", Active: false},
				{Label: "Shut", Active: false},
			},
		}))
		v, ctx := mountProbeView(t, first)
		expectProbePatch(t, v, ctx, probeDataRoot(chadcn.FilterBar(chadcn.FilterBarOptions{
			ID: "fb-1",
			Options: []chadcn.FilterBarOption{
				{Label: "Open", Active: true},
				{Label: "Shut", Active: false},
			},
		})))
	})

	t.Run("StatusSegmentsValues", func(t *testing.T) {
		t.Parallel()
		first := probeDataRoot(chadcn.StatusSegments(chadcn.StatusSegmentsOptions{
			ID: "sg-1",
			Segments: []chadcn.StatusSegment{
				{Label: "Open", Value: 4, Count: "4"},
				{Label: "Shut", Value: 6, Count: "6"},
			},
		}))
		v, ctx := mountProbeView(t, first)
		expectProbePatch(t, v, ctx, probeDataRoot(chadcn.StatusSegments(chadcn.StatusSegmentsOptions{
			ID: "sg-1",
			Segments: []chadcn.StatusSegment{
				{Label: "Open", Value: 2, Count: "2"},
				{Label: "Shut", Value: 8, Count: "8"},
			},
		})))
	})

	t.Run("HexGridTone", func(t *testing.T) {
		t.Parallel()
		first := probeDataRoot(chadcn.HexGrid(chadcn.HexGridOptions{
			ID: "hx-1",
			Cells: []chadcn.HexCell{
				{Label: "a", Tone: 1},
				{Label: "b", Tone: 2},
			},
		}))
		v, ctx := mountProbeView(t, first)
		expectProbePatch(t, v, ctx, probeDataRoot(chadcn.HexGrid(chadcn.HexGridOptions{
			ID: "hx-1",
			Cells: []chadcn.HexCell{
				{Label: "a", Tone: 3},
				{Label: "b", Tone: 2},
			},
		})))
	})

	t.Run("DataTableCellEdit", func(t *testing.T) {
		t.Parallel()
		columns := []chadcn.DataColumn{{Label: "ID"}, {Label: "Total"}}
		first := probeDataRoot(chadcn.DataTable(chadcn.DataTableOptions{
			ID: "dt-1", Caption: "Orders", Columns: columns,
			Rows: [][]string{{"1", "$9"}, {"2", "$14"}},
		}))
		v, ctx := mountProbeView(t, first)
		s0 := v.Page.Probe().Stats()
		next := probeDataRoot(chadcn.DataTable(chadcn.DataTableOptions{
			ID: "dt-1", Caption: "Orders", Columns: columns,
			Rows: [][]string{{"1", "$9"}, {"2", "$15"}},
		}))
		if err := v.Update(ctx, next); err != nil {
			t.Fatal(err)
		}
		s1 := v.Page.Probe().Stats()
		t.Skipf("blocked: table-internal text edits cannot be re-laid out in isolation and the engine "+
			"refuses such subtree patches (parses %d->%d); needs engine support", s0.Parses, s1.Parses)
	})

	t.Run("DataTableAddRowFull", func(t *testing.T) {
		t.Parallel()
		columns := []chadcn.DataColumn{{Label: "ID"}, {Label: "Total"}}
		first := probeDataRoot(chadcn.DataTable(chadcn.DataTableOptions{
			ID: "dt-1", Caption: "Orders", Columns: columns,
			Rows: [][]string{{"1", "$9"}, {"2", "$14"}},
		}))
		v, ctx := mountProbeView(t, first)
		gen0 := v.Page.Probe().Generation()
		s0 := v.Page.Probe().Stats()
		next := probeDataRoot(chadcn.DataTable(chadcn.DataTableOptions{
			ID: "dt-1", Caption: "Orders", Columns: columns,
			Rows: [][]string{{"1", "$9"}, {"2", "$14"}, {"3", "$21"}},
		}))
		if err := v.Update(ctx, next); err != nil {
			t.Fatal(err)
		}
		if got, want := v.Page.Probe().Generation(), gen0+1; got != want {
			t.Fatalf("generation = %d, want %d", got, want)
		}
		s1 := v.Page.Probe().Stats()
		if s1.Parses == s0.Parses {
			t.Fatalf("add-row patched, want the full-render path: parses stayed %d", s1.Parses)
		}
	})
}
