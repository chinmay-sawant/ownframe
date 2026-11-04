package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// screen composes the full Mailpulse page. The section builders (sidebar,
// topbar, contentBody) and appCSS live in other files; this file only wires
// them together.
func (a *app) screen() *ui.Element {
	return ui.Div().Class("mp-shell").
		Width(ui.Percent(100)).
		MinHeight(ui.Vh(100)).
		BackgroundColor(ui.Hex("#f6f6f5")).
		Children(
			a.sidebar(),
			ui.Div().Class("mp-right").Children(
				a.topbar(),
				ui.Div().Class("mp-content").Children(
					a.contentBody(),
				),
				a.toastBar(),
			),
		)
}

// toastBar renders a compact bottom-right dismissal toast while a toast is
// open. It returns nil when closed; ui.Children ignores nil children.
func (a *app) toastBar() *ui.Element {
	if !a.toastOpen {
		return nil
	}
	msg := a.toast
	if msg == "" {
		msg = "All changes saved."
	}
	return ui.Div().Class("mp-toast").
		Position(ui.PositionFixed).
		Right(ui.Px(16)).
		Bottom(ui.Px(16)).
		ZIndex(ui.Number(200)).
		MaxWidth(ui.Px(345)).
		Children(
			ui.Span().Class("mp-toast-msg").Text(msg).
				FlexGrow(ui.Number(1)).
				MinWidth(ui.Px(0)),
			chadcn.Button("Dismiss", chadcn.ButtonOptions{
				Variant: chadcn.ButtonOutlined,
				Size:    chadcn.ButtonSmall,
			}).FlexShrink(ui.Number(0)).OnClick(func(ctx context.Context) error {
				a.toastOpen = false
				return a.refresh(ctx)
			}),
		)
}
