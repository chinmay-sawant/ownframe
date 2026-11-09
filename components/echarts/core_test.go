package echarts_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/echarts"
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

func TestEChartsRender(t *testing.T) {
	t.Parallel()
	series := echarts.Options{
		Title: "Week", Labels: []string{"Mon"},
		Series: []echarts.Series{{Name: "Opened", Values: []float64{sampleValue}}},
	}
	parts := echarts.PartOptions{Title: "Share", Slices: []echarts.Slice{{Name: "Desktop", Value: sampleValue}}}
	tree := echarts.TreeOptions{Title: "Tree", Nodes: []echarts.Node{{Name: "Root", Value: sampleValue}}}
	flow := echarts.FlowOptions{
		Title: "Flow",
		Links: []echarts.Link{{Source: "Source", Target: "Sink", Value: sampleValue}},
	}
	candles := echarts.CandleOptions{
		Title:  "Candles",
		Points: []echarts.OHLC{{Label: "Mon", Open: sampleValue, High: sampleValue, Low: sampleValue, Close: sampleValue}},
	}
	cases := []struct {
		name, want string
		element    *ui.Element
	}{
		{name: "bar", want: "Week", element: echarts.Bar(series)},
		{name: "line", want: "Week", element: echarts.Line(series)},
		{name: "scatter", want: "Week", element: echarts.Scatter(series)},
		{name: "pie", want: "Share", element: echarts.Pie(parts)},
		{name: "radar", want: "Week", element: echarts.Radar(series)},
		{name: "heat", want: "Week", element: echarts.Heatmap(series)},
		{name: "candle", want: "Candles", element: echarts.Candlestick(candles)},
		{name: "box", want: "Week", element: echarts.Boxplot(series)},
		{name: "gauge", want: "Load", element: echarts.Gauge("Load", sampleValue)},
		{name: "funnel", want: "Share", element: echarts.Funnel(parts)},
		{name: "effect", want: "Week", element: echarts.EffectScatter(series)},
		{name: "pictorial", want: "Week", element: echarts.PictorialBar(series)},
		{name: "map", want: "Share", element: echarts.Map(parts)},
		{name: "tree", want: "Tree", element: echarts.Tree(tree)},
		{name: "treemap", want: "Tree", element: echarts.Treemap(tree)},
		{name: "graph", want: "Flow", element: echarts.Graph(flow)},
		{name: "chord", want: "Flow", element: echarts.Chord(flow)},
		{name: "parallel", want: "Week", element: echarts.Parallel(series)},
		{name: "sankey", want: "Flow", element: echarts.Sankey(flow)},
		{name: "lines", want: "Flow", element: echarts.Lines(flow)},
		{name: "river", want: "Week", element: echarts.ThemeRiver(series)},
		{name: "sun", want: "Tree", element: echarts.Sunburst(tree)},
		{name: "custom", want: "Week", element: echarts.Custom(series)},
	}
	for _, tc := range cases {
		source := render(t, tc.element)
		if !strings.Contains(source, tc.want) {
			t.Fatalf("%s missing %q in %s", tc.name, tc.want, source)
		}
	}
}
