package nivo

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	boxValueCount = 5
	waffleCells   = 100
)

// Bar builds grouped vertical bars.
func Bar(options Options) *ui.Element {
	return seriesFrame("Bar", "of-nivo-bar-chart", options, func() *ui.Element {
		return heightColumns(options, "of-nivo-bar")
	})
}

// Line builds one stem per value.
func Line(options Options) *ui.Element {
	return seriesFrame("Line", "of-nivo-line-chart", options, func() *ui.Element {
		return heightColumns(options, "of-nivo-stem")
	})
}

// Radar builds one column per category.
func Radar(options Options) *ui.Element {
	return seriesFrame("Radar", "of-nivo-radar-chart", options, func() *ui.Element {
		return heightColumns(options, "of-nivo-stem")
	})
}

// ScatterPlot builds one dot per value.
func ScatterPlot(options Options) *ui.Element {
	return seriesFrame("Scatter plot", "of-nivo-scatter", options, func() *ui.Element {
		return heightColumns(options, "of-nivo-dot")
	})
}

// PolarBar builds one column per category.
func PolarBar(options Options) *ui.Element {
	return seriesFrame("Polar bar", "of-nivo-polar", options, func() *ui.Element {
		return heightColumns(options, "of-nivo-bar")
	})
}

// HeatMap builds one cell per value.
func HeatMap(options Options) *ui.Element {
	return seriesFrame("Heat map", "of-nivo-heat", options, func() *ui.Element {
		return valueGrid(options, "of-nivo-heat-cell")
	})
}

// Voronoi builds one tile per value.
func Voronoi(options Options) *ui.Element {
	return seriesFrame("Voronoi", "of-nivo-voronoi", options, func() *ui.Element {
		return valueGrid(options, "of-nivo-voronoi-cell")
	})
}

// SwarmPlot builds one dot label per value.
func SwarmPlot(options Options) *ui.Element {
	return seriesFrame("Swarm plot", "of-nivo-swarm", options, func() *ui.Element {
		return valueGrid(options, "of-nivo-swarm-dot")
	})
}

// Bullet builds one horizontal bar per value.
func Bullet(options Options) *ui.Element {
	return seriesFrame("Bullet", "of-nivo-bullet", options, func() *ui.Element {
		return horizontalSeries(options)
	})
}

// Marimekko builds one stacked column per label.
func Marimekko(options Options) *ui.Element {
	return seriesFrame("Marimekko", "of-nivo-marimekko", options, func() *ui.Element {
		return stackedRows(options)
	})
}

// Stream builds one stacked row per label.
func Stream(options Options) *ui.Element {
	return seriesFrame("Stream", "of-nivo-stream", options, func() *ui.Element {
		return stackedRows(options)
	})
}

// ParallelCoordinates builds one row of cells per series.
func ParallelCoordinates(options Options) *ui.Element {
	return seriesFrame("Parallel coordinates", "of-nivo-parallel", options, func() *ui.Element {
		return parallelRows(options)
	})
}

// Bump builds the rank of each series in each column. Rank 1 is the largest value.
func Bump(options Options) *ui.Element {
	return seriesFrame("Bump", "of-nivo-bump", options, func() *ui.Element {
		return rankRows(options, "of-nivo-bump-cell")
	})
}

// AreaBump builds the same ranks as Bump with the area-bump class.
func AreaBump(options Options) *ui.Element {
	return seriesFrame("Area bump", "of-nivo-area-bump", options, func() *ui.Element {
		return rankRows(options, "of-nivo-bump-cell")
	})
}

// BoxPlot prints a five-number summary when a series has five values, and the mean otherwise.
func BoxPlot(options Options) *ui.Element {
	return seriesFrame("Box plot", "of-nivo-box", options, func() *ui.Element {
		return boxRows(options)
	})
}

// Pie builds a horizontal stacked track.
func Pie(options PartOptions) *ui.Element {
	return partFrame("of-nivo-pie", "Pie", options, stackedSlices(options, ""))
}

