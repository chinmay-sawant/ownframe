package ui

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe/ui"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
)

// screen composes the whole window: title bar, rail, sidebar, workspace,
// status bar, and any overlay.
func (a *App) screen() *ui.Element {
	a.livePanes = map[string]bool{}
	body := ui.Div().ID("body").Class("qr-body").Children(a.rail())
	if a.sidebarOpen {
		body.Children(a.sidebar(), ui.Div().ID("split-sidebar").Class("qr-drag-handle").Attr("aria-label", "Resize sidebar"))
	}
	body.Children(a.workspace())
	root := ui.Div().ID("qr-root").Class("qr-root").Children(a.titleBar(), body, a.statusBar())
	root.Children(a.connMenu(), a.overflowMenu(), a.modalView(), a.toastView())
	return root
}

func (a *App) titleBar() *ui.Element {
	bar := ui.Header().ID("titlebar").Class("qr-titlebar").Attr("aria-label", "Title bar")
	bar.Children(ui.Div().ID("tb-brand").Class("qr-brand").Children(
		a.logo(logoSize),
		ui.Span().Text("Quarry"),
	))
	env := "local"
	switch a.cfg.EnvTag {
	case driver.EnvStaging:
		env = "staging"
	case driver.EnvProd:
		env = "prod"
	}
	engine := "SQLite"
	if a.cfg.Engine == driver.EnginePostgres {
		engine = "PostgreSQL"
	}
	pill := ui.Div().ID("tb-conn").Class("qr-pill").Attr("role", "button").
		Attr("aria-label", fmt.Sprintf("Connection %s (%s, %s)", a.connLabel(), engine, a.cfg.EnvTag)).
		Attr("aria-haspopup", "menu").Attr("aria-expanded", fmt.Sprint(a.menuOpen == "conn")).
		Children(
			a.icon("database", toneBrand, iconMedium),
			ui.Span().Class("qr-dot qr-dot-"+env).ID("tb-env-dot"),
			ui.Span().Text(a.connLabel()),
			ui.Span().Class("qr-badge").Text(engineIcon(a.cfg.Engine)),
			a.icon("chevron-down", toneMuted, iconSmall),
		).OnClick(func(ctx context.Context) error {
		if a.menuOpen == "conn" {
			a.menuOpen = ""
		} else {
			a.menuOpen = "conn"
		}
		return a.redraw(ctx)
	})
	bar.Children(pill)
	if a.cfg.ReadOnly {
		bar.Children(ui.Span().ID("tb-lock").Class("qr-lock").Attr("aria-label", "Read-only connection").Children(
			a.icon("lock", toneWarn, iconSmall), ui.Span().Text("READ-ONLY")))
	}
	bar.Children(ui.Div().ID("tb-drag").Class("qr-spacer"))
	bar.Children(ui.Div().ID("tb-palette").Class("qr-pill").Attr("role", "button").Attr("aria-label", "Open command palette").
		Children(a.icon("search", toneMuted, iconSmall), ui.Span().Class("qr-muted").Text("Search commands"), ui.Span().Class("qr-kbd").Text("Ctrl+K")).
		OnClick(func(ctx context.Context) error { return a.openPalette(ctx) }))
	themeLabel, themeIcon := "Light", "sun"
	if !a.dark {
		themeLabel, themeIcon = "Dark", "moon"
	}
	bar.Children(ui.Button("").ID("tb-theme").Class("qr-iconbtn").Attr("type", "button").
		Attr("aria-label", "Switch to "+themeLabel+" theme").Attr("title", "Switch to "+themeLabel+" theme").
		Children(a.icon(themeIcon, toneText, iconMedium)).
		OnClick(func(ctx context.Context) error { return a.toggleTheme(ctx) }))
	maxIcon, maxLabel := "maximize", "Maximize"
	if a.maximized() {
		maxIcon, maxLabel = "restore", "Restore"
	}
	bar.Children(ui.Div().ID("tb-winctl").Class("qr-winctl").Children(
		a.winButton("win-min", "Minimize", "minimize"),
		a.winButton("win-max", maxLabel, maxIcon),
		a.winButton("win-close", "Close", "close"),
	))
	return bar
}

func engineIcon(e driver.EngineKind) string {
	if e == driver.EnginePostgres {
		return "PG"
	}
	return "SQ"
}

func (a *App) winButton(action, label, glyph string) *ui.Element {
	cls := "qr-winbtn"
	if action == "win-close" {
		cls += " qr-winbtn-close"
	}
	tone := toneText
	return ui.Button("").ID(action).Class(cls).Attr("type", "button").Attr("data-action", action).Children(a.icon(glyph, tone, iconMedium)).
		Attr("aria-label", label).Attr("title", label)
}

