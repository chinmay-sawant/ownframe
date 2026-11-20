package tremor

import (
	"math"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	classRoot    = "of-trm"
	percentFull  = 100
	paletteSize  = 5
	boxCount     = 5
	unitCap      = 40
	waffleCells  = 10
	waffleTotal  = 100
	pictorialMax = 8
	pictorialDiv = 12.5
	floatDigits  = 0
	floatBits    = 64
)

// Series is one named run of numbers. Values line up with Options.Labels.
// Color is an optional CSS color. An empty Color uses Palette.
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

// RangePoint is one start-to-end span.
type RangePoint struct {
	Label      string
	Start, End float64
}

// RangeOptions is the input for a range bar or range area.
type RangeOptions struct {
	Title       string
	Description string
	Points      []RangePoint
}

// TrackerCell is one status mark. Status is good, warn, bad, or empty.
type TrackerCell struct {
	Name   string
	Status string
}

// TrackerOptions is the input for a status tracker.
type TrackerOptions struct {
	Title       string
	Description string
	Cells       []TrackerCell
}

// Day is one calendar cell.
type Day struct {
	Label string
	Value float64
}

// CalendarOptions is the input for a day row.
type CalendarOptions struct {
	Title       string
	Description string
	Days        []Day
}

// Palette returns one of five chart colors, wrapping the index.
func Palette(index int) string {
	colors := [paletteSize]string{"#2563eb", "#16a34a", "#d97706", "#dc2626", "#7c3aed"}
	if index < 0 {
		index = 0
	}
	return colors[index%paletteSize]
}

