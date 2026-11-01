package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe/ui"
)

func (a *app) screen() *ui.Element {
	// The canvas fills the window height, so the status strip stays at the
	// window bottom in fullscreen and the panes scroll inside it.
	return ui.Div().Width(ui.Percent(100)).Height(ui.RawValue("100vh")).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).
		BackgroundColor(ui.RawValue(a.palette("#f5f7fb", "#111827"))).Color(ui.RawValue(a.palette("#172b4d", "#e4ecf8"))).
		FontFamily(ui.RawValue("sans-serif")).FontSize(ui.Px(14)).
		Children(
			a.toolbar(),
			ui.Div().Display(ui.DisplayFlex).Flex(ui.RawValue("1")).
				Gap(ui.Px(14)).Padding(ui.Px(12)).
				MinHeight(ui.Px(0)).Position(ui.PositionRelative).
				Children(a.sidebar(), a.messageList(), a.reader(), a.settingsCard(), a.toastCard()),
			ui.Div().Padding(ui.Edges(ui.Px(9), ui.Px(18))).FontSize(ui.Px(11)).Color(ui.RawValue(a.palette("#73829c", "#9aabc3"))).
				BorderTop(ui.RawValue(a.palette("1px solid #dce4f0", "1px solid #39465a"))).
				Children(ui.Text("JMAP-inspired demo  ·  Local sample data  ·  No account, network connection, or sent messages")),
		)
}

