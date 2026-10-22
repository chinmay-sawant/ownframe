package console

func (a *App) navItems() []NavItem {
	if a.view.Screen == screenLogin {
		return []NavItem{}
	}

	listOn := a.view.Screen == screenList || a.view.Screen == screenDetail

	return []NavItem{
		{Label: "Tickets", Action: "nav-list", On: listOn},
		{Label: "Report", Action: "nav-report", On: a.view.Screen == screenReport},
		{Label: "Settings", Action: "nav-settings", On: a.view.Screen == screenSettings},
	}
}
