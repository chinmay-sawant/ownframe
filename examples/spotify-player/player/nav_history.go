package player

import "context"

// navigate records a screen change and drops any forward branch.
func (a *App) navigate(ctx context.Context, screen string) {
	if screen == a.view.Nav {
		return
	}

	a.saveNavState()

	a.navHistory = append(a.navHistory[:a.navAt+1], screen)
	a.navAt = len(a.navHistory) - 1
	a.view.Nav = screen
	a.updateNavButtons()

	a.restoreNavState(ctx, screen)
	a.touchNavRetained(screen)
}

// moveNav restores the previous or next screen when one exists.
func (a *App) moveNav(ctx context.Context, delta int) {
	at := a.navAt + delta
	if at < 0 || at >= len(a.navHistory) {
		return
	}

	a.saveNavState()

	a.navAt = at
	a.view.Nav = a.navHistory[at]
	a.updateNavButtons()

	a.restoreNavState(ctx, a.view.Nav)
	a.touchNavRetained(a.view.Nav)
}

// updateNavButtons keeps the arrows aligned with the current history entry.
func (a *App) updateNavButtons() {
	a.view.CanBack = a.navAt > 0
	a.view.CanForward = a.navAt+1 < len(a.navHistory)
}
