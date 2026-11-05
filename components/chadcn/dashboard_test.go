package chadcn_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
)

// TestDashboardStatCardRendersValueAndDelta checks the KPI cell carries its
// label, headline value, and tone pill.
func TestDashboardStatCardRendersValueAndDelta(t *testing.T) {
	t.Parallel()
	source := feedbackHTML(t, chadcn.StatCard(chadcn.StatCardOptions{
		Label: "Open work orders",
		Value: "137",
		Delta: "23 urgent",
		Tone:  chadcn.StatDeltaWarn,
	}))
	for _, want := range []string{
		`class="of-stat-card"`,
		`class="of-stat-card-label"`,
		`class="of-stat-card-value"`,
		`class="of-stat-card-delta of-stat-card-delta--warn"`,
		"Open work orders",
		"137",
		"23 urgent",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("stat card missing %q in\n%s", want, source)
		}
	}
}

// TestDashboardSidebarSectionGroupsItems checks grouped navigation renders a
// section label with its rows.
func TestDashboardSidebarSectionGroupsItems(t *testing.T) {
	t.Parallel()
	source := feedbackHTML(t, chadcn.SidebarSection("Operations",
		chadcn.SidebarNavItem(chadcn.SidebarItem{Label: "Dashboard", Active: true}),
		chadcn.SidebarNavItem(chadcn.SidebarItem{Label: "Work orders", Badge: "137"}),
	))
	for _, want := range []string{
		`class="of-sidebar-section"`,
		`class="of-sidebar-section-label"`,
		"Operations",
		"Dashboard",
		"Work orders",
		"137",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("sidebar section missing %q in\n%s", want, source)
		}
	}
}

// TestDashboardFilterBarRendersLegendAndActivePill checks the tab bar above a
// plot keeps its legend and marks the active pill with aria-pressed.
func TestDashboardFilterBarRendersLegendAndActivePill(t *testing.T) {
	t.Parallel()
	source := feedbackHTML(t, chadcn.FilterBar(chadcn.FilterBarOptions{
		Label:  "Trend filters",
		Legend: []chadcn.FilterBarLegend{{Label: "Opened", Tone: 1}, {Label: "Resolved", Tone: 2}},
		Options: []chadcn.FilterBarOption{
			{Label: "Weekly", Active: true},
			{Label: "Monthly"},
		},
	}))
	for _, want := range []string{
		`class="of-filter-bar"`,
		`aria-label="Trend filters"`,
		`class="of-filter-bar-legend"`,
		`class="of-filter-bar-pill of-filter-bar-pill--active"`,
		`aria-pressed="true"`,
		`aria-pressed="false"`,
		"Opened",
		"Resolved",
		"Weekly",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("filter bar missing %q in\n%s", want, source)
		}
	}
}

// TestDashboardAreaChartRendersPointsAndBand checks the trend plot renders one
// column per step, a dot per value, and a gap band between extremes.
func TestDashboardAreaChartRendersPointsAndBand(t *testing.T) {
	t.Parallel()
	source := feedbackHTML(t, chadcn.AreaChart(chadcn.AreaChartOptions{
		Title: "Maintenance overview",
		Series: []chadcn.ChartSeries{
			{Label: "Opened", Values: []float64{10, 20}},
			{Label: "Resolved", Values: []float64{6, 14}},
		},
		Labels: []string{"Oct 12", "Oct 13"},
	}))
	for _, want := range []string{
		`class="of-area-chart"`,
		`class="of-area-chart-plot"`,
		`class="of-area-chart-column"`,
		`class="of-area-chart-track"`,
		`class="of-area-chart-band"`,
		`class="of-area-chart-spacer"`,
		`class="of-area-chart-point"`,
		`class="of-area-chart-label"`,
		"Maintenance overview",
		"Oct 12",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("area chart missing %q in\n%s", want, source)
		}
	}
	if got := strings.Count(source, `class="of-area-chart-column"`); got != 2 {
		t.Fatalf("columns=%d, want 2 in\n%s", got, source)
	}
	if got := strings.Count(source, `class="of-area-chart-point"`); got != 4 {
		t.Fatalf("points=%d, want 4 in\n%s", got, source)
	}
	if got := strings.Count(source, `class="of-area-chart-band"`); got != 2 {
		t.Fatalf("bands=%d, want 2 in\n%s", got, source)
	}
}

// TestDashboardHexGridRendersCellsByTone checks the honeycomb map renders one
// cell per entry with its tone slot.
func TestDashboardHexGridRendersCellsByTone(t *testing.T) {
	t.Parallel()
	source := feedbackHTML(t, chadcn.HexGrid(chadcn.HexGridOptions{
		Label:   "Portfolio health",
		Columns: 4,
		Cells: []chadcn.HexCell{
			{Label: "Plumbing", Tone: 1},
			{Label: "HVAC", Tone: 3},
			{Label: "Empty", Tone: 0},
		},
	}))
	for _, want := range []string{
		`class="of-hex-grid"`,
		`aria-label="Portfolio health"`,
		`class="of-hex"`,
		`class="of-hex-top"`,
		`class="of-hex-mid"`,
		`class="of-hex-bottom"`,
		`data-tone="1"`,
		`data-tone="3"`,
		`data-tone="0"`,
		"--of-hex-columns:4",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("hex grid missing %q in\n%s", want, source)
		}
	}
	if got := strings.Count(source, `class="of-hex-top"`); got != 3 {
		t.Fatalf("cells=%d, want 3 in\n%s", got, source)
	}
}

// TestDashboardStatusSegmentsRendersProportionalBars checks the status row
// renders one proportional bar per segment with its count legend.
func TestDashboardStatusSegmentsRendersProportionalBars(t *testing.T) {
	t.Parallel()
	source := feedbackHTML(t, chadcn.StatusSegments(chadcn.StatusSegmentsOptions{
		Label:   "Work order status",
		Overdue: "Overdue work orders: 9",
		Segments: []chadcn.StatusSegment{
			{Label: "New", Value: 38, Count: "38", Tone: 1},
			{Label: "In progress", Value: 54, Count: "54", Tone: 3},
		},
	}))
	for _, want := range []string{
		`class="of-status-segments"`,
		`aria-label="Work order status"`,
		`class="of-status-segments-track"`,
		`class="of-status-segments-bar"`,
		`class="of-status-segments-count"`,
		"Overdue work orders: 9",
		"New",
		"In progress",
		"38",
		"54",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("status segments missing %q in\n%s", want, source)
		}
	}
	if got := strings.Count(source, `class="of-status-segments-bar"`); got != 2 {
		t.Fatalf("bars=%d, want 2 in\n%s", got, source)
	}
}
