package main

import (
	"fmt"
	"time"

	"github.com/chinmay-sawant/ownframe/ui"
)

func (a *app) screen() *ui.Element {
	return ui.Div().Width("100%").Height(mailCanvasHeight).
		Display("flex").FlexDirection("column").
		BackgroundColor(a.palette("#f5f7fb", "#111827")).Color(a.palette("#172b4d", "#e4ecf8")).
		CSS("font-family", "sans-serif").FontSize("14px").
		Children(
			a.toolbar(),
			a.settingsPanel(),
			ui.Div().Display("flex").CSS("flex", "1").
				Gap("14px").Padding("12px").
				CSS("min-height", "0px").
				Children(a.sidebar(), a.messageList(), a.reader()),
			ui.Div().Padding("9px 18px").FontSize("11px").Color(a.palette("#73829c", "#9aabc3")).
				CSS("border-top", a.palette("1px solid #dce4f0", "1px solid #39465a")).
				Children(ui.Text("JMAP-inspired demo  ·  Local sample data  ·  No account, network connection, or sent messages")),
		)
}

func (a *app) toolbar() *ui.Element {
	return ui.Div().Display("flex").FlexDirection("row").
		AlignItems("center").CSS("justify-content", "space-between").
		Padding("16px 22px").BackgroundColor(a.palette("#ffffff", "#1c2738")).
		CSS("border-bottom", a.palette("1px solid #dce4f0", "1px solid #39465a")).
		Children(
			ui.Div().Display("flex").AlignItems("center").Gap("12px").
				Children(
					ui.Div().BackgroundColor(a.palette("#3366d6", "#4e7ff0")).Color(a.palette("#ffffff", "#1c2738")).
						Padding("8px 11px").BorderRadius("10px").
						Children(ui.Text("P").FontSize("19px").FontWeight("bold")),
					ui.Div().Children(
						ui.Text("Postmark").FontSize("20px").FontWeight("bold"),
						ui.P().Text("Mail  /  JMAP preview").Margin("2px 0 0").
							FontSize("11px").Color(a.palette("#7b8ba5", "#9cacc2")),
					),
				),
			ui.Div().Display("flex").AlignItems("center").Gap("10px").
				CSS("height", "40px").
				Children(
					ui.Input().ID("mail-search").Attr("type", "search").
						Attr("placeholder", "Search mail…").Attr("value", a.query).
						CSS("width", "240px").Height("36px").
						CSS("box-sizing", "border-box").Padding("7px 10px").
						CSS("border", a.palette("1px solid #dce4f0", "1px solid #39465a")).BorderRadius("9px").
						BackgroundColor(a.palette("#f6f8fc", "#27354a")),
					a.actionButton("Search", "search", false),
					ui.Text("Alex Morgan").FontWeight("bold").Color(a.palette("#485a77", "#d1deef")),
					a.actionButton("⚙ Settings", "settings", false),
					a.actionButton("−", "win-min", false),
					a.actionButton("□", "win-max", false),
					a.actionButton("×", "win-close", false),
				),
		)
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
			Display("flex").CSS("justify-content", "space-between").
			AlignItems("center").Padding("11px 12px").BorderRadius("9px").
			BackgroundColor(bg).Color(color).
			Children(
				ui.Text(box.Name).FontWeight("bold"),
				ui.Text(fmt.Sprintf("%d", a.count(box.ID))).FontSize("12px"),
			))
	}
	return ui.Nav().Width(mailSidebarWidth).CSS("flex-shrink", "0").
		BackgroundColor(a.palette("#ffffff", "#1c2738")).Padding("19px 12px").
		BorderRadius("10px").
		Children(
			a.actionButton("+  Compose", "compose", true).
				Width("100%").CSS("text-align", "center").Padding("13px 12px"),
			ui.P().Text("MAILBOXES").Margin("25px 10px 12px").
				FontSize("11px").FontWeight("bold").Color(a.palette("#96a3b6", "#a4b3c6")),
			ui.Div().Children(folders...),
			ui.Div().Margin("30px 10px 0").Padding("13px").
				BackgroundColor(a.palette("#f3f6fc", "#27354a")).BorderRadius("9px").
				Children(
					ui.Text("Mailbox status").FontWeight("bold").Color(a.palette("#415372", "#c7d4e8")),
					ui.P().Text(fmt.Sprintf("%d unread · %d messages", a.unreadCount(), len(a.messages))).
						Margin("8px 0 0").FontSize("12px").Color(a.palette("#7688a0", "#a1b3cb")),
				),
		)
}

