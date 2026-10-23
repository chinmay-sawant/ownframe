package ui

import (
	"strings"
	"time"
)

// toggleTheme applies the other mode and persists it.
func (a *App) toggleTheme() {
	dark := !a.view.Dark
	if err := a.backend.SetDark(dark); err != nil {
		a.view.Notice = "Theme not saved: " + err.Error()
	}

	a.view.Dark = dark
	if err := a.page.SetTheme(themeSource(dark)); err != nil {
		a.view.Notice = "Theme failed: " + err.Error()
	}
}

// addJob queues the URL in the form, or explains what is missing.
func (a *App) addJob() {
	raw := strings.TrimSpace(a.page.Form().FormValue("url"))
	if raw == "" {
		a.view.Notice = "Enter a URL first"

		return
	}

	dest := strings.TrimSpace(a.page.Form().FormValue("dest"))
	if err := a.backend.Add(AddRequest{URL: raw, Destination: dest}); err != nil {
		a.view.Notice = "Add failed: " + err.Error()

		return
	}

	a.page.Form().SetFormValue("url", "")
	a.view.Notice = "Queued"
	a.historyDirty = true
}

// refresh re-asks the active set, the summary, and the current page.
func (a *App) refresh() {
	a.lastSummary = time.Time{}
	a.needActive = true

	if err := a.askPage(); err != nil {
		a.historyDirty = true
	}
}

// control sends one row command; a failure becomes a notice.
func (a *App) control(act ControlAction, id string) {
	if err := a.backend.Control(Control{Action: act, ID: id}); err != nil {
		a.view.Notice = "Command failed: " + err.Error()

		return
	}

	if act == ControlRemove && a.detail != nil && a.detail.ID == id {
		a.clearDetail()
	}
}

// setFilter switches the history filter and resets the walk.
func (a *App) setFilter(f Filter) {
	if !a.pager.SetFilter(f) {
		return
	}

	a.view.Notice = ""

	if err := a.askPage(); err != nil {
		a.historyDirty = true
	}
}