func (a *app) toolbar() *ui.Element {
	return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionRow).
		AlignItems(ui.AlignItemsCenter).JustifyContent(ui.JustifyContentSpaceBetween).
		Padding(ui.Edges(ui.Px(16), ui.Px(22))).BackgroundColor(ui.RawValue(a.palette("#ffffff", "#1c2738"))).
		BorderBottom(ui.RawValue(a.palette("1px solid #dce4f0", "1px solid #39465a"))).
		Children(
			ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(12)).
				Children(
					ui.Div().BackgroundColor(ui.RawValue(a.palette("#3366d6", "#4e7ff0"))).Color(ui.RawValue(a.palette("#ffffff", "#1c2738"))).
						Padding(ui.Edges(ui.Px(8), ui.Px(11))).BorderRadius(ui.Px(10)).
						Children(ui.Text("P").FontSize(ui.Px(19)).FontWeight(ui.FontWeightBold)),
					ui.Div().Children(
						ui.Text("Postmark").FontSize(ui.Px(20)).FontWeight(ui.FontWeightBold),
						ui.P().Text("Mail  /  JMAP preview").Margin(ui.RawValue("2px 0 0")).
							FontSize(ui.Px(11)).Color(ui.RawValue(a.palette("#7b8ba5", "#9cacc2"))),
					),
				),
			ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).
				Height(ui.Px(40)).
				Children(
					ui.Input().ID("mail-search").Attr("type", "search").
						Attr("placeholder", "Search mail…").Attr("value", a.query).
						Width(ui.Px(240)).Height(ui.Px(36)).
						BoxSizing(ui.BoxSizingBorderBox).Padding(ui.Edges(ui.Px(7), ui.Px(10))).
						Border(ui.RawValue(a.palette("1px solid #dce4f0", "1px solid #39465a"))).BorderRadius(ui.Px(9)).
						BackgroundColor(ui.RawValue(a.palette("#f6f8fc", "#27354a"))),
					a.actionButton("Search", "search", false),
					ui.Text("Alex Morgan").FontWeight(ui.FontWeightBold).Color(ui.RawValue(a.palette("#485a77", "#d1deef"))),
					a.actionButton("⚙ Settings", "settings", false),
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
		color := a.palette("#60718c", "#a2b1c6")
		bg := "transparent"
		if a.folder == box.ID {
			color = a.palette("#234fa9", "#b4ccff")
			bg = a.palette("#e5edff", "#2a4164")
		}
		folders = append(folders, ui.Div().
			ID("folder-"+box.ID).Attr("data-action", "folder:"+box.ID).
			Display(ui.DisplayFlex).JustifyContent(ui.JustifyContentSpaceBetween).
			AlignItems(ui.AlignItemsCenter).Padding(ui.Edges(ui.Px(11), ui.Px(12))).BorderRadius(ui.Px(9)).
			BackgroundColor(ui.RawValue(bg)).Color(ui.RawValue(color)).
			Children(
				ui.Text(box.Name).FontWeight(ui.FontWeightBold),
				ui.Text(fmt.Sprintf("%d", a.count(box.ID))).FontSize(ui.Px(12)),
			))
	}
	return ui.Nav().Width(ui.RawValue(mailSidebarWidth)).FlexShrink(ui.Number(0)).
		BackgroundColor(ui.RawValue(a.palette("#ffffff", "#1c2738"))).Padding(ui.Edges(ui.Px(19), ui.Px(12))).
		BorderRadius(ui.Px(10)).Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).
		MinHeight(ui.Px(0)).
		Children(
			a.actionButton("+  Compose", "compose", true).
				Width(ui.Percent(100)).TextAlign(ui.TextAlignCenter).Padding(ui.Edges(ui.Px(13), ui.Px(12))).
				FlexShrink(ui.Number(0)),
			ui.P().Text("MAILBOXES").Margin(ui.Edges(ui.Px(25), ui.Px(10), ui.Px(12))).
				FontSize(ui.Px(11)).FontWeight(ui.FontWeightBold).Color(ui.RawValue(a.palette("#96a3b6", "#a4b3c6"))).
				FlexShrink(ui.Number(0)),
			ui.Div().Flex(ui.RawValue("1")).MinHeight(ui.Px(0)).
				OverflowY(ui.OverflowYAuto).Children(folders...),
			ui.Div().Margin(ui.RawValue("30px 10px 0")).Padding(ui.Px(13)).
				FlexShrink(ui.Number(0)).
				BackgroundColor(ui.RawValue(a.palette("#f3f6fc", "#27354a"))).BorderRadius(ui.Px(9)).
				Children(
					ui.Text("Mailbox status").FontWeight(ui.FontWeightBold).Color(ui.RawValue(a.palette("#415372", "#c7d4e8"))),
					ui.P().Text(fmt.Sprintf("%d unread · %d messages", a.unreadCount(), len(a.messages))).
						Margin(ui.RawValue("8px 0 0")).FontSize(ui.Px(12)).Color(ui.RawValue(a.palette("#7688a0", "#a1b3cb"))),
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
			Padding(ui.Px(20)).Color(ui.RawValue(a.palette("#8391a8", "#9aabc3"))))
	}
	label := a.folder
	for _, box := range a.mailboxes {
		if box.ID == a.folder {
			label = box.Name
			break
		}
	}
	viewport := ui.Div().ID("mail-list-viewport").Flex(ui.RawValue("1")).
		MinHeight(ui.Px(0)).Position(ui.PositionRelative).
		Overflow(ui.OverflowHidden).
		Children(rows...)
	if top, height, show := a.listThumb(total, count); show {
		viewport.Children(a.scrollThumb("mail-list-thumb", top, height))
	}
	return ui.Div().Width(ui.RawValue(mailListWidth)).FlexShrink(ui.Number(0)).
		BackgroundColor(ui.RawValue(a.palette("#fbfcff", "#202c3d"))).BorderRadius(ui.Px(10)).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).MinHeight(ui.Px(0)).
		Children(
			ui.Div().Padding(ui.Edges(ui.Px(20), ui.Px(19), ui.Px(14))).
				FlexShrink(ui.Number(0)).
				BorderBottom(ui.RawValue(a.palette("1px solid #e9edf4", "1px solid #39465a"))).
				Children(
					ui.Text(label).FontSize(ui.Px(23)).FontWeight(ui.FontWeightBold),
					ui.P().Text(fmt.Sprintf("%d conversations", len(a.visibleMessages()))).
						Color(ui.RawValue(a.palette("#8797ae", "#a5b4ca"))).FontSize(ui.Px(12)).Margin(ui.RawValue("7px 0 0")),
					ui.Div().Margin(ui.RawValue("12px 0 0")).Children(
						a.actionButton("Clear search", "clear-search", false),
					),
				),
			viewport,
		)
}

