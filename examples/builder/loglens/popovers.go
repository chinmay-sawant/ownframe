package main

import (
	"context"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// topPopovers is the topbar cluster: live status, the on-call card, the
// notification inbox, the account menu, the command trigger, and the theme
// switch. Hover state for the tooltip and the hover card arrives from
// handleHover, which matches the "tooltip-trigger" and "hover-trigger" IDs.
func (a *app) topPopovers() *ui.Element {
	themeLabel := "Dark"
	if a.dark {
		themeLabel = "Light"
	}
	// A closed tooltip must not age its hidden label. The patch path cannot
	// splice an element that has no box, and the tick updates the label while
	// the tooltip is open.
	tooltipLabel := "Updated just now"
	if a.tooltipOpen {
		tooltipLabel = a.liveUpdatedLabel(time.Now())
	}

	unreadDot := ui.Text("●").
		Attr("aria-hidden", "true").
		FontSize(ui.Px(8)).
		Color(ui.ColorVar("--of-danger", ui.Hex("#dc2626")))

	notifications := chadcn.Button("Inbox", chadcn.ButtonOptions{
		Variant: chadcn.ButtonText,
		Size:    chadcn.ButtonSmall,
	}).Children(unreadDot)

	return ui.Div().Class("app-row app-top-popovers").
		Gap(ui.Px(3)).
		FlexWrap(ui.FlexWrapNowrap).
		FlexShrink(ui.Number(0)).
		Children(
			chadcn.Tooltip(chadcn.TooltipOptions{
				ID:       "live-tooltip",
				Open:     a.tooltipOpen,
				CSSHover: true,
				Content:  tooltipLabel,
				Side:     chadcn.PopupBottom,
			}, chadcn.Button("Live", chadcn.ButtonOptions{
				Variant: chadcn.ButtonText,
				Size:    chadcn.ButtonSmall,
				Icon:    ui.Text("●").Color(ui.Hex("#22c55e")),
			}).ID("tooltip-trigger")).Width(ui.Px(72)).FlexShrink(ui.Number(0)),

			chadcn.HoverCard(chadcn.HoverCardOptions{
				ID:       "oncall-hover",
				Open:     a.hoverOpen,
				CSSHover: true,
				Side:     chadcn.PopupBottom,
			}, chadcn.Button("On call", chadcn.ButtonOptions{
				Variant: chadcn.ButtonText,
				Size:    chadcn.ButtonSmall,
			}).ID("hover-trigger"),
				ui.Div().Class("app-col").Children(
					ui.Div().Class("app-row").Children(
						chadcn.Avatar(chadcn.AvatarOptions{
							Initials: "AK",
							Size:     chadcn.AvatarSmall,
							Status:   chadcn.AvatarStatusOnline,
						}),
						ui.Div().Class("app-col").Gap(ui.Px(2)).Children(
							ui.Div().Class("app-row").Children(
								ui.Text("Ada Kim").FontSize(ui.Px(13)).FontWeight(ui.FontWeightBold),
								chadcn.Badge("On call", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
							),
							ui.Text("Platform lead · Identity").Class("app-muted"),
						),
					),
					chadcn.Separator(chadcn.SeparatorOptions{}),
					ui.Text("Responding to INC-2481 since 14:12").Class("app-muted"),
					ui.Div().Class("app-row app-row--between").Children(
						ui.Text("Ask Ada to join the room").Class("app-muted"),
						chadcn.Kbd("⌘⇧A"),
					),
				),
			).Width(ui.Px(68)).FlexShrink(ui.Number(0)),

			chadcn.Popover(chadcn.PopoverOptions{
				ID:      "notifications-popover",
				Open:    a.notifOpen,
				Side:    chadcn.PopupBottom,
				Align:   chadcn.PopoverAlignEnd,
				Trigger: notifications,
				OnOpenChange: func(ctx context.Context, open bool) error {
					if open {
						a.closeHeaderPopovers()
						a.navOpen = ""
						a.menubarOpen = ""
						a.userMenuOpen = false
					}
					a.notifOpen = open
					return a.refresh(ctx)
				},
			},
				ui.Div().Class("app-row app-row--between").Children(
					ui.Text("Notifications").FontSize(ui.Px(13)).FontWeight(ui.FontWeightBold),
					chadcn.Badge("3 new", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
				),
				chadcn.Separator(chadcn.SeparatorOptions{}),
				topNotifRow("AK", "Ada assigned INC-2481", "api-gateway · 2m ago", "error"),
				topNotifRow("W", "Worker queue at 1.8k", "autoscale paused · 6m ago", "warn"),
				topNotifRow("D", "Deploy web@2.14.0", "eu-west rolled out · 12m ago", "info"),
				chadcn.Separator(chadcn.SeparatorOptions{}),
				ui.Div().Class("app-row app-row--between").Children(
					ui.Text("Retention 30 days").Class("app-muted"),
					chadcn.Button("Open inbox", chadcn.ButtonOptions{
						Variant: chadcn.ButtonText,
						Size:    chadcn.ButtonSmall,
					}).OnClick(func(ctx context.Context) error {
						a.notifOpen = false
						return a.showToast(ctx, "Inbox", "Every unread alert is waiting in the incident room.")
					}),
				),
			).Width(ui.Px(64)).FlexShrink(ui.Number(0)),

			a.accountMenu(),

			chadcn.Separator(chadcn.SeparatorOptions{Orientation: chadcn.SeparatorVertical}),

			chadcn.Button("⌘K", chadcn.ButtonOptions{
				Variant: chadcn.ButtonOutlined,
				Size:    chadcn.ButtonSmall,
			}).OnClick(func(ctx context.Context) error {
				a.commandOpen = true
				a.navOpen = ""
				a.menubarOpen = ""
				if a.commandActive == "" {
					a.commandActive = "overview"
				}
				return a.refresh(ctx)
			}),

			chadcn.Button(themeLabel, chadcn.ButtonOptions{
				Variant: chadcn.ButtonText,
				Size:    chadcn.ButtonSmall,
			}).Width(ui.Px(54)).OnClick(a.toggleTheme),
		)
}

// accountMenu is the avatar trigger for the session menu. The initials and
// status ring carry the identity, so the top bar does not spend width on the
// word "Account".
func (a *app) accountMenu() *ui.Element {
	trigger := chadcn.Button("", chadcn.ButtonOptions{
		Variant: chadcn.ButtonText,
	}).ID("profile-trigger").Attr("aria-label", "Account").Attr("title", "Account").
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
			if open {
				a.closeHeaderPopovers()
				a.navOpen = ""
				a.menubarOpen = ""
				a.notifOpen = false
			}
			a.userMenuOpen = open
			return a.refresh(ctx)
		},
	})
	// Keep the account slot fixed while its menu opens.
	menu.Width(ui.Px(48)).Height(ui.Px(36)).FlexShrink(ui.Number(0))
	return menu
}

