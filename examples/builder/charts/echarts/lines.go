package echarts

import "github.com/chinmay-sawant/ownframe/ui"

// Lines uses the sankey row: one link per row and a bar whose width is the
// percent of the widest link. Each row carries of-ech-lines.
func Lines(options FlowOptions) *ui.Element {
	return extraRoot(
		"of-ech-lines",
		"Lines",
		options.Title,
		options.Description,
		extraFlowBars("of-ech-lines", "of-ech-lines-bar", options.Links),
	)
}
