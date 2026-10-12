package board

import "fmt"

// timeWidth and platWidth are the fixed widths of the time and platform
// fields. The destination and status fields take the widest value the seed
// and the status cycle can produce.
const (
	timeWidth = 5
	platWidth = 2
)

// blankTile is U+2800 BRAILLE PATTERN BLANK, the filler for padding cells.
// A whitespace-only tile (space, NBSP, even a zero-width character) gets no
// text op, and a zero-width op skews the paint positions of later tiles, so
// neither can pad. A braille blank paints a normal-width text op and renders
// as an inkless tile.
const blankTile = "\u2800"

// buildCells tiles one row on a fixed grid. Field tiles carry the id
// t-<row>-<field>-<k>; the three gap tiles keep the fields apart. Every cell
// exists at the same id on every feed, and every tile holds a printable rune:
// a space or a short value takes the filler, so the draw paints a text op for
// each tile. The real strings stay in Row.Time, Dest, Platform, and Status.
func buildCells(rowIndex int, r Row) []Cell {
	var cells []Cell

	add := func(field, value string, width int) {
		for k := 0; k < width; k++ {
			ch := blankTile
			if k < len(value) && value[k] != ' ' {
				ch = value[k : k+1]
			}

			cells = append(cells, Cell{
				ID:   fmt.Sprintf("t-%d-%s-%d", rowIndex, field, k),
				Ch:   ch,
				Kind: field,
			})
		}
	}
	gap := func() {
		cells = append(cells, Cell{
			ID:   fmt.Sprintf("t-%d-gap-%d", rowIndex, len(cells)),
			Ch:   blankTile,
			Kind: "gap",
		})
	}

	add("time", r.Time, timeWidth)
	gap()
	add("dest", r.Dest, destWidth)
	gap()
	add("plat", r.Platform, platWidth)
	gap()
	add("stat", r.Status, statWidth)

	return cells
}