// scrollThumb is the overlay thumb for a scrollable pane. It renders only
// when the caller measured overflow, so a fitting list shows no track.
func (a *app) scrollThumb(id string, top, height int) *ui.Element {
	return ui.Div().ID(id).
		Position(ui.PositionAbsolute).Top(ui.RawValue(fmt.Sprintf("%dpx", top))).
		Right(ui.Px(2)).Height(ui.RawValue(fmt.Sprintf("%dpx", height))).Width(ui.Px(8)).
		BorderRadius(ui.Px(4)).BackgroundColor(ui.RawValue(a.palette("#c3cedd", "#5b6b82")))
}

func (a *app) messageRow(m *mail) *ui.Element {
	bg := a.palette("#ffffff", "#1c2738")
	if m.ID == a.selectedID {
		bg = a.palette("#e9f0ff", "#304867")
	}
	weight := "normal"
	if !m.Keywords["$seen"] {
		weight = "bold"
	}
	star := ""
	if m.Keywords["$flagged"] {
		star = "  ★"
	}
	return ui.Div().ID("row-"+m.ID).Attr("data-action", "message:"+m.ID).
		Height(ui.RawValue(fmt.Sprintf("%dpx", mailRowH))).BoxSizing(ui.BoxSizingBorderBox).
		Overflow(ui.OverflowHidden).
		Padding(ui.Edges(ui.Px(16), ui.Px(17))).BackgroundColor(ui.RawValue(bg)).
		BorderBottom(ui.RawValue(a.palette("1px solid #e7ecf4", "1px solid #39465a"))).
		Children(
			ui.Div().Display(ui.DisplayFlex).JustifyContent(ui.JustifyContentSpaceBetween).
				Children(
					ui.Text(m.From.Name+star).FontWeight(ui.RawValue(weight)).Color(ui.RawValue(a.palette("#263954", "#e5edf9"))).
						FontSize(ui.Px(13)),
					ui.Text(shortDate(m.ReceivedAt)).FontSize(ui.Px(11)).Color(ui.RawValue(a.palette("#8b9bb1", "#9aabc3"))),
				),
			ui.P().Text(m.Subject).Margin(ui.RawValue("9px 0 5px")).
				FontWeight(ui.RawValue(weight)).Color(ui.RawValue(a.palette("#263954", "#e5edf9"))).FontSize(ui.Px(14)),
			ui.P().Text(m.Preview).Margin(ui.Px(0)).Color(ui.RawValue(a.palette("#7a8da6", "#b0bed2"))).
				FontSize(ui.Px(12)).LineHeight(ui.Number(1.5)),
		)
}

