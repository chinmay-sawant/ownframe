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
	mailDragLimitY   = 78
	mailPlacement    = "jmap-mail"
)

type app struct {
	page       *ownframe.Page
	mailboxes  []mailbox
	messages   []mail
	folder     string
	selectedID string
	query      string
	drafts     int
	dark       bool
	settings   bool
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
		return nil, fmt.Errorf("new mail app: %w", err)
	}
	a.page = p
	// This example intentionally uses only OwnFrame's exported Page,
	// Handlers, and ui builder APIs. No internal packages are imported.
	p.Handle(ownframe.Handlers{Click: a.handleClick})
	return a, nil
}

func (a *app) Page() *ownframe.Page { return a.page }

func (a *app) handleClick(ctx context.Context, box ownframe.Box) error {
	if id, ok := strings.CutPrefix(box.Action, "folder:"); ok {
		a.selectFolder(id)
	} else if id, ok := strings.CutPrefix(box.Action, "message:"); ok {
		a.selectMessage(id)
	} else {
		switch box.Action {
		case "search":
			a.query = strings.TrimSpace(a.page.Form().FormValue("mail-search"))
			a.selectedID = ""
			a.ensureSelection()
		case "clear-search":
			a.query = ""
			a.selectedID = ""
			a.ensureSelection()
		case "star":
			a.toggleFlaggedCurrent()
		case "toggle-read":
			a.toggleSeenCurrent()
		case "archive":
			a.moveCurrentTo("archive")
		case "delete":
			a.moveCurrentTo("trash")
		case "compose":
			a.newDraft()
		case "settings":
			a.settings = !a.settings
		case "toggle-theme":
			a.dark = !a.dark
		case "win-min":
			a.page.Window().Minimize()
			return nil
		case "win-max":
			a.page.Window().ToggleMaximize()
			return nil
		case "win-close":
			a.page.Window().Close()
			return nil
		default:
			return nil
		}
	}

	return a.refresh(ctx)
}

func (a *app) refresh(ctx context.Context) error {
	fragment, err := ui.Render(a.screen())
	if err != nil {
		return fmt.Errorf("render mail view: %w", err)
	}
	// The builder emits a normal HTML fragment, and the public Load API
	// gives it back to OwnFrame's established Blinkless pipeline.
	if err := a.page.Load(ctx, renderMailDocument(fragment)); err != nil {
		return fmt.Errorf("reload mail view: %w", err)
	}
	return nil
}

func renderMailDocument(fragment string) string {
	return "<!doctype html><html><head><meta charset=\"utf-8\"></head><body>" +
		fragment + "</body></html>"
}

func (a *app) selectFolder(id string) {
	if !a.hasFolder(id) {
		return
	}
	a.folder = id
	a.selectedID = ""
	a.ensureSelection()
}

func (a *app) selectMessage(id string) {
	m := a.findMessage(id)
	if m == nil || !a.isVisible(m) {
		return
	}
	a.selectedID = id
	a.ensureKeywords(m)
	m.Keywords["$seen"] = true
}

func (a *app) toggleFlaggedCurrent() {
	m := a.current()
	if m == nil {
		return
	}
	a.ensureKeywords(m)
	m.Keywords["$flagged"] = !m.Keywords["$flagged"]
	if !a.isVisible(m) {
		a.selectedID = ""
		a.ensureSelection()
	}
}

func (a *app) toggleSeenCurrent() {
	m := a.current()
	if m == nil {
		return
	}
	a.ensureKeywords(m)
	m.Keywords["$seen"] = !m.Keywords["$seen"]
}

func (a *app) moveCurrentTo(folder string) {
	m := a.current()
	if m == nil {
		return
	}
	a.ensureMailboxIDs(m)
	if folder == "trash" {
		for id := range m.MailboxIDs {
			delete(m.MailboxIDs, id)
		}
	} else {
		delete(m.MailboxIDs, "inbox")
	}
	m.MailboxIDs[folder] = true
	a.selectedID = ""
	a.ensureSelection()
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

func (a *app) ensureKeywords(m *mail) {
	if m.Keywords == nil {
		m.Keywords = make(map[string]bool)
	}
}

func (a *app) ensureMailboxIDs(m *mail) {
	if m.MailboxIDs == nil {
		m.MailboxIDs = make(map[string]bool)
	}
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
		From:       contact{Name: "Alex Morgan", Email: "alex@example.com"},
		To:         contact{Name: "New recipient", Email: "recipient@example.net"},
		Subject:    "New draft · OwnFrame Mail",
		Preview:    "This is a local draft, not a sent email.",
		Body:       "Hi,\n\nI'm trying the new OwnFrame builder mail example.\n\nThanks,\nAlex",
		ReceivedAt: time.Now(),
		MailboxIDs: map[string]bool{"drafts": true},
		Keywords:   map[string]bool{"$draft": true, "$seen": true},
	}
	a.messages = append([]mail{m}, a.messages...)
	a.folder, a.selectedID = "drafts", id
	a.query = ""
}
