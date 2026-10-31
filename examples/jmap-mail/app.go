package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	mailWindowWidth  = 1380
	mailWindowHeight = 810
	mailCanvasHeight = "790px"
	mailSidebarWidth = "226px"
	mailListWidth    = "394px"
	mailDemoTitle    = "Postmark · JMAP Mail Preview"
)

type app struct {
	page       *ownframe.Page
	mailboxes  []mailbox
	messages   []mail
	folder     string
	selectedID string
	query      string
	drafts     int
}

func newApp() (*app, error) {
	a := &app{
		mailboxes:  sampleMailboxes(),
		messages:   sampleMessages(),
		folder:     "inbox",
		selectedID: "m-101",
	}
	p, err := ui.NewPage(ownframe.Config{
		Title: mailDemoTitle, Width: mailWindowWidth, Height: mailWindowHeight,
	}, a.screen())
	if err != nil {
		return nil, err
	}
	a.page = p
	// This example intentionally uses only OwnFrame's exported Page,
	// Handlers, and ui builder APIs. No internal packages are imported.
	p.Handle(ownframe.Handlers{Click: a.handleClick})
	return a, nil
}

func (a *app) Page() *ownframe.Page { return a.page }

func (a *app) handleClick(ctx context.Context, box ownframe.Box) error {
	action := box.Action
	switch {
	case strings.HasPrefix(action, "folder:"):
		folderID := strings.TrimPrefix(action, "folder:")
		if !a.hasFolder(folderID) {
			return nil
		}
		a.folder = folderID
		a.selectedID = ""
		a.ensureSelection()
	case strings.HasPrefix(action, "message:"):
		id := strings.TrimPrefix(action, "message:")
		if m := a.findMessage(id); m != nil && a.isVisible(m) {
			a.selectedID = id
			m.Keywords["$seen"] = true
		}
	case action == "search":
		a.query = strings.TrimSpace(a.page.Form().FormValue("mail-search"))
		a.selectedID = ""
		a.ensureSelection()
	case action == "clear-search":
		a.query = ""
		a.selectedID = ""
		a.ensureSelection()
	case action == "star":
		if m := a.current(); m != nil {
			m.Keywords["$flagged"] = !m.Keywords["$flagged"]
			if !a.isVisible(m) {
				a.selectedID = ""
				a.ensureSelection()
			}
		}
	case action == "toggle-read":
		if m := a.current(); m != nil {
			m.Keywords["$seen"] = !m.Keywords["$seen"]
		}
	case action == "archive":
		if m := a.current(); m != nil {
			delete(m.MailboxIDs, "inbox")
			m.MailboxIDs["archive"] = true
			a.selectedID = ""
			a.ensureSelection()
		}
	case action == "delete":
		if m := a.current(); m != nil {
			for id := range m.MailboxIDs {
				delete(m.MailboxIDs, id)
			}
			m.MailboxIDs["trash"] = true
			a.selectedID = ""
			a.ensureSelection()
		}
	case action == "compose":
		a.newDraft()
	default:
		return nil
	}

	fragment, err := ui.Render(a.screen())
	if err != nil {
		return err
	}
	// The builder emits a normal HTML fragment, and the public Load API
	// gives it back to OwnFrame's established Blinkless pipeline.
	source := "<!doctype html><html><head><meta charset=\"utf-8\"></head><body>" +
		fragment + "</body></html>"
	return a.page.Load(ctx, source)
}

func (a *app) hasFolder(id string) bool {
	for _, box := range a.mailboxes {
		if box.ID == id {
			return true
		}
	}
	return false
}

func (a *app) findMessage(id string) *mail {
	for i := range a.messages {
		if a.messages[i].ID == id {
			return &a.messages[i]
		}
	}
	return nil
}

func (a *app) current() *mail {
	return a.findMessage(a.selectedID)
}

func (a *app) isVisible(m *mail) bool {
	if m == nil {
		return false
	}
	matchesFolder := m.MailboxIDs[a.folder]
	if a.folder == "starred" {
		matchesFolder = m.Keywords["$flagged"] && !m.MailboxIDs["trash"]
	}
	if !matchesFolder {
		return false
	}
	if a.query == "" {
		return true
	}
	needle := strings.ToLower(a.query)
	haystack := strings.ToLower(m.Subject + " " + m.From.Name + " " +
		m.From.Email + " " + m.Preview + " " + m.Body)
	return strings.Contains(haystack, needle)
}

func (a *app) visibleMessages() []*mail {
	var found []*mail
	for i := range a.messages {
		m := &a.messages[i]
		if a.isVisible(m) {
			found = append(found, m)
		}
	}
	return found
}

