package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/components"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

func (a *app) screen() *ui.Element {
	// The canvas fills the window height, so the status strip stays at the
	// window bottom in fullscreen and the panes scroll inside it.
	return components.Screen(
		a.toolbar(),
		ui.Div().Display(ui.DisplayFlex).Flex(ui.Number(1)).
			Gap(ui.Px(14)).Padding(ui.Px(12)).
			MinHeight(ui.Px(0)).Position(ui.PositionRelative).
			Children(a.sidebar(), a.messageList(), a.reader(), a.settingsCard(), a.toastCard()),
		ui.Div().Padding(ui.Edges(ui.Px(9), ui.Px(18))).FontSize(ui.Px(11)).Color(theme.Muted()).
			BorderTop(theme.Border()).
			Children(ui.Text("JMAP-inspired demo  ·  Local sample data  ·  No account, network connection, or sent messages")),
	)
}

func (a *app) toolbar() *ui.Element {
	return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionRow).
		AlignItems(ui.AlignItemsCenter).JustifyContent(ui.JustifyContentSpaceBetween).
		Padding(ui.Edges(ui.Px(16), ui.Px(22))).BackgroundColor(theme.Surface()).
		BorderBottom(theme.Border()).
		Children(
			ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(12)).
				Children(
					ui.Div().BackgroundColor(theme.Primary()).Color(theme.OnPrimary()).
						Padding(ui.Edges(ui.Px(8), ui.Px(11))).BorderRadius(ui.Px(10)).
						Children(ui.Text("P").FontSize(ui.Px(19)).FontWeight(ui.FontWeightBold)),
					ui.Div().Children(
						ui.Text("Postmark").FontSize(ui.Px(20)).FontWeight(ui.FontWeightBold),
						ui.P().Text("Mail  /  JMAP preview").Margin(ui.Edges(ui.Px(2), ui.Px(0), ui.Px(0))).
							FontSize(ui.Px(11)).Color(theme.Muted()),
					),
				),
			ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).
				Height(ui.Px(40)).
				Children(
					components.SearchField("mail-search", "Search mail…", a.query).Width(ui.Px(240)).Height(ui.Px(36)),
					a.actionButton("Search", "search", false),
					ui.Text("Alex Morgan").FontWeight(ui.FontWeightBold).Color(theme.Text()),
					settingsButton(),
					a.actionButton(a.devToolsLabel(), "win-dev", false),
					a.actionButton("−", "win-min", false),
					a.actionButton("□", "win-max", false),
					a.actionButton("×", "win-close", false),
				),
		)
}

func (a *app) devToolsLabel() string {
	if a.page != nil && a.page.Window().DevTools() {
		return "Dev tools on"
	}
	return "Dev tools off"
}

func (a *app) sidebar() *ui.Element {
	folders := make([]*ui.Element, 0, len(a.mailboxes))
	for _, box := range a.mailboxes {
		color := theme.Muted()
		bg := ui.Transparent()
		if a.folder == box.ID {
			color = theme.Primary()
			bg = theme.Selected()
		}
		folders = append(folders, ui.Div().
			ID("folder-"+box.ID).Attr("data-action", "folder:"+box.ID).
			Display(ui.DisplayFlex).JustifyContent(ui.JustifyContentSpaceBetween).
			AlignItems(ui.AlignItemsCenter).Padding(ui.Edges(ui.Px(11), ui.Px(12))).BorderRadius(ui.Px(9)).
			BackgroundColor(bg).Color(color).
			Children(
				ui.Text(box.Name).FontWeight(ui.FontWeightBold),
				ui.Text(fmt.Sprintf("%d", a.count(box.ID))).FontSize(ui.Px(12)),
			))
	}
	return ui.Nav().Width(ui.Px(mailSidebarWidth)).FlexShrink(ui.Number(0)).
		BackgroundColor(theme.Surface()).Padding(ui.Edges(ui.Px(19), ui.Px(12))).
		BorderRadius(ui.Px(10)).Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).
		MinHeight(ui.Px(0)).
		Children(
			a.actionButton("+  Compose", "compose", true).
				Width(ui.Percent(100)).TextAlign(ui.TextAlignCenter).Padding(ui.Edges(ui.Px(13), ui.Px(12))).
				FlexShrink(ui.Number(0)),
			ui.P().Text("MAILBOXES").Margin(ui.Edges(ui.Px(25), ui.Px(10), ui.Px(12))).
				FontSize(ui.Px(11)).FontWeight(ui.FontWeightBold).Color(theme.Muted()).
				FlexShrink(ui.Number(0)),
			ui.Div().Flex(ui.Number(1)).MinHeight(ui.Px(0)).
				OverflowY(ui.OverflowYAuto).Children(folders...),
			ui.Div().Margin(ui.Edges(ui.Px(30), ui.Px(10), ui.Px(0))).Padding(ui.Px(13)).
				FlexShrink(ui.Number(0)).
				BackgroundColor(theme.Control()).BorderRadius(ui.Px(9)).
				Children(
					ui.Text("Mailbox status").FontWeight(ui.FontWeightBold).Color(theme.Text()),
					ui.P().Text(fmt.Sprintf("%d unread · %d messages", a.unreadCount(), len(a.messages))).
						Margin(ui.Edges(ui.Px(8), ui.Px(0), ui.Px(0))).FontSize(ui.Px(12)).Color(theme.Muted()),
				),
		)
}

