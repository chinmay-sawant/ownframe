package shadcnchart

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

// Pie builds a Pie chart.
func Pie(options PartOptions) *ui.Element {
	return paintParts("Pie", options, "pie", "")
}

// Radar builds a Radar chart.
func Radar(options Options) *ui.Element {
	return paintSeries("Radar", options, "radar", "")
}

// Radial builds a Radial chart.
func Radial(options PartOptions) *ui.Element {
	return paintParts("Radial", options, "radial", "")
}
