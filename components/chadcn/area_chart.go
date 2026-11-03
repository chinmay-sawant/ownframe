package chadcn

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

// AreaChartOptions describes a two-series trend with a hatched gap band. Each
// series lines up with AreaChartOptions.Labels by index. The gap band fills
// the space between the highest and lowest value in each column, which reads
// as the backlog between opened and resolved work.
type AreaChartOptions struct {
	Title       string
	Description string
	Series      []ChartSeries
	Labels      []string
}

// AreaChart builds a div-geometry trend plot: a header with a legend, columns
// with one dot per series, a hatched band between the column extremes, and
// axis labels. No canvas or SVG is needed.
func AreaChart(options AreaChartOptions) *ui.Element {
	root := ui.Div().Class("of-area-chart").Attr("role", "group").Attr("aria-label", areaChartLabel(options))
	if header := areaChartHeader(options); header != nil {
		root.Children(header)
	}
	if body := areaChartBody(options); body != nil {
		root.Children(body)
	}
	return root
}

func areaChartHeader(options AreaChartOptions) *ui.Element {
	if options.Title == "" && options.Description == "" && len(options.Series) == 0 {
		return nil
	}
	header := ui.Div().Class("of-area-chart-header")
	if options.Title != "" || options.Description != "" {
		heading := ui.Div().Class("of-area-chart-heading")
		if options.Title != "" {
			heading.Children(ui.Tag("h3").Class("of-area-chart-title").Text(options.Title))
		}
		if options.Description != "" {
			heading.Children(ui.P().Class("of-area-chart-description").Text(options.Description))
		}
		header.Children(heading)
	}
	if len(options.Series) > 0 {
		legend := ui.Div().Class("of-area-chart-legend")
		for index, item := range options.Series {
			dot := ui.Span().Class("of-area-chart-dot").Attr("aria-hidden", "true").
				Attr("data-series", chartSeriesSlot(index))
			if item.Color != "" {
				dot.CustomProperty("--of-area-dot", item.Color)
			}
			legend.Children(ui.Span().Class("of-area-chart-legend-item").Children(dot,
				ui.Span().Class("of-area-chart-legend-label").Text(item.Label)))
		}
		header.Children(legend)
	}
	return header
}

func areaChartBody(options AreaChartOptions) *ui.Element {
	columns := areaChartColumns(options.Series)
	if columns == 0 {
		return nil
	}
	body := ui.Div().Class("of-area-chart-body")
	plot := ui.Div().Class("of-area-chart-plot").Attr("aria-hidden", "true")
	maximum := chartMaximum(options.Series)
	for index := 0; index < columns; index++ {
		plot.Children(areaChartColumn(options.Series, index, maximum))
	}
	body.Children(plot)
	if len(options.Labels) > 0 {
		row := ui.Div().Class("of-area-chart-labels")
		for index := 0; index < columns; index++ {
			label := ""
			if index < len(options.Labels) {
				label = options.Labels[index]
			}
			row.Children(ui.Span().Class("of-area-chart-label").Text(label))
		}
		body.Children(row)
	}
	return body
}

func areaChartColumn(series []ChartSeries, index int, maximum float64) *ui.Element {
	column := ui.Div().Class("of-area-chart-column")
	low, high := areaColumnRange(series, index)
	for seriesIndex, item := range series {
		if index >= len(item.Values) {
			continue
		}
		track := ui.Div().Class("of-area-chart-track")
		dot := ui.Span().Class("of-area-chart-point").
			Attr("data-series", chartSeriesSlot(seriesIndex))
		if item.Color != "" {
			dot.CustomProperty("--of-area-dot", item.Color)
		}
		spacer := ui.Div().Class("of-area-chart-spacer").Attr("aria-hidden", "true").
			Height(ui.Percent(chartHeight(item.Values[index], maximum)))
		column.Children(track.Children(dot, spacer))
		if seriesIndex == 0 && high > low && len(series) > 1 {
			bandTrack := ui.Div().Class("of-area-chart-track of-area-chart-track--band")
			band := ui.Div().Class("of-area-chart-band").Attr("aria-hidden", "true").
				Height(ui.Percent(chartHeight(high-low, maximum)))
			lowSpacer := ui.Div().Class("of-area-chart-spacer").Attr("aria-hidden", "true").
				Height(ui.Percent(chartHeight(low, maximum)))
			column.Children(bandTrack.Children(band, lowSpacer))
		}
	}
	return column
}

func areaChartColumns(series []ChartSeries) int {
	columns := 0
	for _, item := range series {
		columns = max(columns, len(item.Values))
	}
	return columns
}

func areaColumnRange(series []ChartSeries, index int) (float64, float64) {
	low := -1.0
	high := 0.0
	for _, item := range series {
		if index >= len(item.Values) || !finite(item.Values[index]) {
			continue
		}
		value := item.Values[index]
		if low < 0 || value < low {
			low = value
		}
		if value > high {
			high = value
		}
	}
	if low < 0 {
		return 0, 0
	}
	return low, high
}

func areaChartLabel(options AreaChartOptions) string {
	label := "Area chart"
	if options.Title != "" {
		label += ": " + options.Title
	}
	if options.Description != "" {
		label += ". " + options.Description
	}
	return label + " " + strconv.Itoa(areaChartColumns(options.Series)) + " points"
}
