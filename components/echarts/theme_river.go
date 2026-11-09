package echarts

import "github.com/chinmay-sawant/ownframe/ui"

// ThemeRiver draws one horizontal band per label. Each series is a segment
// whose width is that value's percent of the row total.
func ThemeRiver(options Options) *ui.Element {
	return extraRoot(
		"of-ech-theme-river",
		"Theme river",
		options.Title,
		options.Description,
		extraBody(extraRiverPlot(options), options.Series),
	)
}

func extraRiverPlot(options Options) *ui.Element {
	rows := extraColumnCount(options.Series)
	if rows == 0 {
		return nil
	}
	plot := ui.Div().Class("of-ech-river")
	for index := range rows {
		plot.Children(extraRiverRow(options, index))
	}
	return plot
}

func extraRiverRow(options Options, index int) *ui.Element {
	total := extraRiverTotal(options.Series, index)
	track := ui.Div().Class("of-ech-river-track")
	for seriesIndex, series := range options.Series {
		if index >= len(series.Values) || !extraFinite(series.Values[index]) || series.Values[index] <= 0 {
			continue
		}
		seg := ui.Div().Class("of-ech-river-seg").
			Width(ui.Percent(extraShare(series.Values[index], total))).
			CustomProperty("--of-ech-color", extraPaint(series.Color, seriesIndex))
		if series.Name != "" {
			seg.Attr("title", series.Name)
		}
		track.Children(seg)
	}
	return ui.Div().Class("of-ech-river-row").Children(
		ui.Span().Class("of-ech-river-label").Text(extraLabelAt(options.Labels, index)),
		track,
	)
}

func extraRiverTotal(series []Series, index int) float64 {
	total := 0.0
	for _, item := range series {
		if index >= len(item.Values) || !extraFinite(item.Values[index]) || item.Values[index] <= 0 {
			continue
		}
		total += item.Values[index]
	}
	return total
}
