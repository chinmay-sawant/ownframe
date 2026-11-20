package nivo

import (
	"cmp"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	paletteSize     = 5
	percentScale    = 100
	numberTolerance = 1e-9
)

// Series is one named run of numbers. Values line up with Options.Labels by
// index. Color is an optional CSS color; an empty Color uses Palette.
type Series struct {
	Name   string
	Values []float64
	Color  string
}

// Options is the shared input for column, rank, and cell charts. Every series
// shares Labels.
type Options struct {
	Title       string
	Description string
	Labels      []string
	Series      []Series
}

// Slice is one named part of a whole. Color is an optional CSS color.
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

// Node is one weighted tree node. Children are drawn after the parent.
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

// Day is one calendar cell. Label is the visible text. Value drives the title
// tooltip, not the cell text.
type Day struct {
	Label string
	Value float64
}

// CalendarOptions is the input for a day grid.
type CalendarOptions struct {
	Title       string
	Description string
	Days        []Day
}

// Palette returns one of five chart colors. The index wraps, including
// negative indexes, in this order: #2563eb, #16a34a, #d97706, #dc2626, #7c3aed.
func Palette(index int) string {
	colors := [paletteSize]string{
		"#2563eb",
		"#16a34a",
		"#d97706",
		"#dc2626",
		"#7c3aed",
	}
	index %= paletteSize
	if index < 0 {
		index += paletteSize
	}
	return colors[index]
}

func newFrame(rootClass, kind, title, description string, body *ui.Element) *ui.Element {
	root := ui.Div().Class("of-nivo "+rootClass).
		Attr("role", "group").
		Attr("aria-label", frameLabel(kind, title, description))
	if title != "" || description != "" {
		root.Children(frameHeading(title, description))
	}
	if body != nil {
		root.Children(body)
	}
	return root
}

func frameLabel(kind, title, description string) string {
	label := kind
	if title != "" {
		label += ": " + title
	}
	if description != "" {
		label += ". " + description
	}
	return label
}

func frameHeading(title, description string) *ui.Element {
	heading := ui.Div().Class("of-nivo-heading")
	if title != "" {
		heading.Children(ui.Tag("h3").Class("of-nivo-title").Text(title))
	}
	if description != "" {
		heading.Children(ui.P().Class("of-nivo-description").Text(description))
	}
	return heading
}

func emptyNote() *ui.Element {
	return ui.Div().Class("of-nivo-body").Children(
		ui.P().Class("of-nivo-empty").Text("No data"),
	)
}

