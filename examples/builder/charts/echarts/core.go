package echarts

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

const boxplotCount = 5

// Bar builds grouped vertical bars.
func Bar(options Options) *ui.Element {
	return columnChart("Bar", options, "of-ech-bar")
}

// Line builds one stem per value.
func Line(options Options) *ui.Element {
	return columnChart("Line", options, "of-ech-stem")
}

// Scatter builds one dot per value.
func Scatter(options Options) *ui.Element {
	return columnChart("Scatter", options, "of-ech-dot")
}

// Radar builds one column per category.
func Radar(options Options) *ui.Element {
	return columnChart("Radar", options, "of-ech-stem")
}

func columnChart(kind string, options Options, markClass string) *ui.Element {
	return extraRoot(className(kind), kind, options.Title, options.Description, extraBody(columnPlot(options, markClass), options.Series))
}

func className(kind string) string {
	return "of-ech-" + kind
}

func columnPlot(options Options, markClass string) *ui.Element {
	columns := extraColumnCount(options.Series)
	if columns == 0 {
		return nil
	}
	maximum := extraSeriesMax(options.Series)
	plot := ui.Div().Class("of-ech-plot")
	for index := range columns {
		column := ui.Div().Class("of-ech-column")
		for seriesIndex, series := range options.Series {
			if index >= len(series.Values) || !extraFinite(series.Values[index]) {
				continue
			}
			column.Children(ui.Div().Class(markClass).
				Height(ui.Percent(extraShare(series.Values[index], maximum))).
				CustomProperty("--of-ech-color", extraPaint(series.Color, seriesIndex)).
				Attr("data-series", strconv.Itoa(seriesIndex)))
		}
		column.Children(ui.Span().Class("of-ech-label").Text(extraLabelAt(options.Labels, index)))
		plot.Children(column)
	}
	return plot
}

// Pie builds a horizontal stacked track of slices.
func Pie(options PartOptions) *ui.Element {
	return extraRoot("of-ech-pie", "Pie", options.Title, options.Description, pieBody(options, ""))
}

// Funnel builds one row per slice, sized against the first positive value.
func Funnel(options PartOptions) *ui.Element {
	return extraRoot("of-ech-funnel", "Funnel", options.Title, options.Description, funnelBody(options))
}

func pieBody(options PartOptions, extra string) *ui.Element {
	if len(options.Slices) == 0 {
		return nil
	}
	total := 0.0
	for _, item := range options.Slices {
		if extraFinite(item.Value) && item.Value > 0 {
			total += item.Value
		}
	}
	if total <= 0 {
		return nil
	}
	className := "of-ech-track"
	if extra != "" {
		className += " " + extra
	}
	track := ui.Div().Class(className)
	for index, item := range options.Slices {
		track.Children(ui.Div().Class("of-ech-slice").
			Width(ui.Percent(extraShare(item.Value, total))).
			CustomProperty("--of-ech-color", extraPaint(item.Color, index)))
	}
	return ui.Div().Class("of-ech-body").Children(track)
}

func funnelBody(options PartOptions) *ui.Element {
	if len(options.Slices) == 0 {
		return nil
	}
	base := 0.0
	for _, item := range options.Slices {
		if extraFinite(item.Value) && item.Value > 0 {
			base = item.Value
			break
		}
	}
	list := ui.Div().Class("of-ech-body")
	for index, item := range options.Slices {
		list.Children(ui.Div().Class("of-ech-row").Children(
			ui.Span().Class("of-ech-label").Text(item.Name),
			ui.Div().Class("of-ech-track").Children(
				ui.Div().Class("of-ech-fill").
					Width(ui.Percent(extraShare(item.Value, base))).
					CustomProperty("--of-ech-color", extraPaint(item.Color, index)),
			),
		))
	}
	return list
}

// Heatmap builds one cell per series value.
func Heatmap(options Options) *ui.Element {
	columns := extraColumnCount(options.Series)
	if columns == 0 {
		return extraRoot("of-ech-heatmap", "Heatmap", options.Title, options.Description, nil)
	}
	grid := ui.Div().Class("of-ech-map-grid")
	for seriesIndex, series := range options.Series {
		for index := range columns {
			if index >= len(series.Values) || !extraFinite(series.Values[index]) {
				continue
			}
			grid.Children(ui.Div().Class("of-ech-heat-cell").
				CustomProperty("--of-ech-color", extraPaint(series.Color, seriesIndex)).
				Text(extraLabelAt(options.Labels, index) + " " + extraText(series.Values[index])))
		}
	}
	return extraRoot("of-ech-heatmap", "Heatmap", options.Title, options.Description, grid)
}

// Candlestick builds one column per open-high-low-close point.
func Candlestick(options CandleOptions) *ui.Element {
	if len(options.Points) == 0 {
		return extraRoot("of-ech-candle", "Candlestick", options.Title, options.Description, nil)
	}
	maximum := 0.0
	for _, point := range options.Points {
		for _, value := range []float64{point.High, point.Open, point.Close} {
			if extraFinite(value) && value > maximum {
				maximum = value
			}
		}
	}
	plot := ui.Div().Class("of-ech-candle-plot")
	for _, point := range options.Points {
		className := "of-ech-up"
		span := point.Close - point.Open
		if point.Close < point.Open {
			className = "of-ech-down"
			span = point.Open - point.Close
		}
		plot.Children(ui.Div().Class("of-ech-column").Children(
			ui.Div().Class(className).Height(ui.Percent(extraShare(span, maximum))),
			ui.Span().Class("of-ech-label").Text(point.Label),
		))
	}
	return extraRoot("of-ech-candle", "Candlestick", options.Title, options.Description, plot)
}

// Boxplot prints the five-number summary when a series has five values, and the mean otherwise.
func Boxplot(options Options) *ui.Element {
	if extraColumnCount(options.Series) == 0 {
		return extraRoot("of-ech-boxplot", "Boxplot", options.Title, options.Description, nil)
	}
	list := ui.Div().Class("of-ech-body")
	for _, series := range options.Series {
		text := extraText(meanOf(series.Values))
		if len(series.Values) == boxplotCount {
			text = joinBox(series.Values)
		}
		list.Children(ui.Div().Class("of-ech-row").Children(
			ui.Span().Class("of-ech-label").Text(series.Name),
			ui.Span().Text(text),
		))
	}
	return extraRoot("of-ech-boxplot", "Boxplot", options.Title, options.Description, list)
}

func meanOf(values []float64) float64 {
	total := 0.0
	count := 0
	for _, value := range values {
		if !extraFinite(value) {
			continue
		}
		total += value
		count++
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func joinBox(values []float64) string {
	text := ""
	for index, value := range values {
		if index > 0 {
			text += " "
		}
		text += extraText(value)
	}
	return text
}

// Gauge builds a 0 to 100 track for one value.
func Gauge(title string, value float64) *ui.Element {
	clamped := value
	if !extraFinite(clamped) || clamped < 0 {
		clamped = 0
	}
	if clamped > extraFull {
		clamped = extraFull
	}
	body := ui.Div().Class("of-ech-body").Children(
		ui.Div().Class("of-ech-track").Children(
			ui.Div().Class("of-ech-fill").Width(ui.Percent(clamped)).
				CustomProperty("--of-ech-color", extraColor(0)),
		),
		ui.Span().Class("of-ech-label").Text(extraText(clamped)),
	)
	return extraRoot("of-ech-gauge", "Gauge", title, "", body)
}
