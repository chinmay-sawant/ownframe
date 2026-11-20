package tremor

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

// Combo builds a Combo chart.
func Combo(options Options) *ui.Element {
	return paintSeries("Combo", options, "combo", "")
}

// Donut builds a Donut chart.
func Donut(options PartOptions) *ui.Element {
	return paintParts("Donut", options, "donut", "")
}

// SparkArea builds a Spark area chart.
func SparkArea(options Options) *ui.Element {
	return paintSeries("Spark area", options, "area", "of-trm-spark")
}

// SparkLine builds a Spark line chart.
func SparkLine(options Options) *ui.Element {
	return paintSeries("Spark line", options, "line", "of-trm-spark")
}

// SparkBar builds a Spark bar chart.
func SparkBar(options Options) *ui.Element {
	return paintSeries("Spark bar", options, "bar", "of-trm-spark")
}

// BarList builds a Bar list chart.
func BarList(options PartOptions) *ui.Element {
	return paintParts("Bar list", options, "barlist", "")
}

// CategoryBar builds a Category bar chart.
func CategoryBar(options PartOptions) *ui.Element {
	return paintParts("Category bar", options, "category", "")
}

// Tracker builds a Tracker chart.
func Tracker(options TrackerOptions) *ui.Element {
	return paintTracker("Tracker", options)
}

// ProgressBar builds a Progress bar chart.
func ProgressBar(title string, value float64) *ui.Element {
	chart := paintGauge("Progress bar", title, value)
	return chart.Class(classRoot + " " + "")
}

// ProgressCircle builds a Progress circle chart.
func ProgressCircle(title string, value float64) *ui.Element {
	chart := paintGauge("Progress circle", title, value)
	return chart.Class(classRoot + " " + "of-trm-progress-circle")
}
