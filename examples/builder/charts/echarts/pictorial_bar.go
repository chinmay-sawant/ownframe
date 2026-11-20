package echarts

import (
	"math"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// PictorialBar draws one bar per value. The bar text is the block mark
// repeated from 1 to 8 times, scaled from that value's percent of the maximum.
func PictorialBar(options Options) *ui.Element {
	return extraRoot(
		"of-ech-pictorial",
		"Pictorial bar",
		options.Title,
		options.Description,
		extraBody(extraPictorialPlot(options), options.Series),
	)
}

func extraPictorialPlot(options Options) *ui.Element {
	columns := extraColumnCount(options.Series)
	if columns == 0 {
		return nil
	}
	maximum := extraSeriesMax(options.Series)
	plot := ui.Div().Class("of-ech-pictorial-plot")
	for index := range columns {
		plot.Children(extraPictorialColumn(options, index, maximum))
	}
	return plot
}

func extraPictorialColumn(options Options, index int, maximum float64) *ui.Element {
	column := ui.Div().Class("of-ech-pictorial-col")
	for seriesIndex, series := range options.Series {
		if index >= len(series.Values) || !extraFinite(series.Values[index]) {
			continue
		}
		value := series.Values[index]
		marks := extraPictorialMarks(extraShare(value, maximum))
		bar := ui.Div().Class("of-ech-pictorial").
			CustomProperty("--of-ech-color", extraPaint(series.Color, seriesIndex)).
			Text(marks)
		column.Children(bar)
	}
	column.Children(ui.Span().Class("of-ech-label").Text(extraLabelAt(options.Labels, index)))
	return column
}

func extraPictorialMarks(percent float64) string {
	return strings.Repeat(extraMark, extraPictorialCount(percent))
}

func extraPictorialCount(percent float64) int {
	if !extraFinite(percent) || percent <= 0 {
		return extraPictorialMin
	}
	scaled := percent / extraFull * float64(extraPictorialMax)
	count := int(math.Round(scaled))
	if count < extraPictorialMin {
		return extraPictorialMin
	}
	if count > extraPictorialMax {
		return extraPictorialMax
	}
	return count
}