func paint(element *ui.Element, color string, index int) *ui.Element {
	if color == "" {
		color = Palette(index)
	}
	return element.CustomProperty("--of-nivo-color", color)
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func formatNumber(value float64) string {
	if !finite(value) {
		return "0"
	}
	rounded := math.Round(value)
	if math.Abs(value-rounded) < numberTolerance {
		return strconv.FormatFloat(rounded, 'f', 0, 64)
	}
	return strconv.FormatFloat(value, 'f', 1, 64)
}

func markTitle(name string, value float64) string {
	text := formatNumber(value)
	if name == "" {
		return text
	}
	return name + " " + text
}

// percentOf maps value onto 0..100 of maximum. Non-positive values and a
// non-positive maximum are 0. A value above the maximum stays at 100 so the
// mark cannot leave its track.
func percentOf(value, maximum float64) float64 {
	if maximum <= 0 || !finite(value) || value <= 0 {
		return 0
	}
	ratio := value / maximum * percentScale
	if ratio > percentScale {
		return percentScale
	}
	return ratio
}

func seriesHasPoint(series Series) bool {
	for _, value := range series.Values {
		if finite(value) {
			return true
		}
	}
	return false
}

func hasPoints(options Options) bool {
	for _, series := range options.Series {
		if seriesHasPoint(series) {
			return true
		}
	}
	return false
}

func seriesMaximum(series []Series) float64 {
	maximum := 0.0
	for _, item := range series {
		for _, value := range item.Values {
			if finite(value) && value > maximum {
				maximum = value
			}
		}
	}
	return maximum
}

func columnCount(options Options) int {
	count := len(options.Labels)
	for _, series := range options.Series {
		if len(series.Values) > count {
			count = len(series.Values)
		}
	}
	return count
}

func valueAt(values []float64, index int) (float64, bool) {
	if index < 0 || index >= len(values) || !finite(values[index]) {
		return 0, false
	}
	return values[index], true
}

func labelAt(labels []string, index int) string {
	if index < 0 || index >= len(labels) {
		return ""
	}
	return labels[index]
}

func labelRow(labels []string, columns int) *ui.Element {
	if columns == 0 || len(labels) == 0 {
		return nil
	}
	row := ui.Div().Class("of-nivo-labels")
	for index := 0; index < columns; index++ {
		row.Children(ui.Span().Class("of-nivo-label").Text(labelAt(labels, index)))
	}
	return row
}

func legend(series []Series) *ui.Element {
	if len(series) == 0 {
		return nil
	}
	row := ui.Div().Class("of-nivo-legend")
	for index, item := range series {
		swatch := paint(ui.Span().Class("of-nivo-swatch"), item.Color, index)
		row.Children(ui.Span().Class("of-nivo-legend-item").Children(
			swatch,
			ui.Span().Class("of-nivo-legend-label").Text(item.Name),
		))
	}
	return row
}

func columnBody(options Options, plot *ui.Element) *ui.Element {
	return ui.Div().Class("of-nivo-body").Children(
		plot,
		labelRow(options.Labels, columnCount(options)),
		legend(options.Series),
	)
}

func heightColumns(options Options, markClass string) *ui.Element {
	maximum := seriesMaximum(options.Series)
	columns := columnCount(options)
	plot := ui.Div().Class("of-nivo-plot")
	for index := 0; index < columns; index++ {
		column := ui.Div().Class("of-nivo-column")
		for seriesIndex, series := range options.Series {
			value, ok := valueAt(series.Values, index)
			if !ok {
				continue
			}
			mark := paint(ui.Div().Class(markClass), series.Color, seriesIndex).
				Height(ui.Percent(percentOf(value, maximum))).
				Attr("title", markTitle(series.Name, value))
			column.Children(mark)
		}
		plot.Children(column)
	}
	return columnBody(options, plot)
}

func sliceTotal(parts []Slice) float64 {
	total := 0.0
	for _, part := range parts {
		if finite(part.Value) && part.Value > 0 {
			total += part.Value
		}
	}
	return total
}

func sliceMaximum(parts []Slice) float64 {
	maximum := 0.0
	for _, part := range parts {
		if finite(part.Value) && part.Value > maximum {
			maximum = part.Value
		}
	}
	return maximum
}

func firstPositive(parts []Slice) float64 {
	for _, part := range parts {
		if finite(part.Value) && part.Value > 0 {
			return part.Value
		}
	}
	return 0
}

func nodeMaximum(nodes []Node) float64 {
	maximum := 0.0
	var walk func([]Node)
	walk = func(level []Node) {
		for _, node := range level {
			if finite(node.Value) && node.Value > maximum {
				maximum = node.Value
			}
			walk(node.Children)
		}
	}
	walk(nodes)
	return maximum
}

func meanOf(values []float64) float64 {
	total := 0.0
	count := 0
	for _, value := range values {
		if !finite(value) {
			continue
		}
		total += value
		count++
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func partFrame(rootClass, kind string, options PartOptions, body *ui.Element) *ui.Element {
	return newFrame(rootClass, kind, options.Title, options.Description, body)
}

func rankedOrder(series []Series, index int) []int {
	order := make([]int, 0, len(series))
	for seriesIndex, item := range series {
		if _, ok := valueAt(item.Values, index); ok {
			order = append(order, seriesIndex)
		}
	}
	slices.SortStableFunc(order, func(left, right int) int {
		leftValue, _ := valueAt(series[left].Values, index)
		rightValue, _ := valueAt(series[right].Values, index)
		return cmp.Compare(rightValue, leftValue)
	})
	return order
}

func ranksAt(series []Series, index int) []int {
	ranks := make([]int, len(series))
	order := rankedOrder(series, index)
	rank := 0
	var previous float64
	for position, seriesIndex := range order {
		value, _ := valueAt(series[seriesIndex].Values, index)
		if position == 0 || value != previous {
			rank++
			previous = value
		}
		ranks[seriesIndex] = rank
	}
	return ranks
}

func joinNumbers(values []float64) string {
	parts := make([]string, len(values))
	for index, value := range values {
		parts[index] = formatNumber(value)
	}
	return strings.Join(parts, " ")
}