// Funnel builds one row per slice, sized against the first positive value.
func Funnel(options PartOptions) *ui.Element {
	return partFrame("of-nivo-funnel", "Funnel", options, sliceRows(options, true))
}

// RadialBar builds one row per slice, sized against the largest slice.
func RadialBar(options PartOptions) *ui.Element {
	return partFrame("of-nivo-radial", "Radial bar", options, sliceRows(options, false))
}

// Choropleth builds labeled region cells. It is not a geographic map.
func Choropleth(options PartOptions) *ui.Element {
	return partFrame("of-nivo-choropleth", "Choropleth", options, regionCells(options))
}

// GeoMap builds labeled region cells. It is not a geographic map.
func GeoMap(options PartOptions) *ui.Element {
	return partFrame("of-nivo-geomap", "Geo map", options, regionCells(options))
}

// Waffle builds a 100-cell grid filled by slice share.
func Waffle(options PartOptions) *ui.Element {
	return partFrame("of-nivo-waffle", "Waffle", options, waffleCellsOf(options))
}

// Calendar builds one cell per day.
func Calendar(options CalendarOptions) *ui.Element {
	return dayFrame("Calendar", "of-nivo-calendar", options)
}

// TimeRange builds one cell per day.
func TimeRange(options CalendarOptions) *ui.Element {
	return dayFrame("Time range", "of-nivo-time-range", options)
}

// CirclePacking lists tree nodes with a depth attribute.
func CirclePacking(options TreeOptions) *ui.Element {
	return treeFrame("Circle packing", "of-nivo-pack", options)
}

// Icicle lists tree nodes with a depth attribute.
func Icicle(options TreeOptions) *ui.Element {
	return treeFrame("Icicle", "of-nivo-icicle", options)
}

// Sunburst lists tree nodes with a depth attribute.
func Sunburst(options TreeOptions) *ui.Element {
	return treeFrame("Sunburst", "of-nivo-sunburst", options)
}

// Tree lists tree nodes with a depth attribute.
func Tree(options TreeOptions) *ui.Element {
	return treeFrame("Tree", "of-nivo-tree", options)
}

// TreeMap lists tree nodes sized against the largest value.
func TreeMap(options TreeOptions) *ui.Element {
	return treeFrame("Tree map", "of-nivo-treemap", options)
}

// Chord builds one row per link.
func Chord(options FlowOptions) *ui.Element {
	return flowFrame("Chord", "of-nivo-chord", options)
}

// Network builds one row per node and one row per link.
func Network(options FlowOptions) *ui.Element {
	return flowFrame("Network", "of-nivo-network", options)
}

// Sankey builds one bar per link.
func Sankey(options FlowOptions) *ui.Element {
	return flowFrame("Sankey", "of-nivo-sankey", options)
}

func seriesFrame(kind, rootClass string, options Options, build func() *ui.Element) *ui.Element {
	body := emptyNote()
	if hasPoints(options) {
		body = build()
	}
	return newFrame(rootClass, kind, options.Title, options.Description, body)
}

func valueGrid(options Options, cellClass string) *ui.Element {
	grid := ui.Div().Class("of-nivo-grid")
	columns := columnCount(options)
	for seriesIndex, series := range options.Series {
		for index := range columns {
			value, ok := valueAt(series.Values, index)
			if !ok {
				continue
			}
			grid.Children(paint(ui.Span().Class(cellClass), series.Color, seriesIndex).
				Text(labelAt(options.Labels, index) + " " + formatNumber(value)))
		}
	}
	return grid
}

func horizontalSeries(options Options) *ui.Element {
	maximum := seriesMaximum(options.Series)
	list := ui.Div().Class("of-nivo-rows")
	columns := columnCount(options)
	for index := range columns {
		row := ui.Div().Class("of-nivo-row")
		row.Children(ui.Span().Class("of-nivo-name").Text(labelAt(options.Labels, index)))
		for seriesIndex, series := range options.Series {
			value, ok := valueAt(series.Values, index)
			if !ok {
				continue
			}
			bar := paint(ui.Div().Class("of-nivo-fill"), series.Color, seriesIndex).
				Width(ui.Percent(percentOf(value, maximum)))
			row.Children(ui.Div().Class("of-nivo-track").Children(bar))
		}
		list.Children(row)
	}
	return list
}

