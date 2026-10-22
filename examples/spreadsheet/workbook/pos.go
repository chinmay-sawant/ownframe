package workbook

import (
	"errors"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/formula"
)

// Bounds are the largest A1 address the model accepts.
const (
	MaxRows = 1048576
	MaxCols = 18278
)

// ErrBadRef reports an address outside the sheet bounds.
var ErrBadRef = errors.New("workbook: bad cell reference")

// Pos is a zero-based cell address.
type Pos struct {
	Row int
	Col int
}

// Valid reports whether the position fits inside the sheet bounds.
func (p Pos) Valid() bool {
	return p.Row >= 0 && p.Row < MaxRows && p.Col >= 0 && p.Col < MaxCols
}

// String renders the position in A1 form.
func (p Pos) String() string { return ColName(p.Col) + strconv.Itoa(p.Row+1) }

// Rect covers Min through Max inclusive.
type Rect struct {
	Min Pos
	Max Pos
}

// Count is the number of cells the rectangle covers.
func (r Rect) Count() int {
	if r.Max.Row < r.Min.Row || r.Max.Col < r.Min.Col {
		return 0
	}

	return (r.Max.Row - r.Min.Row + 1) * (r.Max.Col - r.Min.Col + 1)
}

// Contains reports whether the rectangle covers p.
func (r Rect) Contains(p Pos) bool {
	return p.Row >= r.Min.Row && p.Row <= r.Max.Row &&
		p.Col >= r.Min.Col && p.Col <= r.Max.Col
}

// ColName renders a zero-based column index as letters, A through ZZZ.
func ColName(col int) string {
	if col < 0 || col >= MaxCols {
		return ""
	}

	var b [3]byte

	i := len(b)
	for col >= 0 && i > 0 {
		i--
		b[i] = byte('A' + col%alphaBase)
		col = col/alphaBase - 1
	}

	return string(b[i:])
}

// ParsePos parses an A1 address; $A$1 markers are accepted.
func ParsePos(s string) (Pos, error) {
	row, col, ok := formula.ParseRefWord(strings.TrimSpace(s))
	if !ok || row < 1 || row > MaxRows || col < 1 || col > MaxCols {
		return Pos{}, ErrBadRef
	}

	return Pos{Row: row - 1, Col: col - 1}, nil
}
