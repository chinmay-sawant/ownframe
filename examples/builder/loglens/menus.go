package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// topMenus renders the product navigation and the application menubar that
// sit between the brand and the top bar tools. The account menu lives in
// topPopovers.
func (a *app) topMenus() *ui.Element {
	nav := chadcn.NavigationMenu(chadcn.NavigationMenuOptions{
		ID:    "top-nav",
		Label: "Product",
		Open:  a.navOpen,
		Items: []chadcn.NavigationItem{
			{
				ID:    "product-platform",
				Label: "Platform",
				Children: []chadcn.NavigationItem{
					{Label: "Logs", Description: "Search and filter the request stream", Action: "nav:logs"},
					{Label: "Traces", Description: "Follow one request across services", Action: "nav:traces"},
					{Label: "Incident room", Description: "Coordinate a live response", Action: "nav:incident"},
				},
			},
			{
				ID:    "product-resources",
				Label: "Resources",
				Children: []chadcn.NavigationItem{
					{Label: "Documentation", Description: "Query syntax, guides, and API reference", OnSelect: func(ctx context.Context) error {
						a.navOpen = ""
						return a.showToast(ctx, "Documentation", "Query syntax, guides, and API reference.")
					}},
					{Label: "Status page", Description: "Current availability for every region", OnSelect: func(ctx context.Context) error {
						a.navOpen = ""
						return a.showToast(ctx, "Status page", "All regions are operational.")
					}},
				},
			},
			{
				ID:    "product-settings",
				Label: "Settings",
			},
		},
		OnOpenChange: func(ctx context.Context, id string) error {
			wasOpen := a.navOpen == id
			a.closeHeaderPopovers()
			a.menubarOpen = ""
			if id == "product-settings" {
				a.navOpen = ""
				a.settingsOpen = true
				return a.refresh(ctx)
			}
			if wasOpen {
				a.navOpen = ""
			} else {
				a.navOpen = id
			}
			return a.refresh(ctx)
		},
	})
	// Keep the product links in the same column as the main content and reserve
	// a stable slot for the menu so opening its panel cannot move the header.
	nav.Position(ui.PositionRelative).
		ZIndex(ui.Number(60)).
		Width(ui.Px(navWidth)).
		CustomProperty("--of-navigation-menu-width", "300px")

	// A real slot holds the nav in the flex row and keeps the open panel
	// from resizing either the row or the top bar.
	navSlot := ui.Div().
		Width(ui.Px(navWidth)).
		Height(ui.Px(36)).
		FlexShrink(ui.Number(0)).
		MarginLeft(ui.Px(navOffset)).
		Overflow(ui.OverflowVisible).
		Children(nav)

	selectTab := func(ctx context.Context, tab string) error {
		a.menubarOpen = ""
		return a.setTab(ctx, tab)
	}
	toastItem := func(label, title, body string) chadcn.NavigationItem {
		return chadcn.NavigationItem{Label: label, OnSelect: func(ctx context.Context) error {
			a.menubarOpen = ""
			return a.showToast(ctx, title, body)
		}}
	}
	menu := chadcn.NavigationMenu(chadcn.NavigationMenuOptions{
		ID:    "top-menubar",
		Label: "Application",
		Open:  a.menubarOpen,
		Items: []chadcn.NavigationItem{
			{ID: "menu-file", Label: "File", Children: []chadcn.NavigationItem{
				{Label: "Refresh", OnSelect: func(ctx context.Context) error {
					a.menubarOpen = ""
					return a.refreshTelemetry(ctx)
				}},
				toastItem("Export", "Export queued", "A CSV of the last 24 hours will land in your inbox."),
				{Label: "New alert", OnSelect: func(ctx context.Context) error {
					a.menubarOpen = ""
					a.dialogOpen = true
					return a.refresh(ctx)
				}},
			}},
			{ID: "menu-view", Label: "View", Children: []chadcn.NavigationItem{
				{Label: "Overview", OnSelect: func(ctx context.Context) error { return selectTab(ctx, "overview") }},
				{Label: "Logs", OnSelect: func(ctx context.Context) error { return selectTab(ctx, "logs") }},
				{Label: "Traces", OnSelect: func(ctx context.Context) error { return selectTab(ctx, "traces") }},
				{Label: "Incident room", OnSelect: func(ctx context.Context) error { return selectTab(ctx, "incident") }},
			}},
			{ID: "menu-help", Label: "Help", Children: []chadcn.NavigationItem{
				toastItem("Documentation", "Documentation", "The Loglens handbook lives at docs.acme.dev/loglens."),
				toastItem("About Loglens", "Loglens 2.4.1", "Ownframe dashboard build · commit 9c4f7a0."),
			}},
		},
		OnOpenChange: func(ctx context.Context, id string) error {
			a.closeHeaderPopovers()
			a.navOpen = ""
			if a.menubarOpen == id {
				a.menubarOpen = ""
			} else {
				a.menubarOpen = id
			}
			return a.refresh(ctx)
		},
	})
	menu.Position(ui.PositionRelative).ZIndex(ui.Number(60)).
		Width(ui.Px(applicationMenuWidth)).FlexShrink(ui.Number(0)).
		CustomProperty("--of-navigation-menu-width", "152px")

	return ui.Div().Class("app-row").
		AlignItems(ui.AlignItemsFlexStart).
		Gap(ui.Px(topbarGap)).
		FlexWrap(ui.FlexWrapNowrap).
		FlexShrink(ui.Number(0)).
		Children(navSlot, menu)
}