func (a *app) reader() *ui.Element {
	m := a.current()
	if m == nil || !a.isVisible(m) {
		return ui.Div().Flex(ui.RawValue("1")).Padding(ui.Px(28)).
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
				Color(ui.RawValue(a.palette("#162944", "#f1f5ff"))),
			ui.Div().Margin(ui.RawValue("23px 0")).Padding(ui.RawValue("0 0 18px")).
				BorderBottom(ui.RawValue(a.palette("1px solid #e8edf5", "1px solid #39465a"))).
				Children(
					ui.Text(m.From.Name).FontWeight(ui.FontWeightBold).
						FontSize(ui.Px(15)).Color(ui.RawValue(a.palette("#314560", "#d5e2f4"))),
					ui.P().Text(m.From.Email+"  →  "+m.To.Email).
						Margin(ui.RawValue("7px 0 0")).FontSize(ui.Px(12)).Color(ui.RawValue(a.palette("#7e91a8", "#a1b3cb"))),
					ui.P().Text(displayDate(m.ReceivedAt)).
						Margin(ui.RawValue("5px 0 0")).FontSize(ui.Px(12)).Color(ui.RawValue(a.palette("#7e91a8", "#a1b3cb"))),
				),
			ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(14)).
				Margin(ui.Px(0)).Children(a.bodyParagraphs(m)...),
		)
	if a.readerScroll > 0 {
		body.Transform(ui.RawValue(fmt.Sprintf("translateY(-%dpx)", a.readerScroll)))
	}
	viewport := ui.Div().ID("mail-reader-viewport").
		Flex(ui.RawValue("1")).MinHeight(ui.Px(0)).
		Position(ui.PositionRelative).Overflow(ui.OverflowHidden).
		Children(body)
	if top, height, show := a.readerThumb(); show {
		viewport.Children(a.scrollThumb("mail-reader-thumb", top, height))
	}
	return ui.Div().Flex(ui.RawValue("1")).MinWidth(ui.Px(0)).MinHeight(ui.Px(0)).
		BackgroundColor(ui.RawValue(a.palette("#ffffff", "#1c2738"))).BorderRadius(ui.Px(10)).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).
		Children(
			ui.Div().Padding(ui.Edges(ui.Px(15), ui.Px(22))).
				FlexShrink(ui.Number(0)).
				BorderBottom(ui.RawValue(a.palette("1px solid #e9edf4", "1px solid #39465a"))).
				Display(ui.DisplayFlex).Gap(ui.Px(9)).
				Children(actions...),
			viewport,
			ui.Div().Padding(ui.Edges(ui.Px(15), ui.Px(35))).
				FlexShrink(ui.Number(0)).
				BorderTop(ui.RawValue(a.palette("1px solid #e8edf5", "1px solid #39465a"))).
				BackgroundColor(ui.RawValue(a.palette("#f4f7fe", "#27354a"))).
				Children(
					ui.Text("Local JMAP preview").FontWeight(ui.FontWeightBold).
						Color(ui.RawValue(a.palette("#4a6496", "#b6ceff"))),
					ui.P().Text("Mailbox membership and $seen/$flagged keywords update in memory only. No messages are sent or deleted on a server.").
						FontSize(ui.Px(12)).Color(ui.RawValue(a.palette("#7b8fa8", "#a4b6cf"))).Margin(ui.RawValue("6px 0 0")),
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
			FontSize(ui.Px(15)).Color(ui.RawValue(a.palette("#40526e", "#d1def0"))).
			LineHeight(ui.Number(1.7)))
	}
	return paras
}

func (a *app) palette(light, dark string) string {
	if a.dark {
		return dark
	}
	return light
}

func (a *app) settingsCard() *ui.Element {
	if !a.settings {
		return nil
	}
	return ui.Div().ID("settings-card").Attr("data-action", "settings").
		Position(ui.PositionAbsolute).Top(ui.Px(0)).Left(ui.Px(0)).
		Right(ui.Px(0)).Bottom(ui.Px(0)).ZIndex(ui.Number(50)).
		Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).JustifyContent(ui.JustifyContentCenter).
		BackgroundColor(ui.RawValue(a.palette("rgba(23, 43, 77, 0.35)", "rgba(0, 0, 0, 0.55)"))).
		Children(
			ui.Div().ID("settings-panel").
				Width(ui.Px(520)).
				BackgroundColor(ui.RawValue(a.palette("#ffffff", "#1c2738"))).
				Border(ui.RawValue(a.palette("1px solid #dce4f0", "1px solid #39465a"))).
				BorderRadius(ui.Px(12)).Overflow(ui.OverflowHidden).
				Children(
					ui.Div().Display(ui.DisplayFlex).JustifyContent(ui.JustifyContentSpaceBetween).
						AlignItems(ui.AlignItemsCenter).Padding(ui.Edges(ui.Px(16), ui.Px(20))).
						BorderBottom(ui.RawValue(a.palette("1px solid #dce4f0", "1px solid #39465a"))).
						Children(
							ui.Text("Settings").FontSize(ui.Px(16)).FontWeight(ui.FontWeightBold),
							a.actionButton("×", "settings", false).ID("settings-close"),
						),
					ui.Div().Padding(ui.Edges(ui.Px(18), ui.Px(20), ui.Px(20))).
						Children(
							ui.Text("Appearance").FontWeight(ui.FontWeightBold),
							ui.P().Text("Choose how the mail preview looks.").Margin(ui.RawValue("5px 0 12px")).
								FontSize(ui.Px(12)).Color(ui.RawValue(a.palette("#7b8fa5", "#9cacc2"))),
							ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).
								Children(
									a.themeOption("theme-light", "Light", "Bright panels for daytime reading.", "theme:light", !a.dark),
									a.themeOption("theme-dark", "Dark", "Dim panels for low light.", "theme:dark", a.dark),
								),
							ui.P().Text("Your preference applies to this local preview only.").Margin(ui.RawValue("12px 0 0")).
								FontSize(ui.Px(12)).Color(ui.RawValue(a.palette("#73829c", "#9aabc3"))),
						),
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
		BackgroundColor(ui.RawValue(a.palette("#162944", "#e5edff"))).
		Color(ui.RawValue(a.palette("#eaf1ff", "#162944"))).
		Border(ui.RawValue(a.palette("1px solid #3366d6", "1px solid #4e7ff0"))).
		Children(
			ui.Text("✓").FontWeight(ui.FontWeightBold),
			ui.Div().Flex(ui.RawValue("1")).MinWidth(ui.Px(0)).
				Children(ui.Text(a.toast).FontSize(ui.Px(13))),
			a.actionButton("×", "toast-close", false).ID("toast-close"),
		)
}

