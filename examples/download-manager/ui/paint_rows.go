package ui

// paintRows writes the live progress cells. A changed row is marked by its
// strip, which has no element id, so the dirty rect stays on that line.
func (a *App) paintRows() bool {
	changed := false

	for i, row := range a.view.Active {
		if i >= len(a.bindings.fill) {
			break
		}

		if a.paintRow(i, row) {
			a.page.MarkRect(a.bindings.row[i])
			changed = true
		}
	}

	return changed
}

// paintRow updates one active row's fill and readouts.
func (a *App) paintRow(i int, row Row) bool {
	changed := a.paintFill(i, row)

	if setText(a.bindings.pct[i], row.Progress()) {
		changed = true
	}

	if setText(a.bindings.speed[i], row.SpeedText()) {
		changed = true
	}

	if setText(a.bindings.eta[i], row.ETAText()) {
		changed = true
	}

	return changed
}

// paintFill sets the progress width. Zero hides the fill.
func (a *App) paintFill(i int, row Row) bool {
	op := a.bindings.fill[i]
	if op == nil {
		return false
	}

	w := 0.0
	if f := row.Fraction(); f > 0 {
		w = a.bindings.trackW[i] * f
	}

	if op.W == w {
		return false
	}

	op.W = w

	return true
}
