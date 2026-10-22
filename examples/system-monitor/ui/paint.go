package ui

// paint redraws the retained operations in place: graph bars every frame,
// text only when a result or handler changed it. A nil display list means
// the page fell back to the bitmap path, so there is nothing to repaint.
func (a *App) paint() {
	d := a.page.Probe().Display()
	if d == nil {
		a.state.dirtyText = false

		return
	}

	for _, id := range panelOrder {
		a.paintGraph(d, a.state.panels[id])
	}

	if !a.state.dirtyText {
		return
	}

	a.paintText()
	a.state.dirtyText = false
}

// paintText rewrites the text runs from the current state.
func (a *App) paintText() {
	s := a.state

	for _, id := range panelOrder {
		p := s.panels[id]
		value := unavailable

		if p.have {
			value = p.last.Text
		}

		a.setText(id+"-value", value)
		a.setText(id+"-sub", p.last.Sub)
		a.setText(id+"-peak", peakText(p))
	}

	a.setText("side-at", clock(s.at))
	a.setText("side-mode", modeLine(s))
	a.setText("ov-at", "updated "+clock(s.at))
	a.setText("proc-new", newText(s.table.hasNew()))
	a.setText("proc-at", clock(s.table.shown.At))
	a.setText("sel-name", truncate(nonEmpty(s.sel.proc.Name), detailNameLen))
	a.setText("sel-status", detailStatus(s.sel))

	for _, f := range detailFields(s.sel) {
		a.setText("df-"+f.Key, f.Value)
	}
}

// setText changes one retained text run when its content differs.
func (a *App) setText(id, text string) {
	op := a.state.h.text[id]
	if op != nil && op.Text != text {
		op.Text = text
	}
}
