package echarts

import "github.com/chinmay-sawant/ownframe/ui"

// Treemap fills a row with one tile per node. Tile width is that node's share
// of its siblings. Children repeat the same row inside the parent tile.
func Treemap(options TreeOptions) *ui.Element {
	return extraRoot("of-ech-treemap", "Treemap", options.Title, options.Description, extraTreemap(options.Nodes))
}

func extraTreemap(nodes []Node) *ui.Element {
	if len(nodes) == 0 {
		return nil
	}
	total := extraLevelTotal(nodes)
	row := ui.Div().Class("of-ech-treemap-row")
	for index, node := range nodes {
		row.Children(extraTile(node, index, total, len(nodes)))
	}
	return row
}

func extraTile(node Node, index int, total float64, count int) *ui.Element {
	tile := ui.Div().Class("of-ech-tile").
		Width(ui.Percent(extraPart(extraNodeWeight(node), total, count))).
		CustomProperty("--of-ech-color", extraPaint(node.Color, index))
	tile.Children(
		ui.Span().Class("of-ech-tile-name").Text(node.Name),
		ui.Span().Class("of-ech-tile-value").Text(extraText(node.Value)),
	)
	if len(node.Children) > 0 {
		tile.Children(extraTreemap(node.Children))
	}
	return tile
}