func (a *app) messageList() *ui.Element {
	visible := a.visibleMessages()
	total := len(visible)
	count := a.listVisibleCount()
	start, end := a.listWindow(total, count)
	rows := []*ui.Element{}
	for _, message := range visible[start:end] {
		rows = append(rows, a.messageRow(message))
	}
	if len(rows) == 0 {
		rows = append(rows, ui.P().Text("No messages in this mailbox.").
			Padding(ui.Px(20)).Color(theme.Muted()))
	}
	label := a.folder
	for _, box := range a.mailboxes {
		if box.ID == a.folder {
			label = box.Name
			break
		}
	}
	viewport := ui.Div().ID("mail-list-viewport").Flex(ui.Number(1)).
		MinHeight(ui.Px(0)).Position(ui.PositionRelative).
		Overflow(ui.OverflowHidden).
		Children(rows...)
	viewport.Children(a.paneScrollbar("mail-list-scrollbar", "mail-list-thumb", "mail-list-viewport", "Messages", &a.listScroll, a.listViewH))
	return ui.Div().Width(ui.Px(mailListWidth)).FlexShrink(ui.Number(0)).
		BackgroundColor(theme.Control()).BorderRadius(ui.Px(10)).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).MinHeight(ui.Px(0)).
		Children(
			ui.Div().Padding(ui.Edges(ui.Px(20), ui.Px(19), ui.Px(14))).
				FlexShrink(ui.Number(0)).
				BorderBottom(theme.Border()).
				Children(
					ui.Text(label).FontSize(ui.Px(23)).FontWeight(ui.FontWeightBold),
					ui.P().Text(fmt.Sprintf("%d conversations", len(a.visibleMessages()))).
						Color(theme.Muted()).FontSize(ui.Px(12)).Margin(ui.Edges(ui.Px(7), ui.Px(0), ui.Px(0))),
					ui.Div().Margin(ui.Edges(ui.Px(12), ui.Px(0), ui.Px(0))).Children(
						a.actionButton("Clear search", "clear-search", false),
					),
				),
			viewport,
		)
}

func (a *app) messageRow(m *mail) *ui.Element {
	bg := theme.Surface()
	if m.ID == a.selectedID {
		bg = theme.Selected()
	}
	weight := ui.FontWeightNormal
	if !m.Keywords["$seen"] {
		weight = ui.FontWeightBold
	}
	star := ""
	if m.Keywords["$flagged"] {
		star = "  ★"
	}
	return ui.Div().ID("row-"+m.ID).Attr("data-action", "message:"+m.ID).
		Height(ui.Px(mailRowH)).BoxSizing(ui.BoxSizingBorderBox).
		Overflow(ui.OverflowHidden).
		Padding(ui.Edges(ui.Px(16), ui.Px(17))).BackgroundColor(bg).
		BorderBottom(theme.Border()).
		Children(
			ui.Div().Display(ui.DisplayFlex).JustifyContent(ui.JustifyContentSpaceBetween).
				Children(
					ui.Text(m.From.Name+star).FontWeight(weight).Color(theme.Text()).
						FontSize(ui.Px(13)),
					ui.Text(shortDate(m.ReceivedAt)).FontSize(ui.Px(11)).Color(theme.Muted()),
				),
			ui.P().Text(m.Subject).Margin(ui.Edges(ui.Px(9), ui.Px(0), ui.Px(5))).
				FontWeight(weight).Color(theme.Text()).FontSize(ui.Px(14)),
			ui.P().Text(m.Preview).Margin(ui.Px(0)).Color(theme.Muted()).
				FontSize(ui.Px(12)).LineHeight(ui.Number(1.5)),
		)
}

