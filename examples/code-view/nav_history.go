package main

// navMaxEntries caps the file navigation history. Back and forward walk this
// list; a new navigation drops everything ahead of the current entry, the way
// a browser and VS Code do.
const navMaxEntries = 100

// pushNavHistory records one user navigation. Consecutive duplicates are
// ignored and a new entry truncates the forward history.
func (a *App) pushNavHistory(path string) {
	if path == "" {
		return
	}

	if a.navAt >= 0 && a.navAt < len(a.nav) && a.nav[a.navAt] == path {
		return
	}

	if a.navAt+1 < len(a.nav) {
		a.nav = append([]string(nil), a.nav[:a.navAt+1]...)
	}

	if len(a.nav) == 0 {
		a.nav = []string{path}
		a.navAt = 0

		return
	}

	a.nav = append(a.nav, path)

	if len(a.nav) > navMaxEntries {
		a.nav = append([]string(nil), a.nav[len(a.nav)-navMaxEntries:]...)
	}

	a.navAt = len(a.nav) - 1
}

// canNavBack reports whether a previous file exists.
func (a *App) canNavBack() bool {
	return a.navAt > 0 && a.navAt < len(a.nav)
}

// canNavForward reports whether a next file exists.
func (a *App) canNavForward() bool {
	return a.navAt >= 0 && a.navAt+1 < len(a.nav)
}

// navBackPath is the file a back step would open.
func (a *App) navBackPath() string {
	if !a.canNavBack() {
		return ""
	}

	return a.nav[a.navAt-1]
}

// navForwardPath is the file a forward step would open.
func (a *App) navForwardPath() string {
	if !a.canNavForward() {
		return ""
	}

	return a.nav[a.navAt+1]
}

// goNavBack moves one step back in the file history.
func (a *App) goNavBack() bool {
	target := a.navBackPath()
	if target == "" {
		return false
	}

	if a.ws != nil && a.ws.ByPath[target] == nil {
		a.ws.ensureFileLoaded(target)
	}

	if a.ws == nil || a.ws.ByPath[target] == nil {
		a.nav = append(append([]string(nil), a.nav[:a.navAt-1]...), a.nav[a.navAt:]...)
		a.navAt--

		return false
	}

	a.navAt--

	a.openFileAt(a.activeGroup(), target, false)

	return true
}

// goNavForward moves one step forward in the file history.
func (a *App) goNavForward() bool {
	target := a.navForwardPath()
	if target == "" {
		return false
	}

	if a.ws != nil && a.ws.ByPath[target] == nil {
		a.ws.ensureFileLoaded(target)
	}

	if a.ws == nil || a.ws.ByPath[target] == nil {
		a.nav = append(append([]string(nil), a.nav[:a.navAt+1]...), a.nav[a.navAt+2:]...)

		return false
	}

	a.navAt++

	a.openFileAt(a.activeGroup(), target, false)

	return true
}

// resetNavHistory replaces the history with one entry.
func (a *App) resetNavHistory(path string) {
	if path == "" {
		a.nav = nil
		a.navAt = 0

		return
	}

	a.nav = []string{path}
	a.navAt = 0
}

// pruneNavHistory drops entries for files that no longer exist.
func (a *App) pruneNavHistory() {
	if len(a.nav) == 0 {
		a.navAt = 0

		return
	}

	current := ""
	if a.navAt >= 0 && a.navAt < len(a.nav) {
		current = a.nav[a.navAt]
	}

	kept := make([]string, 0, len(a.nav))

	for _, path := range a.nav {
		if a.ws == nil || a.ws.ByPath[path] == nil {
			continue
		}

		if len(kept) > 0 && kept[len(kept)-1] == path {
			continue
		}

		kept = append(kept, path)
	}

	if len(kept) == 0 {
		if current != "" && (a.ws == nil || a.ws.ByPath[current] != nil) {
			a.nav = []string{current}
			a.navAt = 0
		} else {
			a.nav = nil
			a.navAt = 0
		}

		return
	}

	a.nav = kept
	a.navAt = len(kept) - 1

	if current != "" {
		for i := len(kept) - 1; i >= 0; i-- {
			if kept[i] == current {
				a.navAt = i

				break
			}
		}
	}
}

// navBackTitle is the hover text for the back button.
func (a *App) navBackTitle() string {
	if target := a.navBackPath(); target != "" {
		return "Go Back to " + target + " (Alt+Left)"
	}

	return "Go Back (Alt+Left)"
}

// navForwardTitle is the hover text for the forward button.
func (a *App) navForwardTitle() string {
	if target := a.navForwardPath(); target != "" {
		return "Go Forward to " + target + " (Alt+Right)"
	}

	return "Go Forward (Alt+Right)"
}
