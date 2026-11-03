package chadcn_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
)

// TestChartFixRenderKeepsPlotBarsAndLegend checks the tightened chart still
// renders one plot, a bar per value, and a legend inside the chart body. Bars
// keep percentage heights so the host data contract does not change.
func TestChartFixRenderKeepsPlotBarsAndLegend(t *testing.T) {
	t.Parallel()
	source := feedbackHTML(t, chadcn.Chart(chadcn.ChartOptions{
		Title:       "Request volume",
		Description: "Requests across all services.",
		Series: []chadcn.ChartSeries{
			{Label: "This period", Values: []float64{100, 50, 25}},
			{Label: "Previous period", Values: []float64{80, 40, 20}, Color: "#8b5cf6"},
		},
		Labels: []string{"00:00", "04:00", "08:00"},
	}))
	for _, want := range []string{
		`class="of-chart-header"`,
		`class="of-chart-title"`,
		`class="of-chart-description"`,
		`class="of-chart-summary"`,
		`class="of-chart-stat"`,
		`class="of-chart-body"`,
		`aria-hidden="true" class="of-chart-plot"`,
		`class="of-chart-group"`,
		`class="of-chart-bar"`,
		`class="of-chart-labels"`,
		`class="of-chart-label"`,
		`class="of-chart-legend"`,
		`class="of-chart-legend-item"`,
		`class="of-chart-swatch"`,
		`class="of-chart-legend-label"`,
		`role="group"`,
		`aria-label="Bar chart: Request volume. Requests across all services."`,
		`data-series="1"`,
		`data-series="2"`,
		`--of-chart-bar:#8b5cf6`,
		"height:100%",
		"height:50%",
		"height:25%",
		"height:80%",
		"height:40%",
		"height:20%",
		"Request volume",
		"This period",
		"Previous period",
		"00:00",
		"08:00",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("chart missing %q in\n%s", want, source)
		}
	}
	for _, tt := range []struct {
		name string
		got  int
		want int
	}{
		{"groups", strings.Count(source, `class="of-chart-group"`), 3},
		{"bars", strings.Count(source, `class="of-chart-bar"`), 6},
		{"labels", strings.Count(source, `class="of-chart-label"`), 3},
		{"legend items", strings.Count(source, `class="of-chart-legend-item"`), 2},
		{"legend swatches", strings.Count(source, `class="of-chart-swatch"`), 2},
	} {
		if tt.got != tt.want {
			t.Fatalf("%s=%d, want %d in\n%s", tt.name, tt.got, tt.want, source)
		}
	}

	header := strings.Index(source, `class="of-chart-header"`)
	body := strings.Index(source, `class="of-chart-body"`)
	plot := strings.Index(source, `class="of-chart-plot"`)
	labels := strings.Index(source, `class="of-chart-labels"`)
	legend := strings.Index(source, `class="of-chart-legend"`)
	if header >= body || body >= plot || plot >= labels || labels >= legend {
		t.Fatalf("chart order is header, body, plot, labels, legend:\n%s", source)
	}
}

// TestChartFixStylesheetTightensSpacing pins the plot height, bar widths, and
// the reduced padding around the labels and legend.
func TestChartFixStylesheetTightensSpacing(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		selector string
		wants    []string
	}{
		{
			selector: ".of-chart-header",
			wants:    []string{"padding: var(--of-chart-header-padding, 14px 20px 12px);"},
		},
		{
			selector: ".of-chart-heading",
			wants:    []string{"gap: 2px;"},
		},
		{
			selector: ".of-chart-summary",
			wants:    []string{"gap: 18px;"},
		},
		{
			selector: ".of-chart-body",
			wants:    []string{"padding: var(--of-chart-body-padding, 14px 20px 4px);"},
		},
		{
			selector: ".of-chart-plot",
			wants: []string{
				"height: var(--of-chart-plot-height, 160px);",
				"padding: 0;",
				"gap: var(--of-chart-group-gap, 10px);",
			},
		},
		{
			selector: ".of-chart-group",
			wants:    []string{"gap: var(--of-chart-bar-gap, 4px);", "height: 100%;"},
		},
		{
			selector: ".of-chart-bar",
			wants: []string{
				"width: 100%;",
				"min-width: var(--of-chart-bar-min-width, 12px);",
				"max-width: var(--of-chart-bar-width, 22px);",
			},
		},
		{
			selector: ".of-chart-labels",
			wants:    []string{"gap: var(--of-chart-group-gap, 10px);", "padding-top: 6px;"},
		},
		{
			selector: ".of-chart-legend",
			wants:    []string{"gap: 12px;", "padding: var(--of-chart-legend-padding, 2px 0 8px);"},
		},
	} {
		rule := chartFixRule(t, chadcn.Stylesheet, tt.selector)
		for _, want := range tt.wants {
			if !strings.Contains(rule, want) {
				t.Fatalf("rule %s missing %q in %q", tt.selector, want, rule)
			}
		}
	}
}

// chartFixRule returns one rule body from the stylesheet, without its braces.
func chartFixRule(t *testing.T, sheet, selector string) string {
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
