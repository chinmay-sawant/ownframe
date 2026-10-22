package workbook

import (
	"errors"
)

// CSVOptions controls import and export interpretation.
type CSVOptions struct {
	Delimiter         rune
	InterpretNumbers  bool
	InterpretFormulas bool
	WriteBOM          bool
	WriteValues       bool
	MaxCells          int
}

// DefaultCSVOptions is comma-delimited, turns numeric fields into numbers
// and a leading = into a formula, and caps a parse at two million cells.
func DefaultCSVOptions() CSVOptions {
	return CSVOptions{
		Delimiter:         ',',
		InterpretNumbers:  true,
		InterpretFormulas: true,
		MaxCells:          csvMaxCells,
	}
}

// ErrTooLarge reports a table beyond MaxCells.
var ErrTooLarge = errors.New("workbook: table is too large")

// PageSize is the default page for tables and workbook lists.
const PageSize = 50

// Table is a rectangular block of cells. Unequal row lengths survive, and
// the block may start anywhere.
type Table struct {
	Start Pos
	Cells [][]Cell
}

// Rows is the number of rows.
func (t *Table) Rows() int { return len(t.Cells) }

// Cols is the widest row.
func (t *Table) Cols() int {
	n := 0

	for _, row := range t.Cells {
		if len(row) > n {
			n = len(row)
		}
	}

	return n
}