func (a *App) rail() *ui.Element {
	type item struct{ id, glyph, label string }
	items := []item{
		{"connections", "database", "Connections"},
		{"explorer", "layers", "Explorer"},
		{"monitor", "activity", "Monitor"},
		{"history", "clock", "History"},
		{"settings", "settings", "Settings"},
	}
	rail := ui.Nav().ID("rail").Class("qr-rail").Attr("aria-label", "Activity rail")
	rail.Children(ui.Div().ID("rail-logo").Class("qr-rail-logo").Children(a.logo(railLogoSize)))
	for _, it := range items {
		it := it
		cls := "qr-rail-btn"
		if a.activeRail == it.id {
			cls += " qr-rail-btn-active"
		}
		tone := toneMuted
		if a.activeRail == it.id {
			tone = toneBrand
		}
		rail.Children(ui.Button("").ID("rail-"+it.id).Class(cls).Attr("type", "button").Children(a.icon(it.glyph, tone, iconLarge)).
			Attr("aria-label", it.label).Attr("title", it.label).
			OnClick(func(ctx context.Context) error { return a.railClick(ctx, it.id) }))
	}
	return rail
}

func (a *App) railClick(ctx context.Context, id string) error {
	switch id {
	case "connections":
		return a.openConnManager(ctx)
	case "explorer":
		a.activeRail = "explorer"
		a.sidebarOpen = !a.sidebarOpen
	case "monitor":
		a.activeRail = "monitor"
		a.openMonitorTab()
		a.startMonitor()
	case "history":
		return a.openHistory(ctx)
	case "settings":
		a.modal = "settings"
	}
	return a.redraw(ctx)
}

// connMenu is the title-bar connection dropdown.
func (a *App) connMenu() *ui.Element {
	if a.menuOpen != "conn" {
		return nil
	}
	menu := ui.Div().ID("conn-menu").Class("qr-menu").Attr("role", ariaRole("menu")).Attr("aria-label", "Connections")
	cur := a.cfg
	menu.Children(ui.Div().Class("qr-menu-head qr-muted").Text("Current"))
	menu.Children(ui.Div().ID("conn-menu-current").Class("qr-list-row qr-list-row-active").Text(cur.Name))
	list := a.savedConnections()
	if len(list) > 0 {
		menu.Children(ui.Div().Class("qr-menu-head qr-muted").Text("Recent"))
	}
	for i, sc := range list {
		sc := sc
		menu.Children(ui.Div().ID(fmt.Sprintf("conn-menu-%d", i)).Class("qr-list-row").Attr("role", ariaRole("menuitem")).Children(
			a.icon("database", toneMuted, iconSmall),
			ui.Span().Text(sc.Name),
			ui.Span().Class("qr-muted").Text(sc.EnvTag),
		).OnClick(func(ctx context.Context) error {
			a.menuOpen = ""
			if err := a.connect(ctx, sc.ConnectionConfig); err != nil {
				a.notify(err.Error())
			}
			return a.redraw(ctx)
		}))
	}
	menu.Children(ui.Div().ID("conn-menu-manage").Class("qr-list-row").Attr("role", ariaRole("menuitem")).Text("Manage connections...").
		OnClick(func(ctx context.Context) error { return a.openConnManager(ctx) }))
	return ui.Div().ID("conn-menu-scrim").Class("qr-menu-scrim").Children(menu).
		OnClick(func(ctx context.Context) error { a.menuOpen = ""; return a.redraw(ctx) })
}

func (a *App) toastView() *ui.Element {
	if a.toast == "" {
		return nil
	}
	return ui.Div().ID("toast").Class("qr-toast").Attr("role", ariaRole("status")).Text(a.toast)
}

// prodBanner is the top read-only banner shown for Prod connections.
func (a *App) prodBanner() *ui.Element {
	if a.cfg.EnvTag != driver.EnvProd {
		return nil
	}
	if a.prodUnlocked {
		return ui.Div().ID("prod-banner").Class("qr-banner").Attr("role", ariaRole("alert")).
			Text("Production connection - writes UNLOCKED")
	}
	b := ui.Div().ID("prod-banner").Class("qr-banner").Attr("role", ariaRole("alert")).
		Children(ui.Span().Text("Production connection - Read-Only Mode active"), ui.Div().Class("qr-spacer"))
	b.Children(ui.Button("Unlock writes").ID("prod-unlock").Class("qr-btn qr-btn-danger").Attr("type", "button").
		OnClick(func(ctx context.Context) error { return a.askUnlockProd(ctx) }))
	return b
}