func (a *app) reader() *ui.Element {
	m := a.current()
	if m == nil || !a.isVisible(m) {
		return ui.Div().Flex(ui.Number(1)).Padding(ui.Px(28)).
			Children(
				ui.Text("No message selected").FontSize(ui.Px(22)).FontWeight(ui.FontWeightBold),
				ui.P().Text("Choose a message from the list to read it."),
			)
	}

	status := "Mark unread"
	if !m.Keywords["$seen"] {
		status = "Mark read"
	}
	star := "☆ Star"
	if m.Keywords["$flagged"] {
		star = "★ Starred"
	}
	actions := []*ui.Element{
		a.actionButton(star, "star", false),
		a.actionButton(status, "toggle-read", false),
		a.actionButton("Archive", "archive", false),
		a.actionButton("Delete", "delete", false),
	}
	if m.Keywords["$draft"] {
		actions = append([]*ui.Element{a.actionButton("Send", "send", true)}, actions...)
	}
	body := ui.Div().ID("mail-reader-content").Padding(ui.Edges(ui.Px(30), ui.Px(35))).
		Children(
			ui.Text(m.Subject).FontSize(ui.Px(26)).FontWeight(ui.FontWeightBold).
				Color(theme.Text()),
			ui.Div().Margin(ui.Edges(ui.Px(23), ui.Px(0))).Padding(ui.Edges(ui.Px(0), ui.Px(0), ui.Px(18))).
				BorderBottom(theme.Border()).
				Children(
					ui.Text(m.From.Name).FontWeight(ui.FontWeightBold).
						FontSize(ui.Px(15)).Color(theme.Text()),
					ui.P().Text(m.From.Email+"  →  "+m.To.Email).
						Margin(ui.Edges(ui.Px(7), ui.Px(0), ui.Px(0))).FontSize(ui.Px(12)).Color(theme.Muted()),
					ui.P().Text(displayDate(m.ReceivedAt)).
						Margin(ui.Edges(ui.Px(5), ui.Px(0), ui.Px(0))).FontSize(ui.Px(12)).Color(theme.Muted()),
				),
			ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(14)).
				Margin(ui.Px(0)).Children(a.bodyParagraphs(m)...),
		)
	if a.readerScroll.Offset() > 0 {
		body.Transform(ui.TranslateY(ui.Px(float64(-a.readerScroll.Offset()))))
	}
	viewport := ui.Div().ID("mail-reader-viewport").
		Flex(ui.Number(1)).MinHeight(ui.Px(0)).
		Position(ui.PositionRelative).Overflow(ui.OverflowHidden).
		Children(body)
	viewport.Children(a.paneScrollbar("mail-reader-scrollbar", "mail-reader-thumb", "mail-reader-viewport", "Message body", &a.readerScroll, a.readerScroll.Metrics().Viewport))
	return components.Panel().Flex(ui.Number(1)).MinWidth(ui.Px(0)).MinHeight(ui.Px(0)).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).
		Children(
			ui.Div().Padding(ui.Edges(ui.Px(15), ui.Px(22))).
				FlexShrink(ui.Number(0)).
				BorderBottom(theme.Border()).
				Display(ui.DisplayFlex).Gap(ui.Px(9)).
				Children(actions...),
			viewport,
			ui.Div().Padding(ui.Edges(ui.Px(15), ui.Px(35))).
				FlexShrink(ui.Number(0)).
				BorderTop(theme.Border()).
				BackgroundColor(theme.Control()).
				Children(
					ui.Text("Local JMAP preview").FontWeight(ui.FontWeightBold).
						Color(theme.Text()),
					ui.P().Text("Mailbox membership and $seen/$flagged keywords update in memory only. No messages are sent or deleted on a server.").
						FontSize(ui.Px(12)).Color(theme.Muted()).Margin(ui.Edges(ui.Px(6), ui.Px(0), ui.Px(0))),
				),
		)
}

