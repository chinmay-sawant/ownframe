package ui

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// onClick clears the transient notice and routes the clicked action.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	a.state.notice = ""

	return a.handleClick(ctx, box.Action)
}

// onChange applies the search filter after a keystroke. The bind write
// already put the text in the view, so the table reads it here.
func (a *App) onChange(ctx context.Context, box ownframe.Box) error {
	if box.ID != "proc-search" {
		return nil
	}

	a.state.table.setQuery(a.view.Query)
	a.sync()

	return nil
}

// onKeyDown adds fast navigation keys. While the search field has focus the
// palette keeps every key, so typing never triggers a command.
func (a *App) onKeyDown(ctx context.Context, key string) error {
	if a.page.Form().FocusID() == "proc-search" {
		return nil
	}

	switch key {
	case "1":
		a.state.nav = "overview"
	case "2":
		a.state.nav = "processes"
	case "3":
		a.state.nav = "detail"
	case "t":
		if err := a.toggleTheme(ctx); err != nil {
			return err
		}
	case "r":
		a.refresh()
	case "slash":
		a.state.nav = "processes"
		a.sync()

		if err := a.page.Form().Focus(ctx, "proc-search"); err != nil {
			return err
		}

		return a.page.Redraw(ctx)
	default:
		return nil
	}

	a.sync()

	return a.page.Redraw(ctx)
}
