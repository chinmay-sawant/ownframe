package chartjs_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chartjs"
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
	series := chartjs.Options{
		Title: "Week", Labels: []string{"Mon"},
		Series: []chartjs.Series{{Name: "Opened", Values: []float64{sampleValue}}},
	}
	parts := chartjs.PartOptions{
		Title:  "Share",
		Slices: []chartjs.Slice{{Name: "Desktop", Value: sampleValue}},
	}
	cases := []struct {
		name, want string
		element    *ui.Element
	}{
		{name: "Bar", want: "Week", element: chartjs.Bar(series)},
		{name: "Line", want: "Week", element: chartjs.Line(series)},
		{name: "Pie", want: "Share", element: chartjs.Pie(parts)},
		{name: "Doughnut", want: "Share", element: chartjs.Doughnut(parts)},
		{name: "Radar", want: "Week", element: chartjs.Radar(series)},
		{name: "PolarArea", want: "Share", element: chartjs.PolarArea(parts)},
		{name: "Bubble", want: "Week", element: chartjs.Bubble(series)},
		{name: "Scatter", want: "Week", element: chartjs.Scatter(series)},
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
	source := render(t, chartjs.Bar(chartjs.Options{}))
	if !strings.Contains(source, "No data") {
		t.Fatalf("empty chart: %s", source)
	}
}
