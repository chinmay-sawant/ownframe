package chadcn

import (
	"math"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	chartPercent           = 100
	chartDigitGroup        = 3
	chartRoundingTolerance = 0.05
	chartPaletteSize       = 4
)

// ChartSeries is one measured series in a Chart. Each value lines up with the
// same index in ChartOptions.Labels. Color is an optional CSS color; an empty
// Color uses this set's chart palette.
type ChartSeries struct {
	Label  string
	Values []float64
	Color  string
}

// ChartOptions describes a chart frame. Every series shares Labels. Total is an
// optional headline figure, and zero hides it. When ID is set, each bar id is
// that value plus the bar index, in plot order, and the legend labels, axis
// labels, and summary values take stable suffixed ids, so a same-shape data
// change (same series, labels, and counts, only values) patches instead of
// taking the full-render path. Structural changes stay FULL by design.
type ChartOptions struct {
	ID          string
	Title       string
	Description string
	Series      []ChartSeries
	Labels      []string
	Total       float64
}

// Chart builds a bar-chart frame: a title and description, per-series totals, a
// CSS bar plot with axis labels, and a legend under the plot. Bars are plain div
// geometry, so the page needs no canvas or SVG. The host owns the data and
// rebuilds the tree when it changes.
func Chart(options ChartOptions) *ui.Element {
	root := ui.Section().Class("of-chart").Attr("role", "group").Attr("aria-label", chartLabel(options))
	if options.ID != "" {
		root.ID(options.ID)
	}
	if header := chartHeader(options); header != nil {
		root.Children(header)
	}
	if body := chartBody(options); body != nil {
		root.Children(body)
	}
	return root
}

func chartHeader(options ChartOptions) *ui.Element {
	hasHeading := options.Title != "" || options.Description != ""
	hasSummary := options.Total > 0 || len(options.Series) > 0
	if !hasHeading && !hasSummary {
		return nil
	}
	header := ui.Header().Class("of-chart-header")
	if hasHeading {
		heading := ui.Div().Class("of-chart-heading")
		if options.Title != "" {
			title := ui.Tag("h2").Class("of-chart-title").Text(options.Title)
			if id := partID(options.ID, "-title"); id != "" {
				title.ID(id)
			}
			heading.Children(title)
		}
		if options.Description != "" {
			description := ui.P().Class("of-chart-description").Text(options.Description)
			if id := partID(options.ID, "-description"); id != "" {
				description.ID(id)
			}
			heading.Children(description)
		}
		header.Children(heading)
	}
	if hasSummary {
		header.Children(chartSummary(options))
	}
	return header
}

func chartSummary(options ChartOptions) *ui.Element {
	summary := ui.Div().Class("of-chart-summary")
	if options.Total > 0 {
		summary.Children(chartStat(options.ID, "-total", "Total", options.Total))
	}
	for index, series := range options.Series {
		summary.Children(chartStat(options.ID, "-stat-"+strconv.Itoa(index), series.Label, chartSum(series.Values)))
	}
	return summary
}

func chartStat(chartID, suffix, label string, value float64) *ui.Element {
	labelSpan := ui.Span().Class("of-chart-stat-label").Text(label)
	valueSpan := ui.Span().Class("of-chart-stat-value").Text(chartNumber(value))
	if id := partID(chartID, suffix+"-label"); id != "" {
		labelSpan.ID(id)
	}
	if id := partID(chartID, suffix+"-value"); id != "" {
		valueSpan.ID(id)
	}
	return ui.Div().Class("of-chart-stat").Children(labelSpan, valueSpan)
}

// chartBody groups the plot, its axis labels, and the legend in one padded
// container, so the three share the same horizontal inset.
func chartBody(options ChartOptions) *ui.Element {
	groups := chartGroups(options.Series)
	legend := chartLegend(options.ID, options.Series)
	if groups == 0 && legend == nil {
		return nil
	}
	body := ui.Div().Class("of-chart-body")
	if groups > 0 {
		body.Children(chartPlot(options.ID, options.Series, groups))
		if len(options.Labels) > 0 {
			body.Children(chartLabels(options.ID, options.Labels, groups))
		}
	}
	if legend != nil {
		body.Children(legend)
	}
	return body
}

