package ui

import "context"

// onKey routes keyboard shortcuts. A focused field keeps its keys, so
// typing in the search box never fires a shortcut. KeyDown does not draw;
// the handler renders what it changed.
func (a *App) onKey(ctx context.Context, key string) error {
	if a.page.Edit().FocusedField() != "" {
		return nil
	}

	switch key {
	case "arrowup":
		return a.moveSelection(ctx, -1, false)
	case "arrowdown":
		return a.moveSelection(ctx, 1, false)
	case "shift+arrowup":
		return a.moveSelection(ctx, -1, true)
	case "shift+arrowdown":
		return a.moveSelection(ctx, 1, true)
	case "pageup":
		a.page.ScrollBy(0, -(a.viewHeight() - 2*RowH))
	case "pagedown":
		a.page.ScrollBy(0, a.viewHeight()-2*RowH)
	case "home":
		a.jumpOldest()
	case "end":
		a.jumpNewest()
	case "f":
		a.toggleFollow()
	case "space":
		a.togglePause()
	case "enter":
		if a.selected != 0 {
			a.openDetail(a.selected)
		}
	case "escape":
		if a.detailOpen {
			a.closeDetail()
		}
	}

	return nil
}
