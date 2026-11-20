package echarts

// Series is one named run of numbers. Values line up with Options.Labels.
// Color is an optional CSS color. An empty Color uses the chart palette.
type Series struct {
	Name   string
	Values []float64
	Color  string
}

// Options is the shared input for column and cell charts.
type Options struct {
	Title       string
	Description string
	Labels      []string
	Series      []Series
}

// Slice is one named part of a whole.
type Slice struct {
	Name  string
	Value float64
	Color string
}

// PartOptions is the input for charts built from named parts.
type PartOptions struct {
	Title       string
	Description string
	Slices      []Slice
}

// Node is one weighted tree node.
type Node struct {
	Name     string
	Value    float64
	Color    string
	Children []Node
}

// TreeOptions is the input for a weighted tree.
type TreeOptions struct {
	Title       string
	Description string
	Nodes       []Node
}

// Link is one directed weighted connection.
type Link struct {
	Source string
	Target string
	Value  float64
}

// FlowOptions is the input for a node-and-link chart.
type FlowOptions struct {
	Title       string
	Description string
	Nodes       []Node
	Links       []Link
}

// OHLC is one open-high-low-close session.
type OHLC struct {
	Label                  string
	Open, High, Low, Close float64
}

// CandleOptions is the input for a candlestick chart.
type CandleOptions struct {
	Title       string
	Description string
	Points      []OHLC
}
