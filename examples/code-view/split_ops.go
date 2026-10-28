package main

// toggleSplit opens or closes the second editor group. Opening the split
// mirrors the focused file into the right pane so the new pane is never empty.
func (a *App) toggleSplit() {
	if a.mode != "explorer" {
		a.showToast("info", "Split editor", "Switch to Explorer to open a second code pane.")
		a.refresh()

		return
	}

	if a.split {
		a.finishEdit()
		a.split = false
		a.focusGroup = 0
		a.resetScroll()
		a.refresh()

		return
	}

	a.split = true
	a.focusGroup = 1

	if len(a.tabs2) == 0 {
		if path := a.leftActivePath(); path != "" {
			a.tabs2 = []string{path}
			a.active2 = 0

			if _, ok := a.docs[path]; !ok {
				a.docs[path] = a.ws.openDoc(path)
			}
		}
	}

	a.resetScroll()
	a.refresh()
}

// leftActivePath returns the first group's active path.
func (a *App) leftActivePath() string {
	if a.active < 0 || a.active >= len(a.tabs) {
		return ""
	}

	return a.tabs[a.active]
}

// setFocusGroup moves keyboard and open-file routing to one pane.
func (a *App) setFocusGroup(group int) {
	if !a.split {
		a.focusGroup = 0

		return
	}

	if group != 0 && group != 1 {
		return
	}

	if group != a.focusGroup {
		a.finishEdit()
		a.focusGroup = group
		a.preview = false
		a.resetFind()
	}

	a.refresh()
}
