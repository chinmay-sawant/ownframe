package echarts

import "github.com/chinmay-sawant/ownframe/ui"

// Graph lists one row per node and one row per link.
func Graph(options FlowOptions) *ui.Element {
	nodes := extraGraphNodes(options.Nodes)
	links := extraLinkRows("of-ech-link", options.Links)
	if nodes == nil && links == nil {
		return extraRoot("of-ech-graph", "Graph", options.Title, options.Description, nil)
	}
	body := ui.Div().Class("of-ech-graph-body").Children(nodes, links)
	return extraRoot("of-ech-graph", "Graph", options.Title, options.Description, body)
}

func extraGraphNodes(nodes []Node) *ui.Element {
	if len(nodes) == 0 {
		return nil
	}
	list := ui.Div().Class("of-ech-graph-nodes")
	for index, node := range nodes {
		list.Children(ui.Div().Class("of-ech-node").Children(
			extraSwatch(node.Color, index),
			ui.Span().Class("of-ech-node-name").Text(node.Name),
			ui.Span().Class("of-ech-node-value").Text(extraText(node.Value)),
		))
	}
	return list
}
