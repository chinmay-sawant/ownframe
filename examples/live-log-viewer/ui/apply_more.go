package ui

import "time"

// applySources refreshes the sidebar and falls back to all sources when the
// active one was deleted.
func (a *App) applySources(o out) bool {
	if o.gen != a.srcGen.Load() || o.err != nil {
		return false
	}

	if a.activeSource != "" && !hasSource(o.sources, a.activeSource) {
		a.activeSource = ""
		a.setNote("the filtered source is gone; showing all sources", time.Now())
		a.loadPage(intentFilter)
	}

	if sameSources(a.sources, o.sources) {
		return false
	}

	a.sources = o.sources

	return true
}

// sameSources reports two equal sidebar snapshots.
func sameSources(a, b []SourceInfo) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// hasSource reports whether a snapshot lists key.
func hasSource(srcs []SourceInfo, key string) bool {
	for _, s := range srcs {
		if s.Key == key {
			return true
		}
	}

	return false
}

// applyDetail opens a detail result unless the selection moved on.
func (a *App) applyDetail(o out) bool {
	if o.gen != a.detailGen.Load() || o.id != a.detailID {
		return false
	}

	if o.err != nil {
		a.setNote("detail load failed: "+o.err.Error(), time.Now())

		return true
	}

	a.detail = o.detail
	a.detailOpen = true
	a.scrollTop()

	return true
}

// applySettings loads the persisted view state once.
func (a *App) applySettings(o out) bool {
	if o.err != nil || a.settingsLoaded {
		return false
	}

	a.settingsLoaded = true
	a.dark = o.settings.Dark
	a.minSev = o.settings.Severity
	a.activeSource = o.settings.Source
	a.follow.SetFollow(o.settings.Follow)
	if err := a.applyTheme(); err != nil {
		a.setNote("theme sheet rejected: "+err.Error(), time.Now())
	}
	a.loadPage(intentFilter)

	return true
}

// setNote shows a transient status message.
func (a *App) setNote(text string, now time.Time) {
	a.note = text
	a.noteUntil = now.Add(uiNoteDur)
}
