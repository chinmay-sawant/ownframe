package player

import "context"

func (a *App) saveNavRetention() navFocusState {
	if a == nil || a.page == nil {
		return navFocusState{}
	}
	f := a.page.FocusID()
	if f == "" {
		f = a.page.LastFocusID()
	}
	x, y := a.page.ScrollOffset()
	s := a.view.Nav
	a.ensureNavFocusStates()
	st := navFocusState{Focus: f, ScrollX: x, ScrollY: y}
	a.navFocusStates[s] = st
	a.storeNavStateRetained(s, st)
	return st
}

func (a *App) restoreNavRetention(s string) {
	if a == nil || a.page == nil {
		return
	}
	if s == "" {
		s = a.view.Nav
	}
	a.ensureNavFocusStates()
	st, ok := a.navFocusStates[s]
	if !ok {
		st = a.retainedNavState(s)
	}
	a.page.SetData(a.view)
	a.page.SetScrollOffset(st.ScrollX, st.ScrollY)
	if st.Focus != "" {
		_ = a.page.Focus(context.Background(), st.Focus)
	}
	if ok || st.Focus != "" || st.ScrollX != 0 || st.ScrollY != 0 {
		a.navFocusStates[s] = st
	}
	a.storeNavStateRetained(s, st)
	a.touchNavRetained(s)
}

func (a *App) retainedNavState(s string) navFocusState {
	st := navFocusState{}
	if a == nil || a.page == nil || s == "" {
		return st
	}
	n := a.page.RetainedElement("nav-" + s)
	if n == nil || n.Metadata == nil {
		n = a.page.RetainedElement("nav-row")
	}
	if n == nil || n.Metadata == nil {
		return st
	}
	m := n.Metadata
	for _, k := range []string{"focus", "Focus", "focusID", "FocusID"} {
		if v, ok := m[k].(string); ok && v != "" {
			st.Focus = v
			break
		}
	}
	for _, k := range []string{"scrollx", "scrollX", "ScrollX"} {
		if v, ok := m[k].(int); ok {
			st.ScrollX = v
			break
		}
		if v, ok := m[k].(float64); ok {
			st.ScrollX = int(v)
			break
		}
	}
	for _, k := range []string{"scrolly", "scrollY", "ScrollY"} {
		if v, ok := m[k].(int); ok {
			st.ScrollY = v
			break
		}
		if v, ok := m[k].(float64); ok {
			st.ScrollY = int(v)
			break
		}
	}
	return st
}

func (a *App) touchNavRetention(s string) {
	if a != nil {
		a.touchNavRetained(s)
	}
}
