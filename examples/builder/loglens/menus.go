package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// topMenus renders the product navigation, the application menubar, and the
// account menu that sit between the brand and the top bar tools.
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
					{Label: "Logs", Description: "Search and filter the request stream"},
					{Label: "Traces", Description: "Follow one request across services"},
					{Label: "Incident room", Description: "Coordinate a live response"},
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
	nav.Position(ui.PositionRelative).
		ZIndex(ui.Number(60)).
		CustomProperty("--of-navigation-menu-width", "320px")

	// The open navigation panel flows below its trigger. A fixed-height slot
	// keeps it from resizing the 60px top bar as the panel opens.
	navSlot := ui.Div().Height(ui.Px(36)).Overflow(ui.OverflowVisible).Children(nav)

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

	account := chadcn.Button("", chadcn.ButtonOptions{Variant: chadcn.ButtonText}).
		Attr("aria-label", "Account").
		Height(ui.Px(36))
	account.Children(
		chadcn.Avatar(chadcn.AvatarOptions{
			Initials: "AK",
			Size:     chadcn.AvatarSmall,
			Status:   chadcn.AvatarStatusOnline,
		}),
		ui.Text("Account").FontSize(ui.Px(13)).FontWeight(ui.Number(500)),
		ui.Text("▾").Class("app-muted").Attr("aria-hidden", "true"),
	)

	user := chadcn.DropdownMenu(chadcn.DropdownMenuOptions{
		ID:      "user-menu",
		Label:   "Account",
		Open:    a.userMenuOpen,
		Trigger: account,
		Items: []chadcn.MenuItem{
			{
				Label: "Profile",
				OnSelect: func(ctx context.Context) error {
					a.userMenuOpen = false
					return a.showToast(ctx, "Profile", "Ada Kim · ada@acme.dev · staff SRE, on call this week.")
				},
			},
		},
		Groups: []chadcn.MenuGroup{
			{
				Label: "Project",
				Items: []chadcn.MenuItem{
					{
						Label: "Project settings",
						OnSelect: func(ctx context.Context) error {
							a.userMenuOpen = false
							a.settingsOpen = true
							return a.refresh(ctx)
						},
					},
				},
			},
			{
				Label: "Session",
				Items: []chadcn.MenuItem{
					{
						Label:  "Sign out",
						Danger: true,
						OnSelect: func(ctx context.Context) error {
							a.userMenuOpen = false
							return a.showToast(ctx, "Signed out", "This preview keeps the session open. Reload to sign in again.")
						},
					},
				},
			},
		},
		OnOpenChange: func(ctx context.Context, open bool) error {
			a.userMenuOpen = open
			return a.refresh(ctx)
		},
	})

	return ui.Div().Class("app-row").Children(
		navSlot,
		menubar,
		ui.Text("synced 14:32").Class("app-muted"),
		chadcn.Badge("on-call", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
		user,
	)
}
