package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	// topbarHeight matches the .app-topbar height in appCSS and the native
	// window drag region.
	topbarHeight = 60
	// topbarPadding is the inline horizontal padding on the top bar.
	topbarPadding = 16
	// topbarGap keeps the chrome groups close without crowding the controls.
	topbarGap = 8
	// navLeft aligns the product navigation with the main content edge, past
	// the sidebar's fixed hit box.
	navLeft = 232
	// brandWidth is the mark plus the Loglens wordmark.
	brandWidth = 98
	// navOffset is the slot margin that puts the nav at navLeft after the
	// brand and the top bar's leading gap.
	navOffset = navLeft - topbarPadding - brandWidth - topbarGap
	// navWidth is the Platform/Resources/Settings trigger row width.
	navWidth = 280
	// menubarWidth is the three-trigger File/View/Help row. The flat menubar
	// otherwise reports extra width past its last trigger.
	menubarWidth = 80
	// fullViewport is the percentage that fills the window.
	fullViewport = 100
)

func (a *app) screen() *ui.Element {
	root := ui.Div().Class("app-shell").
		Width(ui.Percent(fullViewport)).
		MinHeight(ui.Vh(fullViewport)).
		BackgroundColor(ui.ColorVar("--of-background", ui.Hex("#0a0a0a"))).
		Children(
			// The fixed sidebar starts below the top bar and comes first so
			// it never wins the hit test for the top-left chrome.
			a.sidebar(),
			ui.Div().Children(a.topbar()),
			ui.Div().Class("app-body").Children(a.main()),
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

// topbar is the custom window chrome: the brand, the application menus, the
// status tools, and the window controls share one row that never wraps.
func (a *app) topbar() *ui.Element {
	return ui.Header().Class("app-topbar").
		Padding(ui.Edges(ui.Px(0), ui.Px(topbarPadding))).
		Gap(ui.Px(topbarGap)).
		Children(
			a.brand(),
			a.topMenus(),
			ui.Div().Class("app-topbar-spacer"),
			a.topPopovers(),
			a.windowControls(),
		)
}

// brand keeps the wordmark to one line; the project name already sits in the
// sidebar header, so repeating it here only widened the bar.
func (a *app) brand() *ui.Element {
	return ui.Div().Class("app-brand").FlexShrink(ui.Number(0)).Children(
		ui.Div().Class("app-brand-mark").Text("L"),
		ui.Text("Loglens").FontWeight(ui.FontWeightBold).FontSize(ui.Px(16)),
	)
}

func (a *app) windowControls() *ui.Element {
	return ui.Div().Class("app-window-controls").FlexShrink(ui.Number(0)).Children(
		windowButton("Minimize", "−", func() { a.page.Window().Minimize() }),
		windowButton("Maximize", "□", func() { a.page.Window().ToggleMaximize() }),
		windowButton("Close", "×", func() { a.page.Window().Close() }),
	)
}

func windowButton(label, glyph string, action func()) *ui.Element {
	return chadcn.Button(glyph, chadcn.ButtonOptions{Variant: chadcn.ButtonText, Size: chadcn.ButtonSmall}).
		Attr("aria-label", label).
		Attr("title", label).
		OnClick(func(context.Context) error {
			action()
			return nil
		})
}

func (a *app) sidebar() *ui.Element {
	side := chadcn.Sidebar(chadcn.SidebarOptions{
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
	// The top bar owns the strip above the sidebar. Starting the fixed rail
	// below it keeps the sidebar's full box out of the chrome hit test.
	side.Top(ui.Px(topbarHeight)).
		PaddingTop(ui.Px(0)).
		CSSRaw("height", "calc(100vh - 60px)")
	return side
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
				chadcn.Badge("EU", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
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
		Render: func(id string) *ui.Element {
			switch id {
			case "overview":
				return a.overviewPanel()
			case "logs":
				return a.logsPanel()
			case "traces":
				return a.tracesPanel()
			case "incident":
				return a.incidentPanel()
			}
			return nil
		},
		Items: []chadcn.TabItem{
			{ID: "overview", Label: "Overview"},
			{ID: "logs", Label: "Logs"},
			{ID: "traces", Label: "Traces"},
			{ID: "incident", Label: "Incident room"},
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
