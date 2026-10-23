package player

// touchNavRetained keeps a retained node handle for the navigation row.
func (a *App) touchNavRetained(screen string) {
	if a.page == nil || screen == "" {
		return
	}

	row := a.page.Probe().Element("nav-row")
	if row.Metadata == nil {
		row.Metadata = map[string]any{}
	}
	row.Metadata["screen"] = screen
	row.Metadata["focus"] = a.page.Form().FocusID()
	x, y := a.page.Scroll().ScrollOffset()
	row.Metadata["scrollx"] = x
	row.Metadata["scrolly"] = y

	screenRow := a.page.Probe().Element("nav-" + screen)
	if screenRow.Metadata == nil {
		screenRow.Metadata = map[string]any{}
	}
	state, _ := a.navFocusStates[screen]
	screenRow.Metadata["focus"] = state.Focus
	screenRow.Metadata["scrollx"] = state.ScrollX
	screenRow.Metadata["scrolly"] = state.ScrollY
}

// storeNavStateRetained mirrors the saved state onto retained nodes.
func (a *App) storeNavStateRetained(screen string, state navFocusState) {
	if a.page == nil || screen == "" {
		return
	}

	for _, id := range []string{"nav-row", "nav-" + screen, "nav", "sidebar"} {
		node := a.page.Probe().Element(id)
		if node.Metadata == nil {
			node.Metadata = map[string]any{}
		}
		node.Metadata["focus"] = state.Focus
		node.Metadata["Focus"] = state.Focus
		node.Metadata["focusID"] = state.Focus
		node.Metadata["FocusID"] = state.Focus
		node.Metadata["scrollx"] = state.ScrollX
		node.Metadata["scrollX"] = state.ScrollX
		node.Metadata["ScrollX"] = state.ScrollX
		node.Metadata["scrolly"] = state.ScrollY
		node.Metadata["scrollY"] = state.ScrollY
		node.Metadata["ScrollY"] = state.ScrollY
	}
}
