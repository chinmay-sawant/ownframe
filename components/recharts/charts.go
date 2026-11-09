package recharts

import "github.com/chinmay-sawant/ownframe/ui"

// Area builds a Area chart.
func Area(options Options) *ui.Element {
	return paintSeries("Area", options, "area", "")
}

// Bar builds a Bar chart.
func Bar(options Options) *ui.Element {
	return paintSeries("Bar", options, "bar", "")
}

// Line builds a Line chart.
func Line(options Options) *ui.Element {
	return paintSeries("Line", options, "line", "")
}

// Composed builds a Composed chart.
func Composed(options Options) *ui.Element {
	return paintSeries("Composed", options, "combo", "")
}

// Pie builds a Pie chart.
func Pie(options PartOptions) *ui.Element {
	return paintParts("Pie", options, "pie", "")
}

// Radar builds a Radar chart.
func Radar(options Options) *ui.Element {
	return paintSeries("Radar", options, "radar", "")
}

// RadialBar builds a Radial bar chart.
func RadialBar(options PartOptions) *ui.Element {
	return paintParts("Radial bar", options, "radial", "")
}

// Scatter builds a Scatter chart.
func Scatter(options Options) *ui.Element {
	return paintSeries("Scatter", options, "scatter", "")
}

// Funnel builds a Funnel chart.
func Funnel(options PartOptions) *ui.Element {
	return paintParts("Funnel", options, "funnel", "")
}

// Sankey builds a Sankey chart.
func Sankey(options FlowOptions) *ui.Element {
	return paintFlow("Sankey", options, "")
}

// Sunburst builds a Sunburst chart.
func Sunburst(options TreeOptions) *ui.Element {
	return paintTree("Sunburst", options, "")
}

// Treemap builds a Treemap chart.
func Treemap(options TreeOptions) *ui.Element {
	return paintTree("Treemap", options, "")
}