func (a *app) bodyParagraphs(m *mail) []*ui.Element {
	parts := strings.Split(m.Body, "\n\n")
	paras := make([]*ui.Element, 0, len(parts))
	for _, part := range parts {
		text := strings.TrimSpace(strings.ReplaceAll(part, "\n", " "))
		if text == "" {
			continue
		}
		paras = append(paras, ui.P().Text(text).Margin(ui.Px(0)).
			FontSize(ui.Px(15)).Color(theme.Text()).
			LineHeight(ui.Number(1.7)))
	}
	return paras
}

func (a *app) settingsCard() *ui.Element {
	if !a.settings {
		return nil
	}
	return components.Dialog(components.DialogOptions{
		BackdropID: "settings-card", PanelID: "settings-panel", Label: "Settings", CloseAction: "settings", Width: ui.Px(520),
	},
		ui.Div().Display(ui.DisplayFlex).JustifyContent(ui.JustifyContentSpaceBetween).
			AlignItems(ui.AlignItemsCenter).Padding(ui.Edges(ui.Px(16), ui.Px(20))).
			BorderBottom(theme.Border()).
			Children(
				ui.Text("Settings").FontSize(ui.Px(16)).FontWeight(ui.FontWeightBold),
				a.actionButton("×", "settings", false).ID("settings-close"),
			),
		ui.Div().Padding(ui.Edges(ui.Px(18), ui.Px(20), ui.Px(20))).
			Children(
				ui.Text("Appearance").FontWeight(ui.FontWeightBold),
				ui.P().Text("Choose how the mail preview looks.").Margin(ui.Edges(ui.Px(5), ui.Px(0), ui.Px(12))).
					FontSize(ui.Px(12)).Color(theme.Muted()),
				ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).
					Children(
						a.themeOption("theme-light", "Light", "Bright panels for daytime reading.", "theme:light", a.appearance == mailThemeLight),
						a.themeOption("theme-dark", "Dark", "Dim panels for low light.", "theme:dark", a.appearance == mailThemeDark),
						a.themeOption("theme-custom", "Custom · Sage", "App-defined colors with soft sage surfaces.", "theme:custom", a.appearance == mailThemeCustom),
					),
				ui.P().Text("Your preference applies to this local preview only.").Margin(ui.Edges(ui.Px(12), ui.Px(0), ui.Px(0))).
					FontSize(ui.Px(12)).Color(theme.Muted()),
			),
	)
}

func (a *app) toastCard() *ui.Element {
	if a.toast == "" {
		return nil
	}
	return ui.Div().ID("mail-toast").
		Position(ui.PositionAbsolute).Left(ui.Px(16)).Bottom(ui.Px(16)).
		ZIndex(ui.Number(60)).Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(12)).
		Width(ui.Px(330)).BoxSizing(ui.BoxSizingBorderBox).
		Padding(ui.Edges(ui.Px(13), ui.Px(14))).BorderRadius(ui.Px(10)).
		BackgroundColor(theme.Control()).
		Color(theme.Text()).
		Border(theme.Border()).
		Children(
			ui.Text("✓").FontWeight(ui.FontWeightBold),
			ui.Div().Flex(ui.Number(1)).MinWidth(ui.Px(0)).
				Children(ui.Text(a.toast).FontSize(ui.Px(13))),
			a.actionButton("×", "toast-close", false).ID("toast-close"),
		)
}

func (a *app) themeOption(id, title, caption, action string, selected bool) *ui.Element {
	return components.Choice(title, caption, selected).ID(id).Attr("data-action", action)
}

func shortDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("02 Jan")
}

func displayDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("02 Jan 2006 · 15:04 MST")
}
