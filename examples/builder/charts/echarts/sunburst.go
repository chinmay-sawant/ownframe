package echarts

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

// Sunburst lists one row per node. data-depth is the level, and the arc width
// is the node's share of its siblings.
func Sunburst(options TreeOptions) *ui.Element {
	return extraRoot(
		"of-ech-sunburst",
		"Sunburst",
		options.Title,
		options.Description,
		extraSunList(options.Nodes),
	)
}

func extraSunList(nodes []Node) *ui.Element {
	if len(nodes) == 0 {
		return nil
	}
	list := ui.Div().Class("of-ech-sun-list")
	extraSunRows(list, nodes, 0)
	return list
}

func extraSunRows(list *ui.Element, nodes []Node, depth int) {
	total := extraLevelTotal(nodes)
	count := len(nodes)
	for index, node := range nodes {
		arc := ui.Div().Class("of-ech-sun-arc").
			Width(ui.Percent(extraPart(extraNodeWeight(node), total, count))).
			CustomProperty("--of-ech-color", extraPaint(node.Color, index))
		row := ui.Div().Class("of-ech-sun-row").
			Attr("data-depth", strconv.Itoa(depth)).
			PaddingLeft(ui.Percent(extraDepthPad(depth))).
			Children(
				arc,
				ui.Span().Class("of-ech-sun-name").Text(node.Name),
				ui.Span().Class("of-ech-sun-value").Text(extraText(node.Value)),
			)
		list.Children(row)
		if len(node.Children) > 0 {
			extraSunRows(list, node.Children, depth+1)
		}
	}
}