func colorOf(color string, index int) string {
	if color != "" {
		return color
	}
	return Palette(index)
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func percentOf(value, total float64) float64 {
	if !finite(value) || !finite(total) || total <= 0 || value <= 0 {
		return 0
	}
	share := value / total * percentFull
	if share > percentFull {
		return percentFull
	}
	return share
}

func clampPercent(value float64) float64 {
	if !finite(value) || value < 0 {
		return 0
	}
	if value > percentFull {
		return percentFull
	}
	return value
}

func formatNum(value float64) string {
	if !finite(value) {
		return "0"
	}
	return strconv.FormatFloat(value, 'f', floatDigits, floatBits)
}

func newRoot(kind, title, description, extra string) *ui.Element {
	label := kind
	if title != "" {
		label += ": " + title
	}
	className := classRoot
	if extra != "" {
		className += " " + extra
	}
	root := ui.Section().Class(className).Attr("role", "group").Attr("aria-label", label)
	if title == "" && description == "" {
		return root
	}
	header := ui.Header().Class(classRoot + "-header")
	if title != "" {
		header.Children(ui.Tag("h2").Class(classRoot + "-title").Text(title))
	}
	if description != "" {
		header.Children(ui.P().Class(classRoot + "-description").Text(description))
	}
	return root.Children(header)
}

func withBody(root, body *ui.Element) *ui.Element {
	if body == nil {
		return root.Children(ui.P().Class(classRoot + "-empty").Text("No data"))
	}
	return root.Children(body)
}

func columnCount(series []Series) int {
	count := 0
	for _, item := range series {
		if len(item.Values) > count {
			count = len(item.Values)
		}
	}
	return count
}

func seriesMax(series []Series) float64 {
	maximum := 0.0
	for _, item := range series {
		for _, value := range item.Values {
			abs := value
			if abs < 0 {
				abs = -abs
			}
			if finite(abs) && abs > maximum {
				maximum = abs
			}
		}
	}
	return maximum
}

func labelAt(labels []string, index int) string {
	if index < 0 || index >= len(labels) {
		return ""
	}
	return labels[index]
}

func legend(series []Series) *ui.Element {
	if len(series) == 0 {
		return nil
	}
	row := ui.Div().Class(classRoot + "-legend")
	for index, item := range series {
		if item.Name == "" {
			continue
		}
		swatch := ui.Span().Class(classRoot+"-swatch").Attr("aria-hidden", "true").
			CustomProperty("--"+classRoot+"-color", colorOf(item.Color, index))
		row.Children(ui.Span().Class(classRoot+"-legend-item").Children(
			swatch,
			ui.Span().Class(classRoot+"-legend-label").Text(item.Name),
		))
	}
	return row
}

func labelRow(labels []string, count int) *ui.Element {
	if count == 0 || len(labels) == 0 {
		return nil
	}
	row := ui.Div().Class(classRoot + "-labels")
	for index := range count {
		row.Children(ui.Span().Class(classRoot + "-label").Text(labelAt(labels, index)))
	}
	return row
}

func paintSeries(kind string, options Options, mode, extra string) *ui.Element {
	root := newRoot(kind, options.Title, options.Description, extra)
	return withBody(root, seriesBody(options, mode))
}

func seriesBody(options Options, mode string) *ui.Element {
	count := columnCount(options.Series)
	if count == 0 {
		return nil
	}
	switch mode {
	case "heat", "voronoi", "parallel":
		return gridBody(options, mode)
	case "stream", "marimekko":
		return bandBody(options, mode)
	case "bump":
		return bumpBody(options)
	case "box", "violin", "raincloud":
		return summaryBody(options, mode)
	case "waterfall":
		return waterfallBody(options)
	case "dumbbell":
		return dumbbellBody(options)
	case "bullet", "hbar":
		return horizontalBody(options, mode)
	default:
		return columnBody(options, mode)
	}
}

func columnBody(options Options, mode string) *ui.Element {
	count := columnCount(options.Series)
	maximum := seriesMax(options.Series)
	plot := ui.Div().Class(classRoot + "-plot")
	for index := range count {
		plot.Children(columnMark(options, index, maximum, mode))
	}
	body := ui.Div().Class(classRoot + "-body").Children(plot)
	if labels := labelRow(options.Labels, count); labels != nil {
		body.Children(labels)
	}
	if row := legend(options.Series); row != nil {
		body.Children(row)
	}
	return body
}

func columnMark(options Options, index int, maximum float64, mode string) *ui.Element {
	column := ui.Div().Class(classRoot + "-column")
	for seriesIndex, item := range options.Series {
		if index >= len(item.Values) || !finite(item.Values[index]) {
			continue
		}
		value := item.Values[index]
		height := percentOf(value, maximum)
		if value < 0 {
			height = percentOf(-value, maximum)
		}
		color := colorOf(item.Color, seriesIndex)
		switch mode {
		case "line", "area", "scatter", "bubble":
			markClass := classRoot + "-stem"
			if mode == "scatter" || mode == "bubble" {
				markClass = classRoot + "-dot"
			}
			mark := ui.Div().Class(markClass).Height(ui.Percent(height)).
				CustomProperty("--"+classRoot+"-color", color).
				Attr("data-series", strconv.Itoa(seriesIndex))
			column.Children(mark)
		case "pictorial":
			text := pictorialText(height)
			column.Children(ui.Div().Class(classRoot+"-bar").Height(ui.Percent(height)).
				CustomProperty("--"+classRoot+"-color", color).Text(text))
		default:
			className := classRoot + "-bar"
			if mode == "radar" || (mode == "combo" && seriesIndex > 0) {
				className = classRoot + "-stem"
			}
			column.Children(ui.Div().Class(className).Height(ui.Percent(height)).
				CustomProperty("--"+classRoot+"-color", color).
				Attr("data-series", strconv.Itoa(seriesIndex)))
		}
	}
	if mode == "area" {
		column.Children(ui.Div().Class(classRoot + "-fill").Height(ui.Percent(columnSpan(options, index, maximum))))
	}
	if mode == "combo" && len(options.Series) > 1 {
		return column
	}
	return column
}

func columnSpan(options Options, index int, maximum float64) float64 {
	low := 0.0
	high := 0.0
	seen := false
	for _, item := range options.Series {
		if index >= len(item.Values) || !finite(item.Values[index]) {
			continue
		}
		value := item.Values[index]
		if !seen || value < low {
			low = value
		}
		if !seen || value > high {
			high = value
		}
		seen = true
	}
	if !seen {
		return 0
	}
	return percentOf(high-low, maximum)
}

func pictorialText(height float64) string {
	count := int(height / pictorialDiv)
	if count < 1 && height > 0 {
		count = 1
	}
	if count > pictorialMax {
		count = pictorialMax
	}
	return strings.Repeat("▮", count)
}

func horizontalBody(options Options, mode string) *ui.Element {
	count := columnCount(options.Series)
	maximum := seriesMax(options.Series)
	list := ui.Div().Class(classRoot + "-body")
	for index := range count {
		row := ui.Div().Class(classRoot + "-row")
		row.Children(ui.Span().Class(classRoot + "-name").Text(labelAt(options.Labels, index)))
		for seriesIndex, item := range options.Series {
			if index >= len(item.Values) || !finite(item.Values[index]) {
				continue
			}
			bar := ui.Div().Class(classRoot+"-fill").
				Width(ui.Percent(percentOf(item.Values[index], maximum))).
				CustomProperty("--"+classRoot+"-color", colorOf(item.Color, seriesIndex))
			row.Children(ui.Div().Class(classRoot + "-track").Children(bar))
		}
		list.Children(row)
	}
	if mode == "bullet" {
		list.Class(classRoot + "-body " + classRoot + "-bullet")
	}
	return list
}

func gridBody(options Options, mode string) *ui.Element {
	count := columnCount(options.Series)
	grid := ui.Div().Class(classRoot + "-grid")
	for _, item := range options.Series {
		for index := range count {
			if index >= len(item.Values) || !finite(item.Values[index]) {
				continue
			}
			className := classRoot + "-cell"
			if mode == "voronoi" {
				className += " " + classRoot + "-tile"
			}
			grid.Children(ui.Div().Class(className).
				CustomProperty("--"+classRoot+"-color", colorOf(item.Color, 0)).
				Text(labelAt(options.Labels, index) + " " + formatNum(item.Values[index])))
		}
	}
	return ui.Div().Class(classRoot + "-body").Children(grid)
}

func bandBody(options Options, mode string) *ui.Element {
	count := columnCount(options.Series)
	list := ui.Div().Class(classRoot + "-body")
	for index := range count {
		total := 0.0
		for _, item := range options.Series {
			if index < len(item.Values) && finite(item.Values[index]) && item.Values[index] > 0 {
				total += item.Values[index]
			}
		}
		track := ui.Div().Class(classRoot + "-track")
		for seriesIndex, item := range options.Series {
			if index >= len(item.Values) || !finite(item.Values[index]) {
				continue
			}
			width := percentOf(item.Values[index], total)
			className := classRoot + "-seg"
			if mode == "marimekko" {
				className = classRoot + "-slice"
			}
			track.Children(ui.Div().Class(className).Width(ui.Percent(width)).
				CustomProperty("--"+classRoot+"-color", colorOf(item.Color, seriesIndex)))
		}
		list.Children(ui.Div().Class(classRoot+"-row").Children(
			ui.Span().Class(classRoot+"-name").Text(labelAt(options.Labels, index)),
			track,
		))
	}
	return list
}

func bumpBody(options Options) *ui.Element {
	count := columnCount(options.Series)
	list := ui.Div().Class(classRoot + "-body")
	for seriesIndex, item := range options.Series {
		row := ui.Div().Class(classRoot + "-row")
		row.Children(ui.Span().Class(classRoot + "-name").Text(item.Name))
		for index := range count {
			rank := rankAt(options.Series, seriesIndex, index)
			row.Children(ui.Span().Class(classRoot + "-cell").Text(strconv.Itoa(rank)))
		}
		list.Children(row)
	}
	return list
}

func rankAt(series []Series, seriesIndex, index int) int {
	if seriesIndex < 0 || seriesIndex >= len(series) || index >= len(series[seriesIndex].Values) {
		return 0
	}
	value := series[seriesIndex].Values[index]
	rank := 1
	for otherIndex, item := range series {
		if otherIndex == seriesIndex || index >= len(item.Values) || !finite(item.Values[index]) {
			continue
		}
		if item.Values[index] > value {
			rank++
		}
	}
	return rank
}

func summaryBody(options Options, mode string) *ui.Element {
	list := ui.Div().Class(classRoot + "-body")
	for _, item := range options.Series {
		text := formatNum(meanOf(item.Values))
		if mode == "box" && len(item.Values) == boxCount {
			text = joinNums(item.Values)
		}
		if mode == "raincloud" {
			text = joinNums(item.Values)
		}
		if mode == "violin" {
			text = formatNum(medianOf(item.Values))
		}
		list.Children(ui.Div().Class(classRoot+"-row").Children(
			ui.Span().Class(classRoot+"-name").Text(item.Name),
			ui.Span().Class(classRoot+"-cell").Text(text),
		))
	}
	return list
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

func medianOf(values []float64) float64 {
	clean := make([]float64, 0, len(values))
	for _, value := range values {
		if finite(value) {
			clean = append(clean, value)
		}
	}
	if len(clean) == 0 {
		return 0
	}
	for i := range clean {
		for j := i + 1; j < len(clean); j++ {
			if clean[j] < clean[i] {
				clean[i], clean[j] = clean[j], clean[i]
			}
		}
	}
	const half = 2
	return clean[len(clean)/half]
}

func joinNums(values []float64) string {
	parts := make([]string, len(values))
	for index, value := range values {
		parts[index] = formatNum(value)
	}
	return strings.Join(parts, " ")
}

func waterfallBody(options Options) *ui.Element {
	if len(options.Series) == 0 {
		return nil
	}
	item := options.Series[0]
	maximum := seriesMax(options.Series)
	plot := ui.Div().Class(classRoot + "-plot")
	for index, value := range item.Values {
		if !finite(value) {
			continue
		}
		className := classRoot + "-up"
		amount := value
		if value < 0 {
			className = classRoot + "-down"
			amount = -value
		}
		plot.Children(ui.Div().Class(classRoot+"-column").Children(
			ui.Div().Class(className).Height(ui.Percent(percentOf(amount, maximum))),
			ui.Span().Class(classRoot+"-label").Text(labelAt(options.Labels, index)),
		))
	}
	return ui.Div().Class(classRoot + "-body").Children(plot)
}

func dumbbellBody(options Options) *ui.Element {
	count := columnCount(options.Series)
	maximum := seriesMax(options.Series)
	list := ui.Div().Class(classRoot + "-body")
	for index := range count {
		low, high, ok := spanAt(options.Series, index)
		if !ok {
			continue
		}
		width := percentOf(high-low, maximum)
		list.Children(ui.Div().Class(classRoot+"-row").Children(
			ui.Span().Class(classRoot+"-name").Text(labelAt(options.Labels, index)),
			ui.Div().Class(classRoot+"-track").Children(
				ui.Div().Class(classRoot+"-fill").Width(ui.Percent(width)),
			),
			ui.Span().Text(formatNum(low)+"–"+formatNum(high)),
		))
	}
	return list
}

func spanAt(series []Series, index int) (float64, float64, bool) {
	low := 0.0
	high := 0.0
	seen := false
	for _, item := range series {
		if index >= len(item.Values) || !finite(item.Values[index]) {
			continue
		}
		value := item.Values[index]
		if !seen || value < low {
			low = value
		}
		if !seen || value > high {
			high = value
		}
		seen = true
	}
	return low, high, seen
}

func paintParts(kind string, options PartOptions, mode, extra string) *ui.Element {
	root := newRoot(kind, options.Title, options.Description, extra)
	return withBody(root, partsBody(options, mode))
}

func partsBody(options PartOptions, mode string) *ui.Element {
	if len(options.Slices) == 0 {
		return nil
	}
	switch mode {
	case "pie", "donut":
		return stackedParts(options, mode)
	case "waffle":
		return waffleParts(options)
	case "unit":
		return unitParts(options)
	case "regions":
		return regionParts(options)
	case "category":
		return stackedParts(options, "pie")
	default:
		return rowParts(options, mode)
	}
}

func stackedParts(options PartOptions, mode string) *ui.Element {
	total := sliceTotal(options.Slices)
	if total <= 0 {
		return nil
	}
	className := classRoot + "-track"
	if mode == "donut" {
		className += " " + classRoot + "-donut"
	}
	track := ui.Div().Class(className)
	legendRow := ui.Div().Class(classRoot + "-legend")
	for index, item := range options.Slices {
		track.Children(ui.Div().Class(classRoot+"-slice").
			Width(ui.Percent(percentOf(item.Value, total))).
			CustomProperty("--"+classRoot+"-color", colorOf(item.Color, index)))
		legendRow.Children(ui.Span().Class(classRoot + "-legend-label").Text(item.Name))
	}
	return ui.Div().Class(classRoot+"-body").Children(track, legendRow)
}

func rowParts(options PartOptions, mode string) *ui.Element {
	maximum := sliceMax(options.Slices)
	base := firstPositive(options.Slices)
	if mode != "funnel" {
		base = maximum
	}
	list := ui.Div().Class(classRoot + "-body")
	for index, item := range options.Slices {
		list.Children(ui.Div().Class(classRoot+"-row").Children(
			ui.Span().Class(classRoot+"-name").Text(item.Name),
			ui.Div().Class(classRoot+"-track").Children(
				ui.Div().Class(classRoot+"-fill").
					Width(ui.Percent(percentOf(item.Value, base))).
					CustomProperty("--"+classRoot+"-color", colorOf(item.Color, index)),
			),
		))
	}
	return list
}

func regionParts(options PartOptions) *ui.Element {
	grid := ui.Div().Class(classRoot + "-grid")
	for index, item := range options.Slices {
		grid.Children(ui.Div().Class(classRoot+"-cell").
			CustomProperty("--"+classRoot+"-color", colorOf(item.Color, index)).
			Text(item.Name + " " + formatNum(item.Value)))
	}
	return grid
}

func unitParts(options PartOptions) *ui.Element {
	grid := ui.Div().Class(classRoot + "-grid")
	for index, item := range options.Slices {
		count := int(item.Value)
		if count < 0 {
			count = 0
		}
		if count > unitCap {
			count = unitCap
		}
		for range count {
			grid.Children(ui.Span().Class(classRoot+"-cell").
				CustomProperty("--"+classRoot+"-color", colorOf(item.Color, index)).
				Text(item.Name))
		}
	}
	return grid
}

func waffleParts(options PartOptions) *ui.Element {
	total := sliceTotal(options.Slices)
	if total <= 0 {
		return nil
	}
	grid := ui.Div().Class(classRoot + "-grid")
	filled := 0
	for index, item := range options.Slices {
		share := int(percentOf(item.Value, total))
		for range share {
			if filled >= waffleTotal {
				break
			}
			grid.Children(ui.Span().Class(classRoot+"-cell "+classRoot+"-slice").
				CustomProperty("--"+classRoot+"-color", colorOf(item.Color, index)))
			filled++
		}
	}
	for filled < waffleTotal {
		grid.Children(ui.Span().Class(classRoot + "-cell " + classRoot + "-blank"))
		filled++
	}
	return grid
}

func sliceTotal(parts []Slice) float64 {
	total := 0.0
	for _, item := range parts {
		if finite(item.Value) && item.Value > 0 {
			total += item.Value
		}
	}
	return total
}

func sliceMax(parts []Slice) float64 {
	maximum := 0.0
	for _, item := range parts {
		if finite(item.Value) && item.Value > maximum {
			maximum = item.Value
		}
	}
	return maximum
}

func firstPositive(parts []Slice) float64 {
	for _, item := range parts {
		if finite(item.Value) && item.Value > 0 {
			return item.Value
		}
	}
	return 0
}

func paintGauge(kind, title string, value float64) *ui.Element {
	root := newRoot(kind, title, "", "")
	fill := ui.Div().Class(classRoot+"-fill").Width(ui.Percent(clampPercent(value))).
		CustomProperty("--"+classRoot+"-color", Palette(0))
	body := ui.Div().Class(classRoot+"-body").Children(
		ui.Div().Class(classRoot+"-track").Children(fill),
		ui.Span().Class(classRoot+"-value").Text(formatNum(value)),
	)
	return root.Children(body)
}

func paintTracker(kind string, options TrackerOptions) *ui.Element {
	root := newRoot(kind, options.Title, options.Description, "")
	if len(options.Cells) == 0 {
		return withBody(root, nil)
	}
	row := ui.Div().Class(classRoot + "-row")
	for _, cell := range options.Cells {
		status := classRoot + "-blank"
		switch cell.Status {
		case "good":
			status = classRoot + "-good"
		case "warn":
			status = classRoot + "-warn"
		case "bad":
			status = classRoot + "-bad"
		}
		row.Children(ui.Span().Class(classRoot + "-cell " + status).Text(cell.Name))
	}
	return withBody(root, row)
}
