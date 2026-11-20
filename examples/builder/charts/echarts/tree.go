package echarts

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

// Tree lists every node in tree order. Rows are siblings, so padding-left is
// the absolute depth and a data-depth attribute records the same level.
func Tree(options TreeOptions) *ui.Element {
	return extraRoot("of-ech-tree", "Tree", options.Title, options.Description, extraTreeList(options.Nodes))
}

func extraTreeList(nodes []Node) *ui.Element {
	if len(nodes) == 0 {
		return nil
	}
	list := ui.Div().Class("of-ech-tree-list")
	extraTreeAdd(list, nodes, 0)
	return list
}

func extraTreeAdd(list *ui.Element, nodes []Node, depth int) {
	for index, node := range nodes {
		row := ui.Div().Class("of-ech-tree-row").
			Attr("data-depth", strconv.Itoa(depth)).
			PaddingLeft(ui.Percent(extraDepthPad(depth)))
		row.Children(
			extraSwatch(node.Color, index),
			ui.Span().Class("of-ech-tree-name").Text(node.Name),
			ui.Span().Class("of-ech-tree-value").Text(extraText(node.Value)),
		)
		list.Children(row)
		if len(node.Children) > 0 {
			extraTreeAdd(list, node.Children, depth+1)
		}
	}
}
