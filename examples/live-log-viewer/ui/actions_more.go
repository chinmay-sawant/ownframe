package ui

import "time"

// toggleFollow detaches from the tail or jumps back to it.
func (a *App) toggleFollow() {
	if a.follow.Follow {
		a.follow.SetFollow(false)
		a.freeze()
		a.setNote("following off; the loaded page is frozen", time.Now())

		return
	}

	a.jumpNewest()
}

// togglePause stops or resumes display updates without stopping ingestion.
func (a *App) togglePause() {
	if a.follow.Paused {
		a.follow.Resume()
		a.setNote("display resumed at the reading position", time.Now())
	} else {
		a.follow.Pause()
		a.setNote("display paused; new entries keep counting", time.Now())
	}

	a.saveSettings()
}

// jumpNewest loads the live tail and clears the unread count.
func (a *App) jumpNewest() {
	a.follow.SetFollow(true)
	a.follow.ClearUnread()
	a.loadPage(intentNewest)
	a.saveSettings()
}

// pickSev sets the minimum severity, or clears it when it is already on.
func (a *App) pickSev(name string) {
	if a.minSev == name {
		name = ""
	}

	a.minSev = name
	a.applyFilterChange()
}

// clearFilters drops the text and severity filters.
func (a *App) clearFilters() {
	a.page.Form().SetFormValue("search", "")
	a.minSev = ""
	a.filters.Apply("")
	a.saveSettings()
	a.loadPage(intentFilter)
}

// applyFilterChange applies the pending filter state to a fresh page load.
func (a *App) applyFilterChange() {
	a.filters.Apply(a.page.Form().FormValue("search"))
	a.saveSettings()
	a.loadPage(intentFilter)
}

// pickSource selects one source or all of them.
func (a *App) pickSource(key string) {
	if key == "all" {
		key = ""
	}

	a.activeSource = key
	a.applyFilterChange()
}
