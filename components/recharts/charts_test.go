package recharts_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/recharts"
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
	series := recharts.Options{
		Title: "Week", Labels: []string{"Mon"},
		Series: []recharts.Series{{Name: "Opened", Values: []float64{sampleValue}}},
	}
	parts := recharts.PartOptions{
		Title:  "Share",
		Slices: []recharts.Slice{{Name: "Desktop", Value: sampleValue}},
	}
	tree := recharts.TreeOptions{
		Title: "Tree",
		Nodes: []recharts.Node{{Name: "Root", Value: sampleValue}},
	}
	flow := recharts.FlowOptions{
		Title: "Flow",
		Nodes: []recharts.Node{{Name: "Source", Value: sampleValue}},
		Links: []recharts.Link{{Source: "Source", Target: "Sink", Value: sampleValue}},
	}
	cases := []struct {
		name, want string
		element    *ui.Element
	}{
		{name: "Area", want: "Week", element: recharts.Area(series)},
		{name: "Bar", want: "Week", element: recharts.Bar(series)},
		{name: "Line", want: "Week", element: recharts.Line(series)},
		{name: "Composed", want: "Week", element: recharts.Composed(series)},
		{name: "Pie", want: "Share", element: recharts.Pie(parts)},
		{name: "Radar", want: "Week", element: recharts.Radar(series)},
		{name: "RadialBar", want: "Share", element: recharts.RadialBar(parts)},
		{name: "Scatter", want: "Week", element: recharts.Scatter(series)},
		{name: "Funnel", want: "Share", element: recharts.Funnel(parts)},
		{name: "Sankey", want: "Flow", element: recharts.Sankey(flow)},
		{name: "Sunburst", want: "Tree", element: recharts.Sunburst(tree)},
		{name: "Treemap", want: "Tree", element: recharts.Treemap(tree)},
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
	source := render(t, recharts.Area(recharts.Options{}))
	if !strings.Contains(source, "No data") {
		t.Fatalf("empty chart: %s", source)
	}
}
