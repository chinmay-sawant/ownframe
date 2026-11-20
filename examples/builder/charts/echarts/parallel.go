package echarts

import "github.com/chinmay-sawant/ownframe/ui"

// Parallel draws one row per series and one cell per value in that series.
func Parallel(options Options) *ui.Element {
	return extraRoot(
		"of-ech-parallel",
		"Parallel",
		options.Title,
		options.Description,
		extraParallelPlot(options),
	)
}

func extraParallelPlot(options Options) *ui.Element {
	if len(options.Series) == 0 {
		return nil
	}
	columns := extraColumnCount(options.Series)
	maximum := extraSeriesMax(options.Series)
	plot := ui.Div().Class("of-ech-parallel-plot")
	if len(options.Labels) > 0 && columns > 0 {
		plot.Children(extraParallelLabels(options.Labels, columns))
	}
	for seriesIndex, series := range options.Series {
		plot.Children(extraParallelRow(series, seriesIndex, columns, maximum))
	}
	return plot
}

func extraParallelLabels(labels []string, columns int) *ui.Element {
	row := ui.Div().Class("of-ech-parallel-labels")
	row.Children(ui.Span().Class("of-ech-parallel-name").Text(""))
	for index := range columns {
		row.Children(ui.Span().Class("of-ech-parallel-axis").Text(extraLabelAt(labels, index)))
	}
	return row
}

func extraParallelRow(series Series, seriesIndex, columns int, maximum float64) *ui.Element {
	row := ui.Div().Class("of-ech-parallel-row")
	row.Children(ui.Span().Class("of-ech-parallel-name").Text(series.Name))
	for index := range columns {
		cell := ui.Div().Class("of-ech-parallel-cell")
		if index < len(series.Values) && extraFinite(series.Values[index]) {
			value := series.Values[index]
			cell.Children(
				ui.Span().Class("of-ech-parallel-bar").
					Width(ui.Percent(extraShare(value, maximum))).
					CustomProperty("--of-ech-color", extraPaint(series.Color, seriesIndex)),
				ui.Span().Class("of-ech-parallel-value").Text(extraText(value)),
			)
		}
		row.Children(cell)
	}
	return row
}
