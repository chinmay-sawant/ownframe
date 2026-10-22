package formula

// Limits bounds parsing and evaluation so one bad formula cannot stall the
// app. A zero field is replaced by the default.
type Limits struct {
	MaxLen        int // formula source length in runes
	MaxDepth      int // nesting depth, and dependency graph depth
	MaxRangeCells int // cells one range may cover
	MaxEvals      int // cell reads one recalculation pass may spend
	MaxRows       int // highest row number a reference may name
	MaxCols       int // highest column number a reference may name
}

// DefaultLimits are the limits the example ships with.
func DefaultLimits() Limits {
	return Limits{
		MaxLen:        limitLen,
		MaxDepth:      limitDepth,
		MaxRangeCells: limitCells,
		MaxEvals:      limitEvals,
		MaxRows:       limitRows,
		MaxCols:       limitCols,
	}
}

// Normalized fills every zero field from DefaultLimits.
func (l Limits) Normalized() Limits {
	d := DefaultLimits()

	if l.MaxLen <= 0 {
		l.MaxLen = d.MaxLen
	}

	if l.MaxDepth <= 0 {
		l.MaxDepth = d.MaxDepth
	}

	if l.MaxRangeCells <= 0 {
		l.MaxRangeCells = d.MaxRangeCells
	}

	if l.MaxEvals <= 0 {
		l.MaxEvals = d.MaxEvals
	}

	if l.MaxRows <= 0 {
		l.MaxRows = d.MaxRows
	}

	if l.MaxCols <= 0 {
		l.MaxCols = d.MaxCols
	}

	return l
}

// Rect is a rectangular block of cells. Bounds are zero-based and
// inclusive.
type Rect struct {
	MinRow int
	MinCol int
	MaxRow int
	MaxCol int
}

// Cells counts the cells the rectangle covers.
func (r Rect) Cells() int {
	if r.MaxRow < r.MinRow || r.MaxCol < r.MinCol {
		return 0
	}

	return (r.MaxRow - r.MinRow + 1) * (r.MaxCol - r.MinCol + 1)
}

// Contains reports whether the rectangle covers the cell.
func (r Rect) Contains(row, col int) bool {
	return row >= r.MinRow && row <= r.MaxRow && col >= r.MinCol && col <= r.MaxCol
}
