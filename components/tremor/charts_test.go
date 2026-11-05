package tremor_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/tremor"
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
	series := tremor.Options{
		Title: "Week", Labels: []string{"Mon"},
		Series: []tremor.Series{{Name: "Opened", Values: []float64{sampleValue}}},
	}
	parts := tremor.PartOptions{
		Title:  "Share",
		Slices: []tremor.Slice{{Name: "Desktop", Value: sampleValue}},
	}
	tracker := tremor.TrackerOptions{
		Title: "Tracker",
		Cells: []tremor.TrackerCell{{Name: "Mon", Status: "good"}},
	}
	cases := []struct {
		name, want string
		element    *ui.Element
	}{
		{name: "Area", want: "Week", element: tremor.Area(series)},
		{name: "Bar", want: "Week", element: tremor.Bar(series)},
		{name: "Line", want: "Week", element: tremor.Line(series)},
		{name: "Combo", want: "Week", element: tremor.Combo(series)},
		{name: "Donut", want: "Share", element: tremor.Donut(parts)},
		{name: "SparkArea", want: "Week", element: tremor.SparkArea(series)},
		{name: "SparkLine", want: "Week", element: tremor.SparkLine(series)},
		{name: "SparkBar", want: "Week", element: tremor.SparkBar(series)},
		{name: "BarList", want: "Share", element: tremor.BarList(parts)},
		{name: "CategoryBar", want: "Share", element: tremor.CategoryBar(parts)},
		{name: "Tracker", want: "Tracker", element: tremor.Tracker(tracker)},
		{name: "ProgressBar", want: "Load", element: tremor.ProgressBar("Load", sampleValue)},
		{name: "ProgressCircle", want: "Load", element: tremor.ProgressCircle("Load", sampleValue)},
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
	source := render(t, tremor.Area(tremor.Options{}))
	if !strings.Contains(source, "No data") {
		t.Fatalf("empty chart: %s", source)
	}
}