func (a *app) themeOption(id, title, caption, action string, selected bool) *ui.Element {
	bg := a.palette("#f6f8fc", "#27354a")
	border := a.palette("1px solid #dce4f0", "1px solid #39465a")
	check := ""
	if selected {
		bg = a.palette("#e5edff", "#2a4164")
		border = a.palette("1px solid #3366d6", "1px solid #4e7ff0")
		check = "✓"
	}
	return ui.Button("").ID(id).Attr("data-action", action).
		Width(ui.Percent(100)).TextAlign(ui.TextAlignLeft).
		BackgroundColor(ui.RawValue(bg)).Color(ui.RawValue(a.palette("#172b4d", "#e4ecf8"))).
		Border(ui.RawValue(border)).BoxSizing(ui.BoxSizingBorderBox).
		Padding(ui.Edges(ui.Px(12), ui.Px(14))).BorderRadius(ui.Px(9)).
		Children(
			ui.Div().Display(ui.DisplayFlex).JustifyContent(ui.JustifyContentSpaceBetween).
				AlignItems(ui.AlignItemsCenter).
				Children(
					ui.Div().Children(
						ui.Text(title).FontWeight(ui.FontWeightBold).FontSize(ui.Px(14)),
						ui.P().Text(caption).Margin(ui.RawValue("4px 0 0")).
							FontSize(ui.Px(12)).Color(ui.RawValue(a.palette("#73829c", "#9aabc3"))),
					),
					ui.Text(check).FontWeight(ui.FontWeightBold).Color(ui.RawValue(a.palette("#3366d6", "#4e7ff0"))),
				),
		)
}

func (a *app) actionButton(label, action string, prominent bool) *ui.Element {
	background := a.palette("#ffffff", "#1c2738")
	foreground := a.palette("#516684", "#d5e2f4")
	border := a.palette("1px solid #dce4f0", "1px solid #39465a")
	if prominent {
		background = a.palette("#3366d6", "#4e7ff0")
		foreground = a.palette("#ffffff", "#1c2738")
		border = a.palette("1px solid #3366d6", "1px solid #4e7ff0")
	}
	return ui.Button(label).Attr("data-action", action).
		BackgroundColor(ui.RawValue(background)).Color(ui.RawValue(foreground)).
		Border(ui.RawValue(border)).Padding(ui.Edges(ui.Px(9), ui.Px(13))).
		BoxSizing(ui.BoxSizingBorderBox).
		BorderRadius(ui.Px(8)).FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold)
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
