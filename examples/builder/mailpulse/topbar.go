package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// topbar renders the MailPulse top bar: a breadcrumb trail on the left and
// the search box, notification bell, and user avatar on the right. The
// search input reads its initial value from a.query; live keystrokes stay
// on the page form and the change handler mirrors them into filter state,
// so the results dropdown filters as you type. The bell and avatar toggle
// their popover panels; panel rows report data-action clicks to handleClick.
func (a *app) topbar() *ui.Element {
	searchBox := ui.Div().Class("mp-search").
		Display(ui.DisplayFlex).
		AlignItems(ui.AlignItemsCenter).
		Gap(ui.Px(8)).
		Height(ui.Px(36)).
		Padding(ui.Edges(ui.Px(0), ui.Px(12))).
		BorderRadius(ui.Px(10)).
		Border(ui.SolidBorder(ui.Px(1), ui.Hex("#e8e8e6"))).
		BackgroundColor(ui.Hex("#ffffff")).
		Children(
			ui.Span().Class("mp-search-icon").Children(
				ui.Span().Class("mp-search-mag").Attr("aria-hidden", "true"),
				ui.Span().Class("mp-search-handle").Attr("aria-hidden", "true"),
			),
			chadcn.InputWithOptions("mp-search", a.query, chadcn.InputOptions{
				Placeholder: "Search",
				Type:        "search",
				Size:        chadcn.InputSmall,
			}),
			a.searchClear(),
		)
	searchWrap := ui.Div().Class("mp-search-wrap").Children(
		searchBox,
		ui.Div().Class("mp-search-anchor").Children(a.searchDrop()),
	)
	devtoolsBtn := chadcn.Button("Dev", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).Attr("aria-label", "Toggle developer tools").
		Attr("title", "Developer tools").
		OnClick(func(ctx context.Context) error {
			open := !a.page.Window().DevTools()
			a.page.Window().SetDevTools(open)
			a.toast = "Developer tools " + devStateLabel(open)
			a.toastOpen = true
			return a.refresh(ctx)
		})
	devtools := ui.Span().Class("mp-dev-wrap").Children(devtoolsBtn)
	if a.page != nil && a.page.Window().DevTools() {
		devtools.Children(ui.Span().Class("mp-dev-dot").Attr("aria-hidden", "true"))
	}
	right := ui.Div().Class("mp-topbar-right").Children(
		searchWrap,
		a.notifTrigger(),
		a.userTrigger(),
		devtools,
		a.windowControls(),
	)
	return ui.Header().Class("mp-topbar").Children(
		chadcn.Breadcrumb([]chadcn.BreadcrumbItem{
			{Label: "Workspace", Action: "nav:overview"},
			{Label: a.pageTitleOr("Engagement"), Current: true},
		}),
		ui.Div().Class("mp-topbar-spacer"),
		right,
	)
}

func devStateLabel(open bool) string {
	if open {
		return "open"
	}
	return "closed"
}

func (a *app) windowControls() *ui.Element {
	return ui.Div().Class("mp-window-controls").Children(
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
