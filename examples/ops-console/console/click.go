package console

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

func (a *App) onClick(_ context.Context, box ownframe.Box) error {
	switch box.Action {
	case "sign-in":
		a.signIn()
	case "sign-out":
		a.signOut()
	case "nav-list":
		a.show(screenList)
	case "nav-settings":
		a.show(screenSettings)
	case "nav-report":
		a.show(screenReport)
	case "back-list":
		a.show(screenList)
	case "save-note":
		a.saveNote()
	default:
		a.clickMore(box.Action)
	}

	a.syncView()
	a.page.SetData(&a.view)

	return nil
}

func (a *App) clickMore(action string) {
	switch {
	case strings.HasPrefix(action, "open-"):
		a.openTicket(strings.TrimPrefix(action, "open-"))
	case strings.HasPrefix(action, "filter-"):
		a.setFilter(strings.TrimPrefix(action, "filter-"))
	case strings.HasPrefix(action, "status-"):
		a.setStatus(strings.TrimPrefix(action, "status-"))
	case action == "toggle-compact":
		a.view.Compact = !a.view.Compact
	case action == "toggle-night":
		a.toggleNight()
	case action == "site-pier":
		a.view.Site = sitePier
	case action == "site-annex":
		a.view.Site = siteAnnex
	}
}