func stackedRows(options Options) *ui.Element {
	list := ui.Div().Class("of-nivo-rows")
	columns := columnCount(options)
	for index := range columns {
		total := 0.0
		for _, series := range options.Series {
			value, ok := valueAt(series.Values, index)
			if ok && value > 0 {
				total += value
			}
		}
		track := ui.Div().Class("of-nivo-track")
		for seriesIndex, series := range options.Series {
			value, ok := valueAt(series.Values, index)
			if !ok {
				continue
			}
			track.Children(paint(ui.Div().Class("of-nivo-slice"), series.Color, seriesIndex).
				Width(ui.Percent(percentOf(value, total))))
		}
		list.Children(ui.Div().Class("of-nivo-row").Children(
			ui.Span().Class("of-nivo-name").Text(labelAt(options.Labels, index)),
			track,
		))
	}
	return list
}

func parallelRows(options Options) *ui.Element {
	list := ui.Div().Class("of-nivo-rows")
	columns := columnCount(options)
	maximum := seriesMaximum(options.Series)
	for seriesIndex, series := range options.Series {
		row := ui.Div().Class("of-nivo-row")
		row.Children(ui.Span().Class("of-nivo-name").Text(series.Name))
		for index := range columns {
			value, ok := valueAt(series.Values, index)
			if !ok {
				continue
			}
			row.Children(paint(ui.Span().Class("of-nivo-cell"), series.Color, seriesIndex).
				Text(formatNumber(value)).
				Attr("data-share", strconv.FormatFloat(percentOf(value, maximum), 'f', 0, 64)))
		}
		list.Children(row)
	}
	return list
}

func rankRows(options Options, cellClass string) *ui.Element {
	list := ui.Div().Class("of-nivo-rows")
	columns := columnCount(options)
	for seriesIndex, series := range options.Series {
		row := ui.Div().Class("of-nivo-row")
		row.Children(ui.Span().Class("of-nivo-name").Text(series.Name))
		for index := range columns {
			ranks := ranksAt(options.Series, index)
			rank := 0
			if seriesIndex < len(ranks) {
				rank = ranks[seriesIndex]
			}
			row.Children(ui.Span().Class(cellClass).Text(strconv.Itoa(rank)))
		}
		list.Children(row)
	}
	return list
}

func boxRows(options Options) *ui.Element {
	list := ui.Div().Class("of-nivo-rows")
	for _, series := range options.Series {
		text := formatNumber(meanOf(series.Values))
		if len(series.Values) == boxValueCount {
			text = joinNumbers(series.Values)
		}
		list.Children(ui.Div().Class("of-nivo-row").Children(
			ui.Span().Class("of-nivo-name").Text(series.Name),
			ui.Span().Class("of-nivo-cell").Text(text),
		))
	}
	return list
}

func stackedSlices(options PartOptions, extra string) *ui.Element {
	total := sliceTotal(options.Slices)
	if total <= 0 {
		return nil
	}
	className := "of-nivo-track"
	if extra != "" {
		className += " " + extra
	}
	track := ui.Div().Class(className)
	for index, item := range options.Slices {
		track.Children(paint(ui.Div().Class("of-nivo-slice"), item.Color, index).
			Width(ui.Percent(percentOf(item.Value, total))))
	}
	return track
}

