package player

import "context"

// navFocusState stores page-level focus and scroll data for one screen.
type navFocusState struct {
	Focus   string
	ScrollX int
	ScrollY int
}

// ensureNavFocusStates creates the per-screen state map.
func (a *App) ensureNavFocusStates() {
	if a.navFocusStates == nil {
		a.navFocusStates = map[string]navFocusState{}
	}
}

// saveNavState stores the focus and scroll state of the current screen.
func (a *App) saveNavState() {
	if a.page == nil {
		return
	}
	a.ensureNavFocusStates()

	focus := a.page.FocusID()
	if focus == "" {
		focus = a.page.LastFocusID()
	}
	state := navFocusState{Focus: focus}
	x, y := a.page.ScrollOffset()
	state.ScrollX, state.ScrollY = x, y
	a.navFocusStates[a.view.Nav] = state

	a.storeNavStateRetained(a.view.Nav, state)
}

// restoreNavState applies a previously saved focus and scroll state.
func (a *App) restoreNavState(screen string) {
	if a.page == nil {
		return
	}
	a.ensureNavFocusStates()

	state, ok := a.navFocusStates[screen]
	a.page.SetData(a.view)
	if ok {
		a.page.SetScrollOffset(state.ScrollX, state.ScrollY)
		if state.Focus != "" {
			_ = a.page.Focus(context.Background(), state.Focus)
		}
	}
	a.storeNavStateRetained(screen, state)
}
