package inputlab

import "context"

// onCopy serves the clipboard fields; other fields use the library.
func (a *App) onCopy(context.Context) (string, bool, error) {
	text := a.page.Form().FormValue(a.lastClip)
	if text == "" {
		a.setClip("nothing selected")
		return "", false, nil
	}
	a.setClip("copied: " + text)
	return text, true, nil
}

// onCut serves the clipboard fields and clears the value.
func (a *App) onCut(context.Context) (string, bool, error) {
	text := a.page.Form().FormValue(a.lastClip)
	if text == "" {
		a.setClip("nothing selected")
		return "", false, nil
	}
	a.pushClip()
	a.page.Form().SetFormValue(a.lastClip, "")
	a.setClip("cut: " + text)
	return text, true, nil
}

// onPaste lands in the last clip field when none is focused.
func (a *App) onPaste(ctx context.Context, text string) error {
	if a.page.Edit().FocusedField() == "" {
		a.focusLast(ctx)
	}
	a.setClip("pasted: " + text)
	return nil
}

// onSelectAll selects the last clip field, else the note.
func (a *App) onSelectAll(ctx context.Context) error {
	if a.page.Edit().FocusedField() != "" {
		return a.page.Edit().SelectAll(ctx)
	}
	if a.focusLast(ctx) {
		return a.page.Edit().SelectAll(ctx)
	}
	if !a.focusNote(ctx) {
		a.setClip("nothing to select")
		return nil
	}
	return a.page.Edit().SelectAll(ctx)
}
