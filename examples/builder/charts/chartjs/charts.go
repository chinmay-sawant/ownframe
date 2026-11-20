package chartjs

import "github.com/chinmay-sawant/ownframe/ui"

// Bar builds a Bar chart.
func Bar(options Options) *ui.Element {
	return paintSeries("Bar", options, "bar", "")
}

// Line builds a Line chart.
func Line(options Options) *ui.Element {
	return paintSeries("Line", options, "line", "")
}

// Pie builds a Pie chart.
func Pie(options PartOptions) *ui.Element {
	return paintParts("Pie", options, "pie", "")
}

// Doughnut builds a Doughnut chart.
func Doughnut(options PartOptions) *ui.Element {
	return paintParts("Doughnut", options, "donut", "")
}

// Radar builds a Radar chart.
func Radar(options Options) *ui.Element {
	return paintSeries("Radar", options, "radar", "")
}

// PolarArea builds a Polar area chart.
func PolarArea(options PartOptions) *ui.Element {
	return paintParts("Polar area", options, "polar", "")
}

// Bubble builds a Bubble chart.
func Bubble(options Options) *ui.Element {
	return paintSeries("Bubble", options, "bubble", "")
}

// Scatter builds a Scatter chart.
func Scatter(options Options) *ui.Element {
	return paintSeries("Scatter", options, "scatter", "")
}
