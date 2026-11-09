package echarts

import "github.com/chinmay-sawant/ownframe/ui"

// Sankey lists one row per link. The bar width is that link's percent of the widest link.
func Sankey(options FlowOptions) *ui.Element {
	return extraRoot(
		"of-ech-sankey",
		"Sankey",
		options.Title,
		options.Description,
		extraFlowBars("of-ech-sankey-row", "of-ech-sankey-bar", options.Links),
	)
}
