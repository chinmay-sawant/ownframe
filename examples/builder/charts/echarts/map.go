package echarts

import "github.com/chinmay-sawant/ownframe/ui"

// Map lays slices out as a region grid. Each cell names its region and value.
// It is not a geographic map.
func Map(options PartOptions) *ui.Element {
	return extraRoot("of-ech-map", "Map", options.Title, options.Description, extraMapGrid(options))
}

func extraMapGrid(options PartOptions) *ui.Element {
	if len(options.Slices) == 0 {
		return nil
	}
	grid := ui.Div().Class("of-ech-map-grid")
	for index, slice := range options.Slices {
		cell := ui.Div().Class("of-ech-map-cell").
			CustomProperty("--of-ech-color", extraPaint(slice.Color, index))
		cell.Children(
			ui.Span().Class("of-ech-map-name").Text(slice.Name),
			ui.Span().Class("of-ech-map-value").Text(extraText(slice.Value)),
		)
		grid.Children(cell)
	}
	return grid
}
