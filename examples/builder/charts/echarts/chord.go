package echarts

import "github.com/chinmay-sawant/ownframe/ui"

// Chord lists one row per link. Each row carries of-ech-chord.
func Chord(options FlowOptions) *ui.Element {
	return extraRoot(
		"of-ech-chord",
		"Chord",
		options.Title,
		options.Description,
		extraLinkRows("of-ech-chord", options.Links),
	)
}
