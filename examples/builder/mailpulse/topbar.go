package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// topbar renders the MailPulse top bar: a breadcrumb trail on the left and
// the search box, notification bell, and user avatar on the right. The
// search input reads its initial value from a.query; live keystrokes stay
// on the page form, so the demo needs no change handler.
func (a *app) topbar() *ui.Element {
	search := ui.Div().Class("mp-search").
		Display(ui.DisplayFlex).
		AlignItems(ui.AlignItemsCenter).
		Gap(ui.Px(8)).
		Height(ui.Px(36)).
		Padding(ui.Edges(ui.Px(0), ui.Px(12))).
		BorderRadius(ui.Px(10)).
		Border(ui.SolidBorder(ui.Px(1), ui.Hex("#e8e8e6"))).
		BackgroundColor(ui.Hex("#ffffff")).
		Children(
			ui.Text("○").Color(ui.Hex("#6b7280")),
			chadcn.InputWithOptions("mp-search", a.query, chadcn.InputOptions{
				Placeholder: "Search",
				Type:        "search",
				Size:        chadcn.InputSmall,
			}),
		)
	bell := chadcn.Button("○", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).Attr("aria-label", "Notifications").OnClick(func(ctx context.Context) error {
		a.toast = "No new notifications"
		a.toastOpen = true
		return a.refresh(ctx)
	})
	devOpen := a.page != nil && a.page.Window().DevTools()
	devVariant := chadcn.ButtonOutlined
	if devOpen {
		devVariant = chadcn.ButtonFilled
	}
	devtools := chadcn.Button("Dev", chadcn.ButtonOptions{
		Variant: devVariant,
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
	return ui.Header().Class("mp-topbar").Children(
		chadcn.Breadcrumb([]chadcn.BreadcrumbItem{
			{Label: "Workspace", Action: "nav:overview"},
			{Label: a.pageTitleOr("Engagement"), Current: true},
		}),
		ui.Div().Class("mp-topbar-spacer"),
		search,
		bell,
		chadcn.Avatar(chadcn.AvatarOptions{Initials: "JD"}),
		devtools,
		a.windowControls(),
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
