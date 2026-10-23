package app

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// shareText is the link the share badge copies.
const shareText = "https://wisprflow.ai/insights"

// navActions maps a sidebar or voice-profile click to the page it opens.
// Allowlist only; demo boundary with no authz.
var navActions = map[string]string{
	"nav-dictation":  "dictation",
	"nav-notetaker":  "notetaker",
	"nav-insights":   "insights",
	"nav-dictionary": "dictionary",
	"nav-snippets":   "snippets",
	"nav-style":      "style",
	"nav-transforms": "transforms",
	"nav-scratchpad": "scratchpad",
	"nav-invite":     "invite",
	"nav-free-month": "free-month",
	"nav-settings":   "settings",
	"nav-help":       "help",
	"nav-voice":      "voice",
}

// onClick runs the control under the click: the sidebar, the get-app
// panel, the tabs, the streak chevrons, the share badge, and the mobile
// download.
func (a *App) onClick(_ context.Context, box ownframe.Box) error {
	switch box.Action {
	case "sidebar-toggle":
		a.view.SidebarCollapsed = !a.view.SidebarCollapsed
	case "get-app":
		a.view.GetApp = true
	case "get-app-close":
		a.view.GetApp = false
	case "get-app-store":
		ownframe.Write("https://wisprflow.ai/download")
		a.view.GetApp = false
		a.view.Note = "Download link copied"
	case "streak-prev":
		a.shiftStreak(1)
	case "streak-next":
		a.shiftStreak(-1)
	case "share":
		ownframe.Write(shareText)
		a.view.Note = "Share link copied"
	case "download":
		a.view.Note = "Download link sent to your phone"
	default:
		a.dispatchExtra(box.Action)
	}

	if page, ok := navActions[box.Action]; ok {
		a.view.ActivePage = page
		a.view.Note = ""
	}

	if name, ok := tabNames[box.Action]; ok {
		a.view.ActiveTab = name
		a.view.Note = ""
	}

	a.page.SetData(a.view)

	return nil
}

// show switches the active page without a click.
func (a *App) show(page string) {
	a.view.ActivePage = page
	a.page.SetData(a.view)
}