func sliceRows(options PartOptions, useFirst bool) *ui.Element {
	if len(options.Slices) == 0 {
		return nil
	}
	base := sliceMaximum(options.Slices)
	if useFirst {
		base = firstPositive(options.Slices)
	}
	list := ui.Div().Class("of-nivo-rows")
	for index, item := range options.Slices {
		list.Children(ui.Div().Class("of-nivo-row").Children(
			ui.Span().Class("of-nivo-name").Text(item.Name),
			ui.Div().Class("of-nivo-track").Children(
				paint(ui.Div().Class("of-nivo-fill"), item.Color, index).
					Width(ui.Percent(percentOf(item.Value, base))),
			),
		))
	}
	return list
}

func regionCells(options PartOptions) *ui.Element {
	if len(options.Slices) == 0 {
		return nil
	}
	grid := ui.Div().Class("of-nivo-grid")
	for index, item := range options.Slices {
		grid.Children(paint(ui.Div().Class("of-nivo-region"), item.Color, index).
			Text(item.Name + " " + formatNumber(item.Value)))
	}
	return grid
}

func waffleCellsOf(options PartOptions) *ui.Element {
	total := sliceTotal(options.Slices)
	if total <= 0 {
		return nil
	}
	grid := ui.Div().Class("of-nivo-grid")
	filled := 0
	for index, item := range options.Slices {
		share := int(percentOf(item.Value, total))
		for range share {
			if filled >= waffleCells {
				break
			}
			grid.Children(paint(ui.Span().Class("of-nivo-waffle-on"), item.Color, index))
			filled++
		}
	}
	for filled < waffleCells {
		grid.Children(ui.Span().Class("of-nivo-waffle-off"))
		filled++
	}
	return grid
}

func dayFrame(kind, rootClass string, options CalendarOptions) *ui.Element {
	body := emptyNote()
	if len(options.Days) > 0 {
		row := ui.Div().Class("of-nivo-row")
		for index, day := range options.Days {
			row.Children(paint(ui.Span().Class("of-nivo-day"), "", index).Text(day.Label))
		}
		body = row
	}
	return newFrame(rootClass, kind, options.Title, options.Description, body)
}

func treeFrame(kind, rootClass string, options TreeOptions) *ui.Element {
	body := emptyNote()
	if len(options.Nodes) > 0 {
		maximum := nodeMaximum(options.Nodes)
		list := ui.Div().Class("of-nivo-rows")
		addTreeRows(list, options.Nodes, 0, maximum)
		body = list
	}
	return newFrame(rootClass, kind, options.Title, options.Description, body)
}

func addTreeRows(list *ui.Element, nodes []Node, depth int, maximum float64) {
	for index, node := range nodes {
		list.Children(ui.Div().Class("of-nivo-node").Attr("data-depth", strconv.Itoa(depth)).Children(
			ui.Span().Class("of-nivo-name").Text(node.Name),
			paint(ui.Div().Class("of-nivo-fill"), node.Color, index).
				Width(ui.Percent(percentOf(node.Value, maximum))),
		))
		addTreeRows(list, node.Children, depth+1, maximum)
	}
}

func flowFrame(kind, rootClass string, options FlowOptions) *ui.Element {
	body := emptyNote()
	if len(options.Nodes) > 0 || len(options.Links) > 0 {
		maximum := 0.0
		for _, link := range options.Links {
			if finite(link.Value) && link.Value > maximum {
				maximum = link.Value
			}
		}
		list := ui.Div().Class("of-nivo-rows")
		for index, node := range options.Nodes {
			list.Children(ui.Div().Class("of-nivo-row").Children(
				paint(ui.Span().Class("of-nivo-swatch"), node.Color, index),
				ui.Span().Class("of-nivo-name").Text(node.Name),
			))
		}
		for index, link := range options.Links {
			list.Children(ui.Div().Class("of-nivo-row").Children(
				ui.Span().Class("of-nivo-name").Text(link.Source),
				ui.Div().Class("of-nivo-track").Children(
					paint(ui.Div().Class("of-nivo-fill"), "", index).
						Width(ui.Percent(percentOf(link.Value, maximum))),
				),
				ui.Span().Text(link.Target),
			))
		}
		body = list
	}
	return newFrame(rootClass, kind, options.Title, options.Description, body)
}
