package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// screen composes the full Discord-clone page. The section builders (rail,
// sidebar, mainCol, membersPanel) and the stylesheet live in other files;
// this file only wires them together with the custom drag-strip topbar, the
// popup dismiss layer, and the toast.
func (a *app) screen() *ui.Element {
	return ui.Div().ID("dc-shell").Class("discord-shell").
		Width(ui.Percent(100)).
		MinHeight(ui.Vh(100)).
		BackgroundColor(ui.Hex("#313338")).
		Children(
			a.topbar(),
			ui.Div().Class("dc-body").Children(
				a.rail(),
				a.sidebar(),
				a.mainCol(),
				a.membersPanel(),
			),
			// dc-float is a plain block wrapper (never a flex item with
			// position overrides) so the dismiss layer and the fixed
			// profile card position against the page instead of stacking
			// in-flow. overlays() stays a direct screen() child: the
			// toast is in-flow.
			ui.Div().ID("dc-float").Children(
				a.dismissLayer(),
				a.profileCard(),
			),
			a.overlays(),
		)
}

// topbar is the custom drag strip: the brand, the quick-switcher search
// button, the settings entry, and the window controls share one row. It is a
// plain div (not a header) so the stylesheet can mark it as the window drag
// region; OwnFrame skips drags that start on a button, so the controls stay
// clickable. Clicks report data-action hits to handleClick in actions.go.
func (a *app) topbar() *ui.Element {
	return ui.Div().ID("dc-topbar").Class("dc-topbar").Children(
		ui.Div().Class("dc-brand").Children(
			chadcn.Avatar(chadcn.AvatarOptions{Initials: "DC", Size: chadcn.AvatarSmall}),
			ui.Text("Discord clone").FontSize(ui.Px(13)).FontWeight(ui.FontWeightBold),
			chadcn.Badge("demo", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
		),
		chadcn.Tooltip(
			chadcn.TooltipOptions{ID: "dc-search-tip", CSSHover: true, Content: "Search (Ctrl+K)"},
			chadcn.Button("Search", chadcn.ButtonOptions{
				Icon:    ui.Text("⌕"),
				Variant: chadcn.ButtonOutlined,
				Size:    chadcn.ButtonSmall,
			}).ID("dc-top-search").
				Attr("aria-label", "Open search").
				Attr("data-action", "search:open"),
		),
		ui.Div().Class("dc-topbar-spacer"),
		chadcn.Button("⚙", chadcn.ButtonOptions{Variant: chadcn.ButtonText, Size: chadcn.ButtonSmall}).
			ID("dc-top-settings").
			Attr("aria-label", "Open settings").
			Attr("title", "Settings").
			Attr("data-action", "settings:open"),
		chadcn.Separator(chadcn.SeparatorOptions{Orientation: chadcn.SeparatorVertical}),
		a.dcWindowControls(),
	)
}

// dcWindowControls renders minimize/maximize/close as text buttons that
// report win:* actions; actions.go owns the window calls.
func (a *app) dcWindowControls() *ui.Element {
	return ui.Div().Class("dc-window-controls").Children(
		dcWindowButton("Minimize", "−", "win:min"),
		dcWindowButton("Maximize", "□", "win:max"),
		dcWindowButton("Close", "×", "win:close"),
	)
}

func dcWindowButton(label, glyph, action string) *ui.Element {
	return chadcn.Button(glyph, chadcn.ButtonOptions{Variant: chadcn.ButtonText, Size: chadcn.ButtonSmall}).
		ID("dc-win-"+label).
		Attr("aria-label", label).
		Attr("title", label).
		Attr("data-action", action)
}

// dismissLayer catches clicks outside open popups (search palette, profile
// popover, message menu, create dialog) and reports overlay:dismiss. It
// returns nil when nothing is open; ui.Children ignores nil children.
func (a *app) dismissLayer() *ui.Element {
	if !a.searchOpen && a.profileOpen == "" && a.msgMenu == "" && !a.createOpen {
		return nil
	}
	return ui.Div().ID("dc-dismiss").Class("dc-dismiss").
		Position(ui.PositionFixed).
		Top(ui.Px(0)).
		Left(ui.Px(0)).
		Width(ui.Percent(100)).
		Height(ui.Percent(100)).
		ZIndex(ui.Number(90)).
		BackgroundColor(ui.Hex("#00000000")).
		Attr("data-action", "overlay:dismiss")
}

// Toasts render through overlays() (see overlays.go); the old dcToast helper
// was removed to avoid stacking two toasts.
