package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// topbar is the inventory chrome: the brand, the product nav, the search box,
// and the alerts and account controls on one row.
func (a *app) topbar() *ui.Element {
	return ui.Header().Class("inv-topbar").Children(
		ui.Div().Class("inv-brand").Children(
			ui.Span().Class("inv-brand-mark").Text("⬢"),
			ui.Text("Northwind Goods").Class("inv-brand-name"),
		),
		ui.Nav().Class("inv-nav").Children(
			invNavItem("Inventory", true).Attr("data-action", "nav:inventory"),
			invNavItem("Purchase orders", false),
			invNavItem("Receiving", false),
			invNavItem("Picking", false),
			invNavItem("Suppliers", false),
			invNavItem("Reports", false),
		),
		ui.Div().Class("inv-spacer"),
		a.searchBox(),
		a.alertsPopover(),
		a.accountMenu(),
		ui.Div().Class("inv-window-controls").Children(
			invWinButton("Minimize", "−", "win-min"),
			invWinButton("Maximize", "□", "win-max"),
			invWinButton("Close", "×", "win-close"),
		),
	)
}

// alertsPopover is the stock-alert inbox: an "Alerts" text button with a red
// unread dot, opening a popover with the two zero-on-hand rows.
func (a *app) alertsPopover() *ui.Element {
	unreadDot := ui.Text("●").
		Attr("aria-hidden", "true").
		FontSize(ui.Px(8)).
		Color(ui.ColorVar("--of-danger", ui.Hex("#dc2626")))

	trigger := chadcn.Button("Alerts", chadcn.ButtonOptions{
		Variant: chadcn.ButtonText,
		Size:    chadcn.ButtonSmall,
	}).Children(unreadDot)

	// Pin the popover slot while it opens: the open root would otherwise
	// size to its 288px panel and shove the search box left. This mirrors
	// the loglens notifications popover, which pins its slot the same way.
	root := chadcn.Popover(chadcn.PopoverOptions{
		ID:      "notifications-popover",
		Open:    a.notifOpen,
		Side:    chadcn.PopupBottom,
		Align:   chadcn.PopoverAlignEnd,
		Trigger: trigger,
		OnOpenChange: func(ctx context.Context, open bool) error {
			a.notifOpen = open
			return a.refresh(ctx)
		},
	},
		ui.Div().Class("inv-row inv-row--between").Children(
			ui.Text("Stock alerts").FontSize(ui.Px(13)).FontWeight(ui.FontWeightBold),
			chadcn.Badge("2 new", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
		),
		chadcn.Separator(chadcn.SeparatorOptions{}),
		chadcn.Item(chadcn.ItemOptions{
			Title:       "Cold brew filter packs",
			Description: "0 on hand · 1,200 inbound today",
		}),
		chadcn.Item(chadcn.ItemOptions{
			Title:       "Ceremonial matcha",
			Description: "0 on hand · reorder advised",
		}),
		chadcn.Separator(chadcn.SeparatorOptions{}),
		ui.Text("Updated just now").Class("inv-muted"),
	)
	// Keep the alerts slot fixed while its panel opens.
	return root.Width(ui.Px(64)).FlexShrink(ui.Number(0))
}

// accountMenu is the avatar trigger for the session menu. The trigger wraps
// the avatar in a text button, and the slot keeps a fixed size so the bar
// does not jump while the menu opens.
func (a *app) accountMenu() *ui.Element {
	trigger := chadcn.Button("", chadcn.ButtonOptions{
		Variant: chadcn.ButtonText,
	}).Attr("aria-label", "Account").Attr("title", "Account").
		Width(ui.Px(48)).Height(ui.Px(36)).
		Children(chadcn.Avatar(chadcn.AvatarOptions{
			Initials: "AK", Size: chadcn.AvatarSmall, Status: chadcn.AvatarStatusOnline,
		}))

	menu := chadcn.DropdownMenu(chadcn.DropdownMenuOptions{
		ID:       "user-menu",
		AlignEnd: true,
		Label:    "Account",
		Open:     a.userMenuOpen,
		Trigger:  trigger,
		Items: []chadcn.MenuItem{
			{
				Label: "Profile",
				OnSelect: func(ctx context.Context) error {
					a.userMenuOpen = false
					return a.refresh(ctx)
				},
			},
		},
		Groups: []chadcn.MenuGroup{
			{
				Label: "Session",
				Items: []chadcn.MenuItem{
					{
						Label:  "Sign out",
						Danger: true,
						OnSelect: func(ctx context.Context) error {
							a.userMenuOpen = false
							return a.refresh(ctx)
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
	// Keep the account slot fixed while its menu opens.
	menu.Width(ui.Px(48)).Height(ui.Px(36)).FlexShrink(ui.Number(0))
	return menu
}

// invWinButton is one window control in the topbar. It emits a data-action
// for the host click handler instead of an OnClick so CustomChrome
// drag-vs-button hit-testing keeps working.
func invWinButton(label, glyph, action string) *ui.Element {
	return chadcn.Button(glyph, chadcn.ButtonOptions{Variant: chadcn.ButtonText, Size: chadcn.ButtonSmall}).Attr("aria-label", label).Attr("title", label).Attr("data-action", action)
}

// invNavItem is one product-nav label. Only Inventory is active, so only it
// carries the active pill class and the no-op inventory action.
func invNavItem(label string, active bool) *ui.Element {
	class := "inv-nav-item"
	if active {
		class += " inv-nav-item--active"
	}
	return ui.Text(label).Class(class)
}

// pageHeader is the "Inventory health" title row: the warehouse subtitle on
// the left and the period segment plus action buttons on the right.
func (a *app) pageHeader() *ui.Element {
	return ui.Div().Class("inv-page-head").Children(
		ui.Div().Class("inv-page-titles").Children(
			chadcn.H1("Inventory health"),
			ui.P().Class("inv-page-sub").Text("Warehouse 1 · Reno, NV · synced 4 min ago from NetSuite"),
		),
		ui.Div().Class("inv-page-actions").Children(
			a.invPeriodSegment(),
			chadcn.Button("All categories", chadcn.ButtonOptions{
				Variant: chadcn.ButtonOutlined,
				Size:    chadcn.ButtonSmall,
			}).Attr("data-action", "category:All categories"),
			// Class replaces the attribute, so repeat the button classes
			// alongside inv-primary instead of appending to them.
			chadcn.Button("+ New PO", chadcn.ButtonOptions{
				Variant: chadcn.ButtonFilled,
				Size:    chadcn.ButtonSmall,
			}).Class("of-button of-button--filled of-button--small inv-primary").
				Attr("data-action", "po:new"),
		),
	)
}

// invPeriodSegment is the 30/90/12-month segmented control. The labels double
// as the a.period values so the host filter compares them directly.
func (a *app) invPeriodSegment() *ui.Element {
	periods := []string{"30 days", "90 days", "12 months"}
	buttons := make([]*ui.Element, 0, len(periods))
	for _, period := range periods {
		variant := chadcn.ButtonOutlined
		if a.invPeriodActive(period) {
			variant = chadcn.ButtonFilled
		}
		buttons = append(buttons, chadcn.Button(period, chadcn.ButtonOptions{
			Variant: variant,
			Size:    chadcn.ButtonSmall,
		}).OnClick(func(ctx context.Context) error {
			a.period = period
			return a.refresh(ctx)
		}))
	}
	return chadcn.ButtonGroup(chadcn.ButtonGroupOptions{Label: "Period"}, buttons...)
}

// invPeriodActive reports whether value is the selected period. An empty
// period defaults to the 30-day view.
func (a *app) invPeriodActive(value string) bool {
	if string(a.period) == "" {
		return value == "30 days"
	}
	return string(a.period) == value
}