// topNotifRow is one inbox row: the actor, the event and its age, and the
// level badge that mirrors the log table.
func topNotifRow(initials, title, meta, level string) *ui.Element {
	return chadcn.Item(chadcn.ItemOptions{
		Media: chadcn.Avatar(chadcn.AvatarOptions{
			Initials: initials,
			Size:     chadcn.AvatarSmall,
		}),
		Content: ui.Div().Class("app-col").Gap(ui.Px(2)).Children(
			ui.Text(title).FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold),
			ui.Text(meta).FontSize(ui.Px(11)).Class("app-muted"),
		),
		Actions: chadcn.Badge(level, chadcn.BadgeOptions{Variant: logLevelVariant(level)}),
	})
}

// commandOverlay is the ⌘K palette. It renders nothing while closed. The
// backdrop closes the palette; the panel carries an ID but no action, so a
// click inside it stops at the panel instead of reaching the backdrop.
func (a *app) commandOverlay() *ui.Element {
	if !a.commandOpen {
		return nil
	}

	dismiss := func(ctx context.Context) error {
		a.commandOpen = false
		a.commandActive = ""
		return a.refresh(ctx)
	}

	groups := []chadcn.CommandGroup{
		{Label: "Navigate", Items: []chadcn.CommandItem{
			{Value: "overview", Label: "Overview", Shortcut: "⌘1", Icon: ui.Text("◈")},
			{Value: "logs", Label: "Logs", Shortcut: "⌘2", Icon: ui.Text("≡")},
			{Value: "traces", Label: "Traces", Shortcut: "⌘3", Icon: ui.Text("⌁")},
			{Value: "incident", Label: "Incident room", Shortcut: "⌘4", Icon: ui.Text("✳")},
		}},
		{Label: "Actions", Items: []chadcn.CommandItem{
			{Value: "alert:new", Label: "New alert rule", Shortcut: "⇧A", Icon: ui.Text("＋")},
			{Value: "settings:open", Label: "Open settings", Shortcut: "⌘,", Icon: ui.Text("⚙")},
			{Value: "theme:toggle", Label: "Toggle light and dark", Shortcut: "⌘⇧L", Icon: ui.Text("◐")},
			{Value: "refresh", Label: "Refresh dashboard", Shortcut: "R", Icon: ui.Text("↻")},
		}},
		{Label: "Services", Items: serviceCommandItems()},
	}

	overlay := ui.Div().Class("app-overlay").OnClick(func(ctx context.Context) error {
		return dismiss(ctx)
	}).Children(
		ui.Div().ID("command-panel").Class("app-command").Children(
			chadcn.Command(chadcn.CommandOptions{
				ID:     "command",
				Query:  a.commandQuery,
				Active: a.commandActive,
				Groups: filterCommandGroups(groups, a.commandQuery),
				OnSelect: func(ctx context.Context, value string) error {
					switch value {
					case "overview", "logs", "traces", "incident":
						a.tab = value
						return dismiss(ctx)
					case "alert:new":
						a.dialogOpen = true
						return dismiss(ctx)
					case "settings:open":
						a.settingsOpen = true
						return dismiss(ctx)
					case "theme:toggle":
						a.commandOpen = false
						a.commandActive = ""
						return a.toggleTheme(ctx)
					case "refresh":
						a.commandOpen = false
						a.commandActive = ""
						return a.refreshTelemetry(ctx)
					}
					if service, ok := strings.CutPrefix(value, "service:"); ok {
						a.service = service
						a.pageIndex = 0
						a.tab = "logs"
					}
					return dismiss(ctx)
				},
			}),
		),
	)
	return ui.Div().Children(overlay)
}

// serviceCommandItems turns the service list into palette rows that jump to
// that service's logs.
func serviceCommandItems() []chadcn.CommandItem {
	items := make([]chadcn.CommandItem, 0, len(serviceOptions))
	for _, service := range serviceOptions {
		if service == "All services" {
			continue
		}
		items = append(items, chadcn.CommandItem{
			Value:    "service:" + service,
			Label:    "Filter logs to " + service,
			Shortcut: "↵",
			Icon:     ui.Text("◉"),
		})
	}
	return items
}

// filterCommandGroups narrows the palette to rows matching the query. The
// groups themselves stay grouped, so the section headers follow the matches.
func filterCommandGroups(groups []chadcn.CommandGroup, query string) []chadcn.CommandGroup {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return groups
	}
	filtered := make([]chadcn.CommandGroup, 0, len(groups))
	for _, group := range groups {
		items := make([]chadcn.CommandItem, 0, len(group.Items))
		for _, item := range group.Items {
			if strings.Contains(strings.ToLower(item.Label+" "+item.Value), q) {
				items = append(items, item)
			}
		}
		if len(items) > 0 {
			filtered = append(filtered, chadcn.CommandGroup{Label: group.Label, Items: items})
		}
	}
	return filtered
}
