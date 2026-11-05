package apexcharts_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/apexcharts"
	"github.com/chinmay-sawant/ownframe/ui"
)

const sampleValue = 10.0

func render(t *testing.T, element *ui.Element) string {
	t.Helper()
	source, err := ui.Render(element)
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestChartsRender(t *testing.T) {
	t.Parallel()
	series := apexcharts.Options{
		Title: "Week", Labels: []string{"Mon"},
		Series: []apexcharts.Series{{Name: "Opened", Values: []float64{sampleValue}}},
	}
	parts := apexcharts.PartOptions{
		Title:  "Share",
		Slices: []apexcharts.Slice{{Name: "Desktop", Value: sampleValue}},
	}
	tree := apexcharts.TreeOptions{
		Title: "Tree",
		Nodes: []apexcharts.Node{{Name: "Root", Value: sampleValue}},
	}
	candles := apexcharts.CandleOptions{
		Title:  "Candles",
		Points: []apexcharts.OHLC{{Label: "Mon", Open: sampleValue, High: sampleValue, Low: sampleValue, Close: sampleValue}},
	}
	ranges := apexcharts.RangeOptions{
		Title:  "Range",
		Points: []apexcharts.RangePoint{{Label: "Mon", Start: sampleValue, End: sampleValue}},
	}
	cases := []struct {
		name, want string
		element    *ui.Element
	}{
		{name: "Line", want: "Week", element: apexcharts.Line(series)},
		{name: "Area", want: "Week", element: apexcharts.Area(series)},
		{name: "Bar", want: "Week", element: apexcharts.Bar(series)},
		{name: "Column", want: "Week", element: apexcharts.Column(series)},
		{name: "Scatter", want: "Week", element: apexcharts.Scatter(series)},
		{name: "Bubble", want: "Week", element: apexcharts.Bubble(series)},
		{name: "Heatmap", want: "Week", element: apexcharts.Heatmap(series)},
		{name: "Candlestick", want: "Candles", element: apexcharts.Candlestick(candles)},
		{name: "BoxPlot", want: "Week", element: apexcharts.BoxPlot(series)},
		{name: "Histogram", want: "Week", element: apexcharts.Histogram(series)},
		{name: "Pie", want: "Share", element: apexcharts.Pie(parts)},
		{name: "Donut", want: "Share", element: apexcharts.Donut(parts)},
		{name: "RadialBar", want: "Share", element: apexcharts.RadialBar(parts)},
		{name: "Gauge", want: "Load", element: apexcharts.Gauge("Load", sampleValue)},
		{name: "Radar", want: "Week", element: apexcharts.Radar(series)},
		{name: "Violin", want: "Week", element: apexcharts.Violin(series)},
		{name: "Raincloud", want: "Week", element: apexcharts.Raincloud(series)},
		{name: "RangeBar", want: "Range", element: apexcharts.RangeBar(ranges)},
		{name: "RangeArea", want: "Range", element: apexcharts.RangeArea(ranges)},
		{name: "Streamgraph", want: "Week", element: apexcharts.Streamgraph(series)},
		{name: "Waterfall", want: "Week", element: apexcharts.Waterfall(series)},
		{name: "Dumbbell", want: "Week", element: apexcharts.Dumbbell(series)},
		{name: "Treemap", want: "Tree", element: apexcharts.Treemap(tree)},
		{name: "PolarArea", want: "Share", element: apexcharts.PolarArea(parts)},
		{name: "Funnel", want: "Share", element: apexcharts.Funnel(parts)},
		{name: "Pyramid", want: "Share", element: apexcharts.Pyramid(parts)},
		{name: "Sunburst", want: "Tree", element: apexcharts.Sunburst(tree)},
		{name: "Icicle", want: "Tree", element: apexcharts.Icicle(tree)},
		{name: "Unit", want: "Share", element: apexcharts.Unit(parts)},
		{name: "Waffle", want: "Share", element: apexcharts.Waffle(parts)},
	}
	for _, tc := range cases {
		source := render(t, tc.element)
		if !strings.Contains(source, tc.want) {
			t.Fatalf("%s missing %q in %s", tc.name, tc.want, source)
		}
	}
}

func TestEmptySeriesSaysNoData(t *testing.T) {
	t.Parallel()
	source := render(t, apexcharts.Line(apexcharts.Options{}))
	if !strings.Contains(source, "No data") {
		t.Fatalf("empty chart: %s", source)
	}
}
