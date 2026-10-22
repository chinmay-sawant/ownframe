package clipboard

import "context"

// onCopy returns the last field when no field is focused and writes the
// status line. A focused field is handled by the library.
func (a *App) onCopy(context.Context) (string, bool, error) {
	text := a.page.Form().FormValue(a.last)
	if text == "" {
		a.setStatus("nothing selected")

		return "", false, nil
	}

	a.setStatus("copied: " + text)

	return text, true, nil
}

// onCut returns the last field, clears it, and writes the status line.
func (a *App) onCut(context.Context) (string, bool, error) {
	text := a.page.Form().FormValue(a.last)
	if text == "" {
		a.setStatus("nothing selected")

		return "", false, nil
	}

	a.push()
	a.page.Form().SetFormValue(a.last, "")
	a.setStatus("cut: " + text)

	return text, true, nil
}

// onPaste focuses the last field when none is focused so the insert lands
// there, and writes the status line.
func (a *App) onPaste(ctx context.Context, text string) error {
	if a.page.Edit().FocusedField() == "" {
		a.focusLast(ctx)
	}

	a.setStatus("pasted: " + text)

	return nil
}

// onSelectAll focuses the last field and selects it when none is focused.
func (a *App) onSelectAll(ctx context.Context) error {
	if !a.focusLast(ctx) {
		a.setStatus("nothing to select")

		return nil
	}

	return a.page.Edit().SelectAll(ctx)
}

// setStatus stores the status line for the next draw.
func (a *App) setStatus(text string) {
	a.view.Status = text
	a.page.SetData(a.view)
}
