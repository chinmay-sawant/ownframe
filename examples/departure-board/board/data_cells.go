package board

import "fmt"

// buildCells tiles one row. Field tiles carry the id t-<row>-<field>-<k>;
// invisible gap tiles separate the fields. A gap paints no text op, so the
// flipper skips it, and its id is t-<row>-gap-<cellIndex> for uniqueness.
func buildCells(rowIndex int, r Row) []Cell {
	var cells []Cell

	add := func(field, value string) {
		for k := 0; k < len(value); k++ {
			cells = append(cells, Cell{
				ID:   fmt.Sprintf("t-%d-%s-%d", rowIndex, field, k),
				Ch:   value[k : k+1],
				Kind: field,
			})
		}
	}
	gap := func() {
		cells = append(cells, Cell{
			ID:   fmt.Sprintf("t-%d-gap-%d", rowIndex, len(cells)),
			Ch:   " ",
			Kind: "gap",
		})
	}

	add("time", r.Time)
	gap()
	add("dest", r.Dest)
	gap()
	add("plat", r.Platform)
	gap()
	add("stat", r.Status)

	return cells
}
