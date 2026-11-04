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
					ui.Div().Class("mp-card").Children(
						a.contentBody(),
					),
				),
				a.toastBar(),
			),
		)
}

// toastBar renders the fixed bottom dismissal bar while a toast is open.
// It returns nil when closed; ui.Children ignores nil children.
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
		Left(ui.Px(0)).
		Right(ui.Px(0)).
		Bottom(ui.Px(0)).
		ZIndex(ui.Number(60)).
		Children(
			ui.Text(msg),
			chadcn.Button("Dismiss", chadcn.ButtonOptions{
				Variant: chadcn.ButtonOutlined,
				Size:    chadcn.ButtonSmall,
			}).OnClick(func(ctx context.Context) error {
				a.toastOpen = false
				return a.refresh(ctx)
			}),
		)
}
