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
					{Label: "Documentation", Description: "Query syntax, guides, and API reference"},
					{Label: "Status page", Description: "Current availability for every region"},
				},
			},
			{
				ID:    "product-settings",
				Label: "Settings",
			},
		},
		OnOpenChange: func(ctx context.Context, id string) error {
			if id == "product-settings" {
				a.navOpen = ""
				a.settingsOpen = true
				return a.refresh(ctx)
			}
			if a.navOpen == id {
				a.navOpen = ""
			} else {
				a.navOpen = id
			}
			return a.refresh(ctx)
		},
	})
	// The panel is absolute, so the nav keeps a fixed width and never shifts
	// the triggers when it opens. It starts at the sidebar edge so the panel
	// clears the sidebar's fixed hit box.
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

	viewItem := func(label, tab string) chadcn.MenuItem {
		return chadcn.MenuItem{
			Label: label,
			OnSelect: func(ctx context.Context) error {
				a.tab = tab
				a.menubarOpen = ""
				return a.refresh(ctx)
			},
		}
	}

	menubar := chadcn.Menubar(chadcn.MenubarOptions{
		ID:    "top-menubar",
		Label: "Application",
		Flat:  true,
		Open:  a.menubarOpen,
		Menus: []chadcn.MenubarMenu{
			{
				ID:    "menu-file",
				Label: "File",
				Items: []chadcn.MenuItem{
					{
						Label:    "Refresh",
						Shortcut: "⌘R",
						OnSelect: func(ctx context.Context) error {
							a.menubarOpen = ""
							return a.showToast(ctx, "Refreshed", "All six services report within budget.")
						},
					},
					{
						Label:    "Export",
						Shortcut: "⇧⌘E",
						OnSelect: func(ctx context.Context) error {
							a.menubarOpen = ""
							return a.showToast(ctx, "Export queued", "A CSV of the last 24 hours will land in your inbox.")
						},
					},
					{Separator: true},
					{
						Label:    "New alert",
						Shortcut: "⌘N",
						OnSelect: func(ctx context.Context) error {
							a.menubarOpen = ""
							a.dialogOpen = true
							return a.refresh(ctx)
						},
					},
				},
			},
			{
				ID:    "menu-view",
				Label: "View",
				Items: []chadcn.MenuItem{
					viewItem("Overview", "overview"),
					viewItem("Logs", "logs"),
					viewItem("Traces", "traces"),
					viewItem("Incident room", "incident"),
				},
			},
			{
				ID:    "menu-help",
				Label: "Help",
				Items: []chadcn.MenuItem{
					{
						Label: "Documentation",
						OnSelect: func(ctx context.Context) error {
							a.menubarOpen = ""
							return a.showToast(ctx, "Documentation", "The Loglens handbook lives at docs.acme.dev/loglens.")
						},
					},
					{
						Label: "About Loglens",
						OnSelect: func(ctx context.Context) error {
							a.menubarOpen = ""
							return a.showToast(ctx, "Loglens 2.4.1", "Ownframe dashboard build · commit 9c4f7a0.")
						},
					},
				},
			},
		},
		OnOpenChange: func(ctx context.Context, id string) error {
			a.menubarOpen = id
			return a.refresh(ctx)
		},
	})
	// The flat menubar reports a wider box than its three triggers draw, so
	// pin it to the trigger row and let the spacer take the slack.
	menubar.Width(ui.Px(menubarWidth)).FlexShrink(ui.Number(0))

	return ui.Div().Class("app-row").
		Gap(ui.Px(2*topbarGap)).
		FlexWrap(ui.FlexWrapNowrap).
		FlexShrink(ui.Number(0)).
		Children(navSlot, menubar)
}
