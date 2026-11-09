package echarts

import "github.com/chinmay-sawant/ownframe/ui"

// Custom draws one bar per value. Each bar carries of-ech-custom.
func Custom(options Options) *ui.Element {
	return extraRoot(
		"of-ech-custom",
		"Custom",
		options.Title,
		options.Description,
		extraBody(extraCustomPlot(options), options.Series),
	)
}

func extraCustomPlot(options Options) *ui.Element {
	columns := extraColumnCount(options.Series)
	if columns == 0 {
		return nil
	}
	maximum := extraSeriesMax(options.Series)
	plot := ui.Div().Class("of-ech-custom-plot")
	for index := range columns {
		plot.Children(extraCustomColumn(options, index, maximum))
	}
	return plot
}

func extraCustomColumn(options Options, index int, maximum float64) *ui.Element {
	group := ui.Div().Class("of-ech-custom-group")
	for seriesIndex, series := range options.Series {
		if index >= len(series.Values) || !extraFinite(series.Values[index]) {
			continue
		}
		group.Children(ui.Div().Class("of-ech-custom").
			Height(ui.Percent(extraShare(series.Values[index], maximum))).
			CustomProperty("--of-ech-color", extraPaint(series.Color, seriesIndex)))
	}
	return ui.Div().Class("of-ech-custom-col").Children(
		group,
		ui.Span().Class("of-ech-label").Text(extraLabelAt(options.Labels, index)),
	)
}
