package ui

// buildTabs lays out the sheet tabs along the frozen bottom edge.
func (a *App) buildTabs() []Box {
	y := a.scrollY + a.viewH - TabH
	out := make([]Box, 0, len(a.sheets)+1)
	x := a.scrollX + uiTabX
	for _, sh := range a.sheets {
		cls := "tab"
		if sh.ID == a.active {
			cls += " on"
		}

		out = append(out, Box{Action: act("sheet:" + sh.ID), Text: sh.Name, Class: cls, X: x, Y: y, W: uiTabW, H: TabH - uiChromePad})
		x += uiTabStep
	}

	return out
}

// formulaText is the raw source shown in the formula bar: the editor
// buffer while an editor is open, otherwise the active cell's source.
func (a *App) formulaText() string {
	if a.edit != nil {
		return a.edit.Text
	}

	s := a.selection()
	cell, _ := a.tiles.get(a.active, s.ActiveR, s.ActiveC)

	return cell.Raw
}
