package nivo_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/nivo"
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

func TestNivoChartsRender(t *testing.T) {
	t.Parallel()
	series := nivo.Options{
		Title: "Week", Labels: []string{"Mon"},
		Series: []nivo.Series{{Name: "Opened", Values: []float64{sampleValue}}},
	}
	parts := nivo.PartOptions{
		Title:  "Share",
		Slices: []nivo.Slice{{Name: "Desktop", Value: sampleValue}},
	}
	tree := nivo.TreeOptions{Title: "Tree", Nodes: []nivo.Node{{Name: "Root", Value: sampleValue}}}
	flow := nivo.FlowOptions{
		Title: "Flow",
		Nodes: []nivo.Node{{Name: "Source", Value: sampleValue}},
		Links: []nivo.Link{{Source: "Source", Target: "Sink", Value: sampleValue}},
	}
	days := nivo.CalendarOptions{Title: "Days", Days: []nivo.Day{{Label: "Mon", Value: sampleValue}}}
	cases := []struct {
		name, want string
		element    *ui.Element
	}{
		{name: "bar", want: "Week", element: nivo.Bar(series)},
		{name: "line", want: "Week", element: nivo.Line(series)},
		{name: "pie", want: "Share", element: nivo.Pie(parts)},
		{name: "radar", want: "Week", element: nivo.Radar(series)},
		{name: "scatter", want: "Week", element: nivo.ScatterPlot(series)},
		{name: "heat", want: "Week", element: nivo.HeatMap(series)},
		{name: "funnel", want: "Share", element: nivo.Funnel(parts)},
		{name: "calendar", want: "Days", element: nivo.Calendar(days)},
		{name: "bullet", want: "Week", element: nivo.Bullet(series)},
		{name: "bump", want: "Week", element: nivo.Bump(series)},
		{name: "area-bump", want: "Week", element: nivo.AreaBump(series)},
		{name: "box", want: "Week", element: nivo.BoxPlot(series)},
		{name: "choropleth", want: "Share", element: nivo.Choropleth(parts)},
		{name: "pack", want: "Tree", element: nivo.CirclePacking(tree)},
		{name: "chord", want: "Flow", element: nivo.Chord(flow)},
		{name: "geomap", want: "Share", element: nivo.GeoMap(parts)},
		{name: "icicle", want: "Tree", element: nivo.Icicle(tree)},
		{name: "marimekko", want: "Week", element: nivo.Marimekko(series)},
		{name: "network", want: "Flow", element: nivo.Network(flow)},
		{name: "parallel", want: "Week", element: nivo.ParallelCoordinates(series)},
		{name: "polar", want: "Week", element: nivo.PolarBar(series)},
		{name: "radial", want: "Share", element: nivo.RadialBar(parts)},
		{name: "sankey", want: "Flow", element: nivo.Sankey(flow)},
		{name: "stream", want: "Week", element: nivo.Stream(series)},
		{name: "sunburst", want: "Tree", element: nivo.Sunburst(tree)},
		{name: "swarm", want: "Week", element: nivo.SwarmPlot(series)},
		{name: "time", want: "Days", element: nivo.TimeRange(days)},
		{name: "tree", want: "Tree", element: nivo.Tree(tree)},
		{name: "treemap", want: "Tree", element: nivo.TreeMap(tree)},
		{name: "voronoi", want: "Week", element: nivo.Voronoi(series)},
		{name: "waffle", want: "Share", element: nivo.Waffle(parts)},
	}
	for _, tc := range cases {
		source := render(t, tc.element)
		if !strings.Contains(source, tc.want) {
			t.Fatalf("%s missing %q in %s", tc.name, tc.want, source)
		}
	}
}
