package apexcharts

import "github.com/chinmay-sawant/ownframe/ui"

// Line builds a Line chart.
func Line(options Options) *ui.Element {
	return paintSeries("Line", options, "line", "")
}

// Area builds a Area chart.
func Area(options Options) *ui.Element {
	return paintSeries("Area", options, "area", "")
}

// Bar builds a Bar chart.
func Bar(options Options) *ui.Element {
	return paintSeries("Bar", options, "hbar", "")
}

// Column builds a Column chart.
func Column(options Options) *ui.Element {
	return paintSeries("Column", options, "bar", "")
}

// Scatter builds a Scatter chart.
func Scatter(options Options) *ui.Element {
	return paintSeries("Scatter", options, "scatter", "")
}

// Bubble builds a Bubble chart.
func Bubble(options Options) *ui.Element {
	return paintSeries("Bubble", options, "bubble", "")
}

// Heatmap builds a Heatmap chart.
func Heatmap(options Options) *ui.Element {
	return paintSeries("Heatmap", options, "heat", "")
}

// Candlestick builds a Candlestick chart.
func Candlestick(options CandleOptions) *ui.Element {
	return paintCandles("Candlestick", options)
}

// BoxPlot builds a Box plot chart.
func BoxPlot(options Options) *ui.Element {
	return paintSeries("Box plot", options, "box", "")
}

// Histogram builds a Histogram chart.
func Histogram(options Options) *ui.Element {
	return paintSeries("Histogram", options, "bar", "")
}

// Pie builds a Pie chart.
func Pie(options PartOptions) *ui.Element {
	return paintParts("Pie", options, "pie", "")
}

// Donut builds a Donut chart.
func Donut(options PartOptions) *ui.Element {
	return paintParts("Donut", options, "donut", "")
}

// RadialBar builds a Radial bar chart.
func RadialBar(options PartOptions) *ui.Element {
	return paintParts("Radial bar", options, "radial", "")
}

// Gauge builds a Gauge chart.
func Gauge(title string, value float64) *ui.Element {
	return paintGauge("Gauge", title, value)
}

// Radar builds a Radar chart.
func Radar(options Options) *ui.Element {
	return paintSeries("Radar", options, "radar", "")
}

// Violin builds a Violin chart.
func Violin(options Options) *ui.Element {
	return paintSeries("Violin", options, "violin", "")
}

// Raincloud builds a Raincloud chart.
func Raincloud(options Options) *ui.Element {
	return paintSeries("Raincloud", options, "raincloud", "")
}

// RangeBar builds a Range bar chart.
func RangeBar(options RangeOptions) *ui.Element {
	return paintRange("Range bar", options, "")
}

// RangeArea builds a Range area chart.
func RangeArea(options RangeOptions) *ui.Element {
	return paintRange("Range area", options, "")
}

// Streamgraph builds a Streamgraph chart.
func Streamgraph(options Options) *ui.Element {
	return paintSeries("Streamgraph", options, "stream", "")
}

// Waterfall builds a Waterfall chart.
func Waterfall(options Options) *ui.Element {
	return paintSeries("Waterfall", options, "waterfall", "")
}

// Dumbbell builds a Dumbbell chart.
func Dumbbell(options Options) *ui.Element {
	return paintSeries("Dumbbell", options, "dumbbell", "")
}

// Treemap builds a Treemap chart.
func Treemap(options TreeOptions) *ui.Element {
	return paintTree("Treemap", options, "")
}

// PolarArea builds a Polar area chart.
func PolarArea(options PartOptions) *ui.Element {
	return paintParts("Polar area", options, "polar", "")
}

// Funnel builds a Funnel chart.
func Funnel(options PartOptions) *ui.Element {
	return paintParts("Funnel", options, "funnel", "")
}

// Pyramid builds a Pyramid chart.
func Pyramid(options PartOptions) *ui.Element {
	return paintParts("Pyramid", options, "pyramid", "")
}

// Sunburst builds a Sunburst chart.
func Sunburst(options TreeOptions) *ui.Element {
	return paintTree("Sunburst", options, "")
}

// Icicle builds a Icicle chart.
func Icicle(options TreeOptions) *ui.Element {
	return paintTree("Icicle", options, "")
}

// Unit builds a Unit chart.
func Unit(options PartOptions) *ui.Element {
	return paintParts("Unit", options, "unit", "")
}

// Waffle builds a Waffle chart.
func Waffle(options PartOptions) *ui.Element {
	return paintParts("Waffle", options, "waffle", "")
}
