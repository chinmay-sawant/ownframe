package ui

// buildGrid draws the visible grid lines: one fill per column boundary and
// one per row boundary, instead of a border on every cell.
func (a *App) buildGrid() []Box {
	w := a.win
	if w.Empty() {
		return nil
	}

	out := make([]Box, 0, (w.C1-w.C0+uiGridPad)+(w.R1-w.R0+uiGridPad))
	h := (w.R1 - w.R0 + 1) * RowH
	for c := w.C0; c <= w.C1+1; c++ {
		out = append(out, Box{Class: "gline", X: cellX(c), Y: cellY(w.R0), W: 1, H: h})
	}

	width := (w.C1 - w.C0 + 1) * ColW
	for r := w.R0; r <= w.R1+1; r++ {
		out = append(out, Box{Class: "gline", X: cellX(w.C0), Y: cellY(r), W: width, H: 1})
	}

	return out
}

// buildSelection draws one rectangle for a multi-cell selection. It is
// clamped to the rendered window, so a whole-sheet selection stays one op.
func (a *App) buildSelection() *Box {
	r0, c0, r1, c1 := a.selection().Rect()
	if r0 == r1 && c0 == c1 {
		return nil
	}

	w := a.win
	if w.Empty() {
		return nil
	}

	r0, c0 = max(r0, w.R0), max(c0, w.C0)
	r1, c1 = min(r1, w.R1), min(c1, w.C1)
	if r1 < r0 || c1 < c0 {
		return nil
	}

	return &Box{
		Class: "selrect",
		X:     cellX(c0),
		Y:     cellY(r0),
		W:     (c1 - c0 + 1) * ColW,
		H:     (r1 - r0 + 1) * RowH,
	}
}