func chartPlot(chartID string, series []ChartSeries, groups int) *ui.Element {
	plot := ui.Div().Class("of-chart-plot").Attr("aria-hidden", "true")
	maximum := chartMaximum(series)
	barIndex := 0
	for index := 0; index < groups; index++ {
		group := ui.Div().Class("of-chart-group")
		for seriesIndex, item := range series {
			if index >= len(item.Values) {
				continue
			}
			bar := ui.Div().Class("of-chart-bar").Attr("data-series", chartSeriesSlot(seriesIndex)).
				Height(ui.Percent(chartHeight(item.Values[index], maximum)))
			if item.Color != "" {
				bar.CustomProperty("--of-chart-bar", item.Color)
			}
			if id := partID(chartID, "-bar-"+strconv.Itoa(barIndex)); id != "" {
				bar.ID(id)
			}
			barIndex++
			group.Children(bar)
		}
		plot.Children(group)
	}
	return plot
}

func chartLabels(chartID string, labels []string, groups int) *ui.Element {
	row := ui.Div().Class("of-chart-labels")
	for index := 0; index < groups; index++ {
		label := ""
		if index < len(labels) {
			label = labels[index]
		}
		span := ui.Span().Class("of-chart-label").Text(label)
		if id := partID(chartID, "-label-"+strconv.Itoa(index)); id != "" {
			span.ID(id)
		}
		row.Children(span)
	}
	return row
}

func chartLegend(chartID string, series []ChartSeries) *ui.Element {
	if len(series) == 0 {
		return nil
	}
	legend := ui.Div().Class("of-chart-legend")
	for index, item := range series {
		swatch := ui.Span().Class("of-chart-swatch").Attr("aria-hidden", "true").
			Attr("data-series", chartSeriesSlot(index))
		if item.Color != "" {
			swatch.CustomProperty("--of-chart-bar", item.Color)
		}
		if id := partID(chartID, "-legend-"+strconv.Itoa(index)+"-swatch"); id != "" {
			swatch.ID(id)
		}
		label := ui.Span().Class("of-chart-legend-label").Text(item.Label)
		if id := partID(chartID, "-legend-"+strconv.Itoa(index)+"-label"); id != "" {
			label.ID(id)
		}
		legend.Children(ui.Div().Class("of-chart-legend-item").Children(swatch, label))
	}
	return legend
}

// chartGroups returns the widest series so every label keeps its column.
func chartGroups(series []ChartSeries) int {
	groups := 0
	for _, item := range series {
		groups = max(groups, len(item.Values))
	}
	return groups
}

func chartMaximum(series []ChartSeries) float64 {
	maximum := 0.0
	for _, item := range series {
		for _, value := range item.Values {
			if finite(value) && value > maximum {
				maximum = value
			}
		}
	}
	return maximum
}

func chartHeight(value, maximum float64) float64 {
	if maximum <= 0 || !finite(value) || value <= 0 {
		return 0
	}
	return min(value/maximum*chartPercent, chartPercent)
}

func chartSeriesSlot(index int) string {
	return strconv.Itoa(index%chartPaletteSize + 1)
}

func chartSum(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		if finite(value) {
			total += value
		}
	}
	return total
}

func chartNumber(value float64) string {
	if !finite(value) {
		return "0"
	}
	rounded := math.Round(value)
	if math.Abs(value-rounded) < chartRoundingTolerance {
		return chartGroupDigits(strconv.FormatInt(int64(rounded), 10))
	}
	return strconv.FormatFloat(value, 'f', 1, 64)
}

func chartGroupDigits(text string) string {
	sign := ""
	if strings.HasPrefix(text, "-") {
		sign, text = "-", strings.TrimPrefix(text, "-")
	}
	var out strings.Builder
	for index, digit := range text {
		if index > 0 && (len(text)-index)%chartDigitGroup == 0 {
			out.WriteByte(',')
		}
		out.WriteRune(digit)
	}
	return sign + out.String()
}

func chartLabel(options ChartOptions) string {
	label := "Bar chart"
	if options.Title != "" {
		label += ": " + options.Title
	}
	if options.Description != "" {
		label += ". " + options.Description
	}
	return label
}
