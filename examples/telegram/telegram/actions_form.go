package telegram

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// onChange rebuilds the list after a search keystroke and swaps the theme
// after the dark-mode toggle.
func (a *App) onChange(_ context.Context, box ownframe.Box) error {
	switch box.ID {
	case "search":
		a.rebuild()
	case "dark":
		if err := a.page.SetTheme(themeSource(a.view.Dark)); err != nil {
			return err
		}
	}

	a.page.SetData(&a.view)

	return nil
}

// onSubmit sends the composer on Enter or NumpadEnter.
func (a *App) onSubmit(ctx context.Context) error {
	if a.page.Form().FocusID() != "compose" {
		return nil
	}

	a.send()
	if err := a.refocusComposer(ctx); err != nil {
		return err
	}

	a.rebuild()
	a.page.SetData(&a.view)

	return nil
}

// refocusComposer returns the focus to the composer after a composer button
// click or the keyboard's action key. A click outside a control blurs the
// form, which hides the keyboard; the phone keeps it open until back.
func (a *App) refocusComposer(ctx context.Context) error {
	if a.view.InsetBottom < keyboardMin {
		return nil
	}

	return a.page.Form().Focus(ctx, "compose")
}
