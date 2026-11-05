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
	scaleLow, scaleHigh := areaPlotScale(options.Series)
	for index := 0; index < columns; index++ {
		plot.Children(areaChartColumn(options.Series, index, scaleLow, scaleHigh))
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

func areaChartColumn(series []ChartSeries, index int, scaleLow, scaleHigh float64) *ui.Element {
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
			Height(ui.Percent(areaOffset(item.Values[index], scaleLow, scaleHigh)))
		column.Children(track.Children(dot, spacer))
		if seriesIndex == 0 && high > low && len(series) > 1 {
			bandTrack := ui.Div().Class("of-area-chart-track of-area-chart-track--band")
			bandHeight := areaOffset(high, scaleLow, scaleHigh) - areaOffset(low, scaleLow, scaleHigh)
			band := ui.Div().Class("of-area-chart-band").Attr("aria-hidden", "true").
				Height(ui.Percent(bandHeight))
			lowSpacer := ui.Div().Class("of-area-chart-spacer").Attr("aria-hidden", "true").
				Height(ui.Percent(areaOffset(low, scaleLow, scaleHigh)))
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

// areaColumnRange returns the lowest and highest finite sample in one column.
// Negatives count. An empty column returns 0, 0 so no band is drawn.
func areaColumnRange(series []ChartSeries, index int) (float64, float64) {
	var low, high float64
	seen := false
	for _, item := range series {
		if index >= len(item.Values) || !finite(item.Values[index]) {
			continue
		}
		value := item.Values[index]
		if !seen || value < low {
			low = value
		}
		if !seen || value > high {
			high = value
		}
		seen = true
	}
	if !seen {
		return 0, 0
	}
	return low, high
}

// areaPlotScale is the shared vertical range. Positive series keep a zero
// baseline. A negative sample extends the floor so the band can sit between
// the highest and lowest value instead of collapsing.
func areaPlotScale(series []ChartSeries) (float64, float64) {
	var low, high float64
	seen := false
	for _, item := range series {
		for _, value := range item.Values {
			if !finite(value) {
				continue
			}
			if !seen || value < low {
				low = value
			}
			if !seen || value > high {
				high = value
			}
			seen = true
		}
	}
	if !seen {
		return 0, 0
	}
	if low > 0 {
		low = 0
	}
	return low, high
}

// areaOffset maps value into the plot as a percentage of the scale.
func areaOffset(value, low, high float64) float64 {
	span := high - low
	if span <= 0 || !finite(value) {
		return 0
	}
	portion := (value - low) / span * chartPercent
	if portion < 0 {
		return 0
	}
	if portion > chartPercent {
		return chartPercent
	}
	return portion
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