func (a *app) count(folderID string) int {
	total := 0
	for i := range a.messages {
		m := &a.messages[i]
		if folderID == "starred" {
			if m.Keywords["$flagged"] && !m.MailboxIDs["trash"] {
				total++
			}
			continue
		}
		if m.MailboxIDs[folderID] {
			total++
		}
	}
	return total
}

func (a *app) unreadCount() int {
	n := 0
	for i := range a.messages {
		m := &a.messages[i]
		if m.MailboxIDs["inbox"] && !m.Keywords["$seen"] {
			n++
		}
	}
	return n
}

func (a *app) ensureSelection() {
	if current := a.current(); a.isVisible(current) {
		return
	}
	a.selectedID = ""
	if visible := a.visibleMessages(); len(visible) != 0 {
		a.selectedID = visible[0].ID
	}
}

func (a *app) newDraft() {
	a.drafts++
	id := fmt.Sprintf("draft-%d", a.drafts)
	m := mail{
		ID: id, ThreadID: "thread-" + id,
		From: contact{Name: "Alex Morgan", Email: "alex@example.com"},
		To: contact{Name: "New recipient", Email: "recipient@example.net"},
		Subject: "New draft · OwnFrame Mail",
		Preview: "This is a local draft, not a sent email.",
		Body: "Hi,\n\nI'm trying the new OwnFrame builder mail example.\n\nThanks,\nAlex",
		ReceivedAt: "2026-10-31T18:30:00+05:30",
		MailboxIDs: map[string]bool{"drafts": true},
		Keywords: map[string]bool{"$draft": true, "$seen": true},
	}
	a.messages = append([]mail{m}, a.messages...)
	a.folder, a.selectedID = "drafts", id
	a.query = ""
}

func (a *app) screen() *ui.Element {
	return ui.Div().Width("100%").Height(mailCanvasHeight).
		Display("flex").FlexDirection("column").
		BackgroundColor("#f5f7fb").Color("#172b4d").
		CSS("font-family", "sans-serif").FontSize("14px").
		Children(
			a.toolbar(),
			ui.Div().Display("flex").CSS("flex", "1").
				CSS("min-height", "0px").
				Children(a.sidebar(), a.messageList(), a.reader()),
			ui.Div().Padding("9px 18px").FontSize("11px").Color("#73829c").
				CSS("border-top", "1px solid #dce4f0").
				Children(ui.Text("JMAP-inspired demo  ·  Local sample data  ·  No account, network connection, or sent messages")),
		)
}

func (a *app) toolbar() *ui.Element {
	return ui.Div().Display("flex").FlexDirection("row").
		AlignItems("center").CSS("justify-content", "space-between").
		Padding("16px 22px").BackgroundColor("#ffffff").
		CSS("border-bottom", "1px solid #dce4f0").
		Children(
			ui.Div().Display("flex").AlignItems("center").Gap("12px").
				Children(
					ui.Div().BackgroundColor("#3366d6").Color("#ffffff").
						Padding("8px 11px").BorderRadius("10px").
						Children(ui.Text("P").FontSize("19px").FontWeight("bold")),
					ui.Div().Children(
						ui.Text("Postmark").FontSize("20px").FontWeight("bold"),
						ui.P().Text("Mail  /  JMAP preview").Margin("2px 0 0").
							FontSize("11px").Color("#7b8ba5"),
					),
				),
			ui.Div().Display("flex").AlignItems("center").Gap("10px").
				Children(
					ui.Input().ID("mail-search").Attr("type", "search").
						Attr("placeholder", "Search mail…").Attr("value", a.query).
						CSS("width", "240px").Padding("9px").
						CSS("border", "1px solid #dce4f0").BorderRadius("9px").
						BackgroundColor("#f6f8fc"),
					actionButton("Search", "search", false),
					ui.Text("Alex Morgan").FontWeight("bold").Color("#485a77"),
				),
		)
}

