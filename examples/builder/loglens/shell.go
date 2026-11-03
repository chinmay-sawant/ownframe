package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func (a *app) screen() *ui.Element {
	root := chadcn.Screen(
		ui.Div().Children(a.topbar()),
		ui.Div().Class("app-body").Children(a.sidebar(), a.main()),
		a.settingsSheet(),
		a.globalDialogs(),
		a.logsDetailDrawer(),
		a.commandOverlay(),
	)
	if a.sidebarCollapsed {
		return root.Class("app-shell app-shell--collapsed")
	}
	return root.Class("app-shell")
}

func (a *app) topbar() *ui.Element {
	return ui.Header().Class("app-topbar").Children(
		ui.Div().Class("app-brand").Children(
			ui.Div().Class("app-brand-mark").Text("L"),
			ui.Div().Class("app-brand-text").Children(
				ui.Text("Loglens").FontWeight(ui.FontWeightBold).FontSize(ui.Px(16)),
				ui.Text("acme / production").FontSize(ui.Px(11)).Color(ui.ColorVar("--of-muted", ui.Hex("#737373"))),
			),
		),
		a.topMenus(),
		ui.Div().Class("app-topbar-spacer"),
		a.topPopovers(),
	)
}

func (a *app) sidebar() *ui.Element {
	return chadcn.Sidebar(chadcn.SidebarOptions{
		ID:        "app-sidebar",
		Label:     "Project navigation",
		Open:      true,
		Collapsed: a.sidebarCollapsed,
		Width:     ui.Px(232),
		Header: ui.Div().Class("app-sidebar-head").Children(
			ui.Div().Class("app-sidebar-project").Children(
				ui.Text("Acme observability").FontWeight(ui.FontWeightBold).FontSize(ui.Px(12)),
				ui.Text("Production").FontSize(ui.Px(11)).Color(ui.ColorVar("--of-muted", ui.Hex("#737373"))),
			),
			chadcn.Button("‹", chadcn.ButtonOptions{Size: chadcn.ButtonSmall, Variant: chadcn.ButtonText}).
				Attr("aria-label", "Collapse sidebar").
				OnClick(func(ctx context.Context) error {
					a.sidebarCollapsed = !a.sidebarCollapsed
					return a.refresh(ctx)
				}),
		),
		Content: ui.Div().Class("app-sidebar-nav").Children(
			sidebarLabel("Workspace"),
			a.navItem("◈", "Overview", "overview", ""),
			a.navItem("≡", "Logs", "logs", "24"),
			a.navItem("⌁", "Traces", "traces", ""),
			a.navItem("✳", "Incident room", "incident", "1"),
			sidebarLabel("Saved"),
			chadcn.SidebarNavItem(chadcn.SidebarItem{Label: "EU latency", Badge: "7"}),
			chadcn.SidebarNavItem(chadcn.SidebarItem{Label: "Checkout errors", Badge: "3"}),
			chadcn.SidebarNavItem(chadcn.SidebarItem{Label: "+ New view"}),
		),
		Footer: ui.Div().Class("app-sidebar-foot").Children(
			chadcn.Avatar(chadcn.AvatarOptions{Initials: "AK", Status: chadcn.AvatarStatusOnline}),
			ui.Div().Class("app-sidebar-user").Children(
				ui.Text("Ada Kim").FontWeight(ui.FontWeightBold).FontSize(ui.Px(12)),
				ui.Text("On call").FontSize(ui.Px(11)).Color(ui.ColorVar("--of-muted", ui.Hex("#737373"))),
			),
			chadcn.Button("⚙", chadcn.ButtonOptions{Variant: chadcn.ButtonText, Size: chadcn.ButtonSmall}).
				Attr("aria-label", "Open settings").
				OnClick(func(ctx context.Context) error {
					a.settingsOpen = true
					return a.refresh(ctx)
				}),
		),
	})
}

func (a *app) navItem(icon, label, tab, badge string) *ui.Element {
	return chadcn.SidebarNavItem(chadcn.SidebarItem{
		Icon:   ui.Text(icon),
		Label:  label,
		Active: a.tab == tab,
		Badge:  badge,
		Action: "nav:" + tab,
	})
}

func sidebarLabel(text string) *ui.Element {
	return ui.Text(text).FontSize(ui.Px(10)).FontWeight(ui.FontWeightBold).
		Color(ui.ColorVar("--of-muted", ui.Hex("#737373"))).Padding(ui.Edges(ui.Px(12), ui.Px(4)))
}

func (a *app) main() *ui.Element {
	return ui.Tag("main").Class("app-main").Children(
		a.pageHeader(),
		a.panelTabs(),
	)
}

func (a *app) pageHeader() *ui.Element {
	title := a.tabTitle()
	return ui.Div().Class("app-page-head").Children(
		chadcn.Breadcrumb([]chadcn.BreadcrumbItem{
			{Label: "Acme"},
			{Label: "Production"},
			{Label: title, Current: true},
		}),
		ui.Div().Class("app-page-title").Children(
			chadcn.H1(title),
			ui.Div().Class("app-page-badges").Children(
				chadcn.Badge("Live", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
				chadcn.Badge("EU", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
				chadcn.Kbd("⌘K"),
			),
		),
		ui.P().Class("app-page-sub").Text(a.tabSubtitle()),
		ui.Div().Class("app-page-actions").Children(
			chadcn.ButtonGroup(chadcn.ButtonGroupOptions{},
				chadcn.Button("Refresh", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined}).OnClick(func(ctx context.Context) error {
					a.loading = true
					return a.showToast(ctx, "Refreshed", "Dashboard data is up to date.")
				}),
				chadcn.Button("Export", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined}).OnClick(func(ctx context.Context) error {
					return a.showToast(ctx, "Export queued", "A CSV will land in your inbox.")
				}),
			),
			chadcn.Button("New alert", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).OnClick(func(ctx context.Context) error {
				a.dialogOpen = true
				return a.refresh(ctx)
			}),
		),
	)
}

func (a *app) panelTabs() *ui.Element {
	return chadcn.Tabs(chadcn.TabsOptions{
		Active: a.tab,
		OnSelect: func(ctx context.Context, tab string) error {
			a.tab = tab
			return a.refresh(ctx)
		},
		Items: []chadcn.TabItem{
			{ID: "overview", Label: "Overview", Content: a.overviewPanel()},
			{ID: "logs", Label: "Logs", Content: a.logsPanel()},
			{ID: "traces", Label: "Traces", Content: a.tracesPanel()},
			{ID: "incident", Label: "Incident room", Content: a.incidentPanel()},
		},
	})
}

func (a *app) tabTitle() string {
	switch a.tab {
	case "logs":
		return "Logs"
	case "traces":
		return "Traces"
	case "incident":
		return "Incident room"
	default:
		return "Overview"
	}
}

func (a *app) tabSubtitle() string {
	switch a.tab {
	case "logs":
		return "Search and filter the request stream across every service."
	case "traces":
		return "Follow a request through the gateway, auth, search, and the database."
	case "incident":
		return "Coordinate the response, capture decisions, and close the loop."
	default:
		return "Request volume, errors, and service health for the last 24 hours."
	}
}