func (a *app) messageList() *ui.Element {
	rows := []*ui.Element{}
	for _, message := range a.visibleMessages() {
		rows = append(rows, a.messageRow(message))
	}
	if len(rows) == 0 {
		rows = append(rows, ui.P().Text("No messages in this mailbox.").
			Padding("20px").Color(a.palette("#8391a8", "#9aabc3")))
	}
	label := a.folder
	for _, box := range a.mailboxes {
		if box.ID == a.folder {
			label = box.Name
			break
		}
	}
	return ui.Div().Width(mailListWidth).CSS("flex-shrink", "0").
		BackgroundColor(a.palette("#fbfcff", "#202c3d")).BorderRadius("10px").
		Display("flex").FlexDirection("column").
		Children(
			ui.Div().Padding("20px 19px 14px").
				CSS("border-bottom", a.palette("1px solid #e9edf4", "1px solid #39465a")).
				Children(
					ui.Text(label).FontSize("23px").FontWeight("bold"),
					ui.P().Text(fmt.Sprintf("%d conversations", len(a.visibleMessages()))).
						Color(a.palette("#8797ae", "#a5b4ca")).FontSize("12px").Margin("7px 0 0"),
					ui.Div().Margin("12px 0 0").Children(
						a.actionButton("Clear search", "clear-search", false),
					),
				),
			ui.Div().CSS("flex", "1").CSS("overflow-y", "auto").
				Children(rows...),
		)
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
		Padding("16px 17px").BackgroundColor(bg).
		CSS("border-bottom", a.palette("1px solid #e7ecf4", "1px solid #39465a")).
		Children(
			ui.Div().Display("flex").CSS("justify-content", "space-between").
				Children(
					ui.Text(m.From.Name+star).FontWeight(weight).Color(a.palette("#263954", "#e5edf9")).
						FontSize("13px"),
					ui.Text(shortDate(m.ReceivedAt)).FontSize("11px").Color(a.palette("#8b9bb1", "#9aabc3")),
				),
			ui.P().Text(m.Subject).Margin("9px 0 5px").
				FontWeight(weight).Color(a.palette("#263954", "#e5edf9")).FontSize("14px"),
			ui.P().Text(m.Preview).Margin("0").Color(a.palette("#7a8da6", "#b0bed2")).
				FontSize("12px").CSS("line-height", "1.5"),
		)
}

func (a *app) reader() *ui.Element {
	m := a.current()
	if m == nil || !a.isVisible(m) {
		return ui.Div().CSS("flex", "1").Padding("28px").
			Children(
				ui.Text("No message selected").FontSize("22px").FontWeight("bold"),
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
	return ui.Div().CSS("flex", "1").CSS("min-width", "0px").
		BackgroundColor(a.palette("#ffffff", "#1c2738")).BorderRadius("10px").
		Display("flex").FlexDirection("column").
		Children(
			ui.Div().Padding("15px 22px").
				CSS("border-bottom", a.palette("1px solid #e9edf4", "1px solid #39465a")).
				Display("flex").Gap("9px").
				Children(
					a.actionButton(star, "star", false),
					a.actionButton(status, "toggle-read", false),
					a.actionButton("Archive", "archive", false),
					a.actionButton("Delete", "delete", false),
				),
			ui.Div().Padding("30px 35px").CSS("overflow-y", "auto").
				CSS("flex", "1").
				Children(
					ui.Text(m.Subject).FontSize("26px").FontWeight("bold").
						Color(a.palette("#162944", "#f1f5ff")),
					ui.Div().Margin("23px 0").Padding("0 0 18px").
						CSS("border-bottom", a.palette("1px solid #e8edf5", "1px solid #39465a")).
						Children(
							ui.Text(m.From.Name).FontWeight("bold").
								FontSize("15px").Color(a.palette("#314560", "#d5e2f4")),
							ui.P().Text(m.From.Email+"  →  "+m.To.Email).
								Margin("7px 0 0").FontSize("12px").Color(a.palette("#7e91a8", "#a1b3cb")),
							ui.P().Text(displayDate(m.ReceivedAt)).
								Margin("5px 0 0").FontSize("12px").Color(a.palette("#7e91a8", "#a1b3cb")),
						),
					ui.P().Text(m.Body).FontSize("15px").Color(a.palette("#40526e", "#d1def0")).
						CSS("white-space", "pre-wrap").CSS("line-height", "1.7"),
					ui.Div().Margin("36px 0 0").Padding("15px").
						BackgroundColor(a.palette("#f4f7fe", "#27354a")).BorderRadius("9px").
						Children(
							ui.Text("Local JMAP preview").FontWeight("bold").
								Color(a.palette("#4a6496", "#b6ceff")),
							ui.P().Text("Mailbox membership and $seen/$flagged keywords update in memory only. No messages are sent or deleted on a server.").
								FontSize("12px").Color(a.palette("#7b8fa8", "#a4b6cf")),
						),
				),
		)
}

func (a *app) palette(light, dark string) string {
	if a.dark {
		return dark
	}
	return light
}

func (a *app) settingsPanel() *ui.Element {
	if !a.settings {
		return nil
	}
	mode := "Switch to dark theme"
	if a.dark {
		mode = "Switch to light theme"
	}
	return ui.Div().Display("flex").AlignItems("center").Gap("14px").
		Padding("10px 22px").BackgroundColor(a.palette("#ffffff", "#1c2738")).
		CSS("border-bottom", a.palette("1px solid #dce4f0", "1px solid #39465a")).
		Children(
			ui.Text("Appearance").FontWeight("bold"),
			a.actionButton(mode, "toggle-theme", false),
			ui.Text("Your preference applies to this local preview.").FontSize("12px").
				Color(a.palette("#73829c", "#9aabc3")),
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
		BackgroundColor(background).Color(foreground).
		CSS("border", border).Padding("9px 13px").
		BorderRadius("8px").FontSize("12px").FontWeight("bold")
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