func (a *app) sidebar() *ui.Element {
	folders := make([]*ui.Element, 0, len(a.mailboxes))
	for _, box := range a.mailboxes {
		color := "#60718c"
		bg := "transparent"
		if a.folder == box.ID {
			color = "#234fa9"
			bg = "#e5edff"
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
		BackgroundColor("#ffffff").Padding("19px 12px").
		CSS("border-right", "1px solid #dce4f0").
		Children(
			actionButton("+  Compose", "compose", true).
				Width("100%").CSS("text-align", "center").Padding("13px 12px"),
			ui.P().Text("MAILBOXES").Margin("25px 10px 12px").
				FontSize("11px").FontWeight("bold").Color("#96a3b6"),
			ui.Div().Children(folders...),
			ui.Div().Margin("30px 10px 0").Padding("13px").
				BackgroundColor("#f3f6fc").BorderRadius("9px").
				Children(
					ui.Text("Mailbox status").FontWeight("bold").Color("#415372"),
					ui.P().Text(fmt.Sprintf("%d unread · %d messages", a.unreadCount(), len(a.messages))).
						Margin("8px 0 0").FontSize("12px").Color("#7688a0"),
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
			Padding("20px").Color("#8391a8"))
	}
	label := a.folder
	for _, box := range a.mailboxes {
		if box.ID == a.folder {
			label = box.Name
			break
		}
	}
	return ui.Div().Width(mailListWidth).CSS("flex-shrink", "0").
		BackgroundColor("#fbfcff").CSS("border-right", "1px solid #dce4f0").
		Display("flex").FlexDirection("column").
		Children(
			ui.Div().Padding("20px 19px 14px").
				CSS("border-bottom", "1px solid #e9edf4").
				Children(
					ui.Text(label).FontSize("23px").FontWeight("bold"),
					ui.P().Text(fmt.Sprintf("%d conversations", len(a.visibleMessages()))).
						Color("#8797ae").FontSize("12px").Margin("7px 0 0"),
					ui.Div().Margin("12px 0 0").Children(
						actionButton("Clear search", "clear-search", false),
					),
				),
			ui.Div().CSS("flex", "1").CSS("overflow-y", "auto").
				Children(rows...),
		)
}

func (a *app) messageRow(m *mail) *ui.Element {
	bg := "#ffffff"
	if m.ID == a.selectedID {
		bg = "#e9f0ff"
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
		CSS("border-bottom", "1px solid #e7ecf4").
		Children(
			ui.Div().Display("flex").CSS("justify-content", "space-between").
				Children(
					ui.Text(m.From.Name+star).FontWeight(weight).Color("#263954").
						FontSize("13px"),
					ui.Text(shortDate(m.ReceivedAt)).FontSize("11px").Color("#8b9bb1"),
				),
			ui.P().Text(m.Subject).Margin("9px 0 5px").
				FontWeight(weight).Color("#263954").FontSize("14px"),
			ui.P().Text(m.Preview).Margin("0").Color("#7a8da6").
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
		BackgroundColor("#ffffff").Display("flex").FlexDirection("column").
		Children(
			ui.Div().Padding("15px 22px").
				CSS("border-bottom", "1px solid #e9edf4").
				Display("flex").Gap("9px").
				Children(
					actionButton(star, "star", false),
					actionButton(status, "toggle-read", false),
					actionButton("Archive", "archive", false),
					actionButton("Delete", "delete", false),
				),
			ui.Div().Padding("30px 35px").CSS("overflow-y", "auto").
				CSS("flex", "1").
				Children(
					ui.Text(m.Subject).FontSize("26px").FontWeight("bold").
						Color("#162944"),
					ui.Div().Margin("23px 0").Padding("0 0 18px").
						CSS("border-bottom", "1px solid #e8edf5").
						Children(
							ui.Text(m.From.Name).FontWeight("bold").
								FontSize("15px").Color("#314560"),
							ui.P().Text(m.From.Email+"  →  "+m.To.Email).
								Margin("7px 0 0").FontSize("12px").Color("#7e91a8"),
							ui.P().Text(displayDate(m.ReceivedAt)).
								Margin("5px 0 0").FontSize("12px").Color("#7e91a8"),
						),
					ui.P().Text(m.Body).FontSize("15px").Color("#40526e").
						CSS("white-space", "pre-wrap").CSS("line-height", "1.7"),
					ui.Div().Margin("36px 0 0").Padding("15px").
						BackgroundColor("#f4f7fe").BorderRadius("9px").
						Children(
							ui.Text("Local JMAP preview").FontWeight("bold").
								Color("#4a6496"),
							ui.P().Text("Mailbox membership and $seen/$flagged keywords update in memory only. No messages are sent or deleted on a server.").
								FontSize("12px").Color("#7b8fa8"),
						),
				),
		)
}

func actionButton(label, action string, prominent bool) *ui.Element {
	background, foreground, border := "#ffffff", "#516684", "1px solid #dce4f0"
	if prominent {
		background, foreground, border = "#3366d6", "#ffffff", "1px solid #3366d6"
	}
	return ui.Button(label).Attr("data-action", action).
		BackgroundColor(background).Color(foreground).
		CSS("border", border).Padding("9px 13px").
		BorderRadius("8px").FontSize("12px").FontWeight("bold")
}

func shortDate(value string) string {
	date, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}
	return date.Format("02 Jan")
}

func displayDate(value string) string {
	date, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}
	return date.Format("02 Jan 2006 · 15:04 MST")
}
