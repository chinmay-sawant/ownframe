package echarts

import "github.com/chinmay-sawant/ownframe/ui"

// EffectScatter plots one dot per series value in a category column.
// Each dot carries of-ech-effect so the ripple stays on the point.
func EffectScatter(options Options) *ui.Element {
	return extraRoot(
		"of-ech-scatter-effect",
		"Effect scatter",
		options.Title,
		options.Description,
		extraBody(extraScatterPlot(options), options.Series),
	)
}

func extraScatterPlot(options Options) *ui.Element {
	columns := extraColumnCount(options.Series)
	if columns == 0 {
		return nil
	}
	maximum := extraSeriesMax(options.Series)
	plot := ui.Div().Class("of-ech-scatter-plot")
	for index := range columns {
		plot.Children(extraScatterColumn(options, index, maximum))
	}
	return plot
}

func extraScatterColumn(options Options, index int, maximum float64) *ui.Element {
	track := ui.Div().Class("of-ech-scatter-track")
	seriesCount := len(options.Series)
	for seriesIndex, series := range options.Series {
		if index >= len(series.Values) || !extraFinite(series.Values[index]) {
			continue
		}
		value := series.Values[index]
		dot := ui.Span().Class("of-ech-dot of-ech-effect").
			Attr("title", extraText(value)).
			CustomProperty("--of-ech-color", extraPaint(series.Color, seriesIndex)).
			Left(ui.Percent(extraScatterLeft(seriesIndex, seriesCount))).
			Bottom(ui.Percent(extraShare(value, maximum)))
		dot.Children(ui.Span().Class("of-ech-effect-ring").Attr("aria-hidden", "true"))
		track.Children(dot)
	}
	return ui.Div().Class("of-ech-scatter-col").Children(
		track,
		ui.Span().Class("of-ech-label").Text(extraLabelAt(options.Labels, index)),
	)
}

func extraScatterLeft(seriesIndex, seriesCount int) float64 {
	slots := seriesCount + 1
	if slots <= 1 {
		return 0
	}
	return extraFull * float64(seriesIndex+1) / float64(slots)
}
