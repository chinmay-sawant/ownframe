package scene

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

// boardCells lays out the 10x20 board cells.
func boardCells(f Frame) []cellV {
	out := make([]cellV, 0, game.Rows*game.Cols)

	for r := 0; r < game.Rows; r++ {
		for c := 0; c < game.Cols; c++ {
			out = append(out, cellV{
				ID:    cellID("b", r, c),
				Class: classAt(f, r, c),
				Left:  boardX + c*cellStep,
				Top:   boardY + r*cellStep,
			})
		}
	}

	return out
}

// classAt names the class for one board cell.
func classAt(f Frame, r, c int) string {
	k, ghost := f.cellKind(r, c)

	switch {
	case k.Valid() && ghost:
		return "k-g"
	case k.Valid():
		return "k-" + string(letter(k))
	default:
		return ""
	}
}

// previewCells lays out the 4x4 next-piece box.
func previewCells(k game.Piece) []cellV {
	set := nextSet(k)
	out := make([]cellV, 0, previewCap)

	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			class := ""
			if set[r][c] {
				class = "k-" + string(letter(k))
			}

			out = append(out, cellV{
				ID:    cellID("n", r, c),
				Class: class,
				Left:  nextX + c*cellStep,
				Top:   nextY + r*cellStep,
			})
		}
	}

	return out
}

// cellID names a board or preview cell, such as "b2-7" or "n0-3".
func cellID(prefix string, r, c int) string {
	return prefix + strconv.Itoa(r) + "-" + strconv.Itoa(c)
}
