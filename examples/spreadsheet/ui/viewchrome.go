package ui

// buildToolbar lays out the frozen toolbar and the formula bar.
func (a *App) buildToolbar() []Box {
	x := a.scrollX + uiBarX
	y := a.scrollY + uiBarY
	theme := "Theme: Light"
	if a.dark {
		theme = "Theme: Dark"
	}

	out := make([]Box, 0, uiBarCap)
	for _, b := range []Box{
		{Action: "undo", Text: "Undo", W: uiUndoW},
		{Action: "redo", Text: "Redo", W: uiUndoW},
		{Action: "import", Text: "Import", W: uiImportW},
		{Action: "export", Text: "Export", W: uiImportW},
		{Action: "theme", Text: theme, W: uiThemeW},
	} {
		b.Y, b.H, b.X = y, ToolH-uiToolTrim, x
		b.Class = "btn"
		out = append(out, b)
		x += b.W + uiBarGap
	}

	fy := a.scrollY + ToolH + uiFormulaPad
	out = append(out, Box{Action: "formula", Text: "fx", Class: "fx", X: a.scrollX + uiFxX, Y: fy, W: uiFxW, H: FormulaH - uiChromePad})
	out = append(out, Box{Action: "formula", Text: a.activeRef(), Class: "reflab", X: a.scrollX + uiRefX, Y: fy, W: uiRefW, H: FormulaH - uiChromePad})
	out = append(out, Box{Action: "formula", Text: a.formulaText(), Class: "ftext", X: a.scrollX + uiTextX, Y: fy, W: max(uiTextMin, a.viewW-uiTextX-uiChromePad), H: FormulaH - uiChromePad})

	return out
}
