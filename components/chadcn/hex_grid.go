package chadcn

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

const maxHexColumns = 24

// HexCell is one honeycomb cell. Tone selects the palette slot, 1-based; zero
// or out-of-range tones render the muted empty cell. Label names the cell for
// assistive technology.
type HexCell struct {
	Label string
	Tone  int
}

// HexGridOptions describes a honeycomb status map. Columns sets the row width;
// zero keeps the CSS default. Cells render in order. When ID is set, each cell
// id is that value plus the cell index, so a tone change patches instead of
// taking the full-render path. Added or removed cells stay FULL by design.
type HexGridOptions struct {
	ID      string
	Label   string
	Columns int
	Cells   []HexCell
}

// HexGrid builds the property health honeycomb: a grid of pointy-top hexagon
// cells colored by tone. Each hexagon is three solid divs, a top triangle, a
// middle band, and a bottom triangle, so the map needs no canvas, SVG, or
// clip-path. The host owns the tones and the legend.
func HexGrid(options HexGridOptions) *ui.Element {
	label := options.Label
	if label == "" {
		label = "Status map"
	}
	root := ui.Div().Class("of-hex-grid").Attr("role", "group").Attr("aria-label", label)
	if options.ID != "" {
		root.ID(options.ID)
	}
	if options.Columns > 0 {
		root.CustomProperty("--of-hex-columns", hexColumns(options.Columns))
	}
	for index, cell := range options.Cells {
		element := HexCellElement(cell)
		if id := partID(options.ID, "-cell-"+strconv.Itoa(index)); id != "" {
			element.ID(id)
		}
		root.Children(element)
	}
	return root
}

// HexCellElement builds one honeycomb cell from three solid parts.
func HexCellElement(cell HexCell) *ui.Element {
	slot := 0
	if cell.Tone >= 1 && cell.Tone <= chartPaletteSize+1 {
		slot = cell.Tone
	}
	element := ui.Div().Class("of-hex").Attr("data-tone", hexToneSlot(slot)).Children(
		ui.Div().Class("of-hex-top").Attr("aria-hidden", "true"),
		ui.Div().Class("of-hex-mid").Attr("aria-hidden", "true"),
		ui.Div().Class("of-hex-bottom").Attr("aria-hidden", "true"),
	)
	if cell.Label == "" {
		return element.Attr("aria-hidden", "true")
	}
	return element.Attr("role", "img").Attr("aria-label", cell.Label).Attr("title", cell.Label)
}

func hexToneSlot(slot int) string {
	if slot <= 0 {
		return "0"
	}
	return chartSeriesSlot(slot - 1)
}

func hexColumns(columns int) string {
	if columns <= 0 {
		return "12"
	}
	if columns > maxHexColumns {
		return "24"
	}
	return strconv.Itoa(columns)
}
