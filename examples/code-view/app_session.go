package main

import (
	"encoding/json"
	"strings"
)

// EnableSession restores the state saved for this workspace, if any, and
// starts persisting later changes. The tests call New directly, so they never
// read or write the user config directory.
func (a *App) EnableSession() {
	path, err := defaultSessionPath(a.ws.Root)
	if err != nil {
		return
	}

	a.sessionPath = path

	state, err := loadSession(path)
	if err != nil {
		a.showToast("info", "Session not restored", err.Error())

		return
	}

	state = state.sanitize(a.ws.Root)
	if state.Root == "" || !sameWorkspace(state.Root, a.ws.Root) {
		a.sessionReady = true
		a.persistSession()

		return
	}

	a.restoreState(state)
	a.sessionReady = true
	a.sessionRestored = true
	a.persistSession()
}

// restoreState applies one saved snapshot.
func (a *App) restoreState(state sessionState) {
	a.theme = validTheme(state.Theme)
	a.zoom = state.Zoom
	_ = a.page.SetTheme(a.activeThemeCSS())
	a.mode = "explorer"
	if state.Mode == "changes" {
		a.mode = "changes"
	}

	a.recent = cleanSessionPaths(state.Recents)
	a.tabs = nil
	a.active = 0
	a.tabs2 = nil
	a.active2 = 0

	if state.SplitRatio > 0 {
		a.splitRatio = clampSessionRatio(state.SplitRatio, sessionSplitMin, sessionSplitMax)
	}

	if len(state.BranchLogs) > 0 {
		a.branchLogs.Restore(state.BranchLogs)
	}

	for _, path := range cleanSessionPaths(state.Tabs) {
		if a.ws.ByPath[path] == nil {
			a.ws.ensureFileLoaded(path)
		}
		if a.ws.ByPath[path] == nil {
			continue
		}

		if _, ok := a.docs[path]; !ok {
			a.docs[path] = a.ws.openDoc(path)
		}

		a.tabs = append(a.tabs, path)
	}

	if len(a.tabs) > maxTabs {
		a.tabs = a.tabs[:maxTabs]
	}

	if len(a.tabs) > 0 {
		a.active = clampSessionIndex(state.Active, len(a.tabs))
	}

	for _, path := range cleanSessionPaths(state.Tabs2) {
		if a.ws.ByPath[path] == nil {
			a.ws.ensureFileLoaded(path)
		}
		if a.ws.ByPath[path] == nil {
			continue
		}

		if _, ok := a.docs[path]; !ok {
			a.docs[path] = a.ws.openDoc(path)
		}

		a.tabs2 = append(a.tabs2, path)
	}

	if len(a.tabs2) > maxTabs {
		a.tabs2 = a.tabs2[:maxTabs]
	}

	if len(a.tabs2) > 0 {
		a.active2 = clampSessionIndex(state.Active2, len(a.tabs2))
	}

	a.split = state.Split && len(a.tabs2) > 0
	a.focusGroup = 0
	if a.split && state.FocusGroup == 1 {
		a.focusGroup = 1
	}

	a.details = state.Details
	a.wrap = state.Wrap
	a.preview = state.Preview
	a.diffSplit = state.DiffSplit
	a.historyManual = state.HistoryManual
	if state.HistoryRatio > 0 {
		a.historyRatio = state.HistoryRatio
	}

	a.devOn = state.DevTools
	a.page.Window().SetDevTools(state.DevTools)

	if len(a.tabs) == 0 {
		a.openInitial()
	}

	if path := a.activeDocPath(); path != "" {
		a.addRecent(path)
	}

	a.ensureTabVisible(0)
	a.ensureTabVisible(1)
	a.resetScroll()
	a.refresh()
}

// activeDocPath is the focused group's active path.
func (a *App) activeDocPath() string {
	if doc := a.activeDoc(); doc != nil {
		return doc.Path
	}

	return ""
}

// sameWorkspace compares two workspace roots after cleaning.
func sameWorkspace(a, b string) bool {
	return strings.TrimSpace(a) == strings.TrimSpace(b)
}

// persistSession writes the state when it differs from the last snapshot.
func (a *App) persistSession() {
	if !a.sessionReady || a.sessionPath == "" {
		return
	}

	state := a.sessionState()
	data, err := json.Marshal(state)
	if err != nil {
		return
	}

	if string(data) == a.sessionLast {
		return
	}

	if err := saveSession(a.sessionPath, state); err != nil {
		return
	}

	a.sessionLast = string(data)
}

// sessionState snapshots the fields that survive a restart.
func (a *App) sessionState() sessionState {
	state := sessionState{
		Root:          a.ws.Root,
		Mode:          a.mode,
		Theme:         a.theme,
		Recents:       append([]string(nil), a.recent...),
		Tabs:          append([]string(nil), a.tabs...),
		Active:        a.active,
		Tabs2:         append([]string(nil), a.tabs2...),
		Active2:       a.active2,
		FocusGroup:    a.focusGroup,
		Split:         a.split,
		SplitRatio:    a.splitRatio,
		Details:       a.details,
		DevTools:      a.devOn,
		Wrap:          a.wrap,
		Preview:       a.preview,
		DiffSplit:     a.diffSplit,
		HistoryRatio:  a.historyRatio,
		HistoryManual: a.historyManual,
		Zoom:          a.uiZoom(),
	}

	branches := a.branchLogs.Snapshot()
	if len(branches) > 0 {
		state.BranchLogs = branches
	}

	return state
}
