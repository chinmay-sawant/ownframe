package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

const (
	mailWindowWidth  = 1380
	mailWindowHeight = 810
	mailSidebarWidth = 226
	mailListWidth    = 394
	mailDemoTitle    = "Postmark · JMAP Mail Preview"
	mailDragLimitY   = 78
	mailPlacement    = "jmap-mail"
	// mailRowH is the fixed message-row height. Fixed rows keep wheel
	// scrolling and scrollbar math exact: visible rows derive from the
	// viewport height, so a shorter window shows a scrollbar.
	mailRowH = 104
	// mailListWheelRows is rows per wheel notch; fractions accumulate for
	// trackpads. The reader scrolls 48 px per notch like a page.
	mailListWheelRows   = 1.0
	mailReaderWheelStep = 48.0
	mailToastLifetime   = 10 * time.Second
	// mailRefreshPasses bounds the measure-then-paint loop in refresh.
	mailRefreshPasses = 2
	// mailDefaultVisibleRows is the list window before the first box
	// measurement lands in the scroll cache.
	mailDefaultVisibleRows = 5
)

type app struct {
	modalFocus chadcn.DialogFocus
	view       *ui.View
	themeDirty bool
	page       *ownframe.Page
	mailboxes  []mailbox
	messages   []mail
	folder     string
	selectedID string
	query      string
	drafts     int
	appearance mailTheme
	settings   bool
	// The list scrolls in whole rows; the reader scrolls in CSS pixels.
	listScroll           chadcn.ScrollState
	readerScroll         chadcn.ScrollState
	listViewH            int
	scrollMeasurePending bool
	toast                string
	toastUntil           time.Time
	lastW                int
	lastH                int
}

func newApp() (*app, error) {
	a := &app{
		mailboxes:  sampleMailboxes(),
		messages:   sampleMessages(),
		folder:     "inbox",
		appearance: mailThemeLight,
		selectedID: "m-101",
	}
	view, err := ui.Mount(ownframe.Config{
		Theme: theme.Light + chadcn.Stylesheet,
		Title: mailDemoTitle, Width: mailWindowWidth, Height: mailWindowHeight,
	}, a.screen())
	if err != nil {
		return nil, fmt.Errorf("new mail app: %w", err)
	}
	a.view = view
	p := view.Page
	a.page = p
	a.lastW, a.lastH = p.Size()
	// This example intentionally uses only OwnFrame's exported Page,
	// Handlers, and ui builder APIs. No internal packages are imported.
	view.Handle(ownframe.Handlers{Click: a.handleClick, Wheel: a.handleWheel, KeyDown: a.handleKey,
		DragStart: a.startScrollDrag, DragMove: a.moveScrollDrag, DragEnd: a.endScrollDrag})
	p.SetTick(a.tick)
	return a, nil
}

func (a *app) Page() *ownframe.Page { return a.page }

func (a *app) handleClick(ctx context.Context, box ownframe.Box) error {
	if id, ok := strings.CutPrefix(box.Action, "folder:"); ok {
		a.selectFolder(id)
	} else if id, ok := strings.CutPrefix(box.Action, "message:"); ok {
		a.selectMessage(id)
	} else if name, ok := strings.CutPrefix(box.Action, "theme:"); ok {
		a.selectTheme(name)
	} else {
		switch box.Action {
		case "search":
			a.query = strings.TrimSpace(a.page.Form().FormValue("mail-search"))
			a.selectedID = ""
			a.listScroll.ScrollTo(0)
			a.readerScroll.ScrollTo(0)
			a.ensureSelection()
		case "clear-search":
			a.query = ""
			a.page.Form().SetFormValue("mail-search", "")
			a.selectedID = ""
			a.listScroll.ScrollTo(0)
			a.readerScroll.ScrollTo(0)
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
		case "send":
			a.sendCurrent()
		case "toast-close":
			a.toast = ""
			a.toastUntil = time.Time{}
		case "settings":
			a.settings = !a.settings
		case "win-dev":
			a.page.Window().SetDevTools(!a.page.Window().DevTools())
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
	if a.settings {
		a.modalFocus.Activate(a.page, "settings-close", "theme-light", "theme-dark", "theme-custom")
	} else {
		a.modalFocus.Deactivate(a.page)
	}

	if a.themeDirty {
		sheet, err := a.appearance.stylesheet()
		if err != nil {
			return err
		}
		if err := a.page.SetTheme(sheet); err != nil {
			return err
		}
		a.themeDirty = false
	}

	// Render twice at most: the scroll cache (viewport and content boxes)
	// is measured after each Load, so a second pass paints the thumbs the
	// fresh measurements call for. Heights are layout-stable, so it
	// converges immediately.
	for i := 0; i < mailRefreshPasses; i++ {
		if err := a.view.Update(ctx, a.screen()); err != nil {
			return fmt.Errorf("update mail view: %w", err)
		}
		before := a.scrollCacheKey()
		a.updateScrollCache()
		if a.scrollCacheKey() == before {
			break
		}
	}
	a.scrollMeasurePending = true
	a.lastW, a.lastH = a.page.Size()
	if a.toast != "" {
		if remaining := time.Until(a.toastUntil); remaining > 0 {
			a.page.WantFrameIn(remaining)
		}
	}
	return nil
}

func (a *app) scrollCacheKey() [4]int {
	reader := a.readerScroll.Metrics()
	return [4]int{a.listViewH, a.listScroll.Offset(), reader.Viewport, reader.Content}
}

func (a *app) selectFolder(id string) {
	if !a.hasFolder(id) {
		return
	}
	a.folder = id
	a.selectedID = ""
	a.listScroll.ScrollTo(0)
	a.readerScroll.ScrollTo(0)
	a.ensureSelection()
}

func (a *app) selectMessage(id string) {
	m := a.findMessage(id)
	if m == nil || !a.isVisible(m) {
		return
	}
	a.selectedID = id
	a.readerScroll.ScrollTo(0)
	a.ensureKeywords(m)
	m.Keywords["$seen"] = true
}

func (a *app) selectTheme(name string) {
	switch mailTheme(name) {
	case mailThemeLight, mailThemeDark, mailThemeCustom:
		a.appearance = mailTheme(name)
		a.themeDirty = true
	}
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
		a.readerScroll.ScrollTo(0)
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
	a.readerScroll.ScrollTo(0)
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
	a.listScroll.ScrollTo(0)
	a.readerScroll.ScrollTo(0)
}

// sendCurrent moves the open draft to Sent and raises the toast. A
// non-draft selection has nothing to send.
func (a *app) sendCurrent() {
	m := a.current()
	if m == nil {
		return
	}
	a.ensureKeywords(m)
	if !m.Keywords["$draft"] {
		return
	}
	a.ensureMailboxIDs(m)
	delete(m.MailboxIDs, "drafts")
	m.MailboxIDs["sent"] = true
	delete(m.Keywords, "$draft")
	m.Keywords["$seen"] = true
	m.ReceivedAt = time.Now()
	a.folder, a.selectedID = "sent", m.ID
	a.query = ""
	a.listScroll.ScrollTo(0)
	a.readerScroll.ScrollTo(0)
	a.showToast("Message sent — moved to Sent.")
}

func (a *app) showToast(message string) {
	a.toast = message
	a.toastUntil = time.Now().Add(mailToastLifetime)
}

// tick clears an expired toast and refreshes once when the window size
// changes, so the visible row counts track the live resolution.
func (a *app) tick(ctx context.Context) error {
	if a.scrollMeasurePending {
		a.scrollMeasurePending = false
		before := a.scrollCacheKey()
		a.updateScrollCache()
		if before != a.scrollCacheKey() {
			return a.refresh(ctx)
		}
	}
	if w, h := a.page.Size(); w != a.lastW || h != a.lastH {
		return a.refresh(ctx)
	}
	if a.toast == "" {
		return nil
	}
	if !time.Now().After(a.toastUntil) {
		a.page.WantFrameIn(time.Until(a.toastUntil))
		return nil
	}
	a.toast = ""
	a.toastUntil = time.Time{}
	return a.refresh(ctx)
}

// handleWheel scrolls the message list or the reader body under the cursor.
// It consumes the event so the window does not pan the 100vh page, which
// would move the toolbar instead of the pane.
func (a *app) handleWheel(ctx context.Context, x, y, dx, dy float64) (bool, error) {
	if dx == 0 && dy == 0 {
		return false, nil
	}
	if a.settings {
		return true, nil
	}
	sx, sy := a.page.Scroll().ScrollOffset()
	x -= float64(sx)
	y -= float64(sy)
	boxes := a.page.Boxes()
	if overBoxID(boxes, "mail-list-viewport", x, y) {
		if a.scrollListByWheel(dy) {
			return true, a.refresh(ctx)
		}
		return true, nil
	}
	if overBoxID(boxes, "mail-reader-viewport", x, y) {
		if a.scrollReaderByWheel(dy) {
			return true, a.refresh(ctx)
		}
		return true, nil
	}
	return false, nil
}

func overBoxID(boxes []ownframe.Box, id string, x, y float64) bool {
	for _, b := range boxes {
		if b.ID == id {
			return x >= b.X && x <= b.X+b.W && y >= b.Y && y <= b.Y+b.H
		}
	}
	return false
}

func (a *app) scrollListByWheel(dy float64) bool {
	a.listScroll.SetRange(len(a.visibleMessages()), a.listVisibleCount())
	return a.listScroll.Wheel(dy, mailListWheelRows)
}

func (a *app) scrollReaderByWheel(dy float64) bool {
	return a.readerScroll.Wheel(dy, mailReaderWheelStep)
}

// listVisibleCount derives the row window from the live viewport box, so a
// shorter window fits fewer rows and shows the scrollbar sooner.
func (a *app) listVisibleCount() int {
	if a.listViewH > 0 {
		return max(a.listViewH/mailRowH, 1)
	}
	return mailDefaultVisibleRows
}

// listWindow pins the offset and returns the visible slice bounds.
func (a *app) listWindow(total, visible int) (int, int) {
	a.listScroll.SetRange(total, visible)
	start := a.listScroll.Offset()
	return start, min(start+visible, total)
}

func (a *app) updateScrollCache() {
	reader := a.readerScroll.Metrics()
	for _, box := range a.page.Boxes() {
		switch box.ID {
		case "mail-list-viewport":
			a.listViewH = max(int(box.H), 0)
		case "mail-reader-viewport":
			reader.Viewport = max(int(box.H), 0)
		case "mail-reader-content":
			reader.Content = max(int(box.H), 0)
		}
	}
	a.listScroll.SetRange(len(a.visibleMessages()), a.listVisibleCount())
	a.readerScroll.SetRange(reader.Content, reader.Viewport)
}

func (a *app) handleKey(ctx context.Context, key string) error {
	if !a.settings {
		switch a.page.Form().FocusID() {
		case "mail-list-thumb":
			if a.listScroll.Key(key, 1) {
				return a.refresh(ctx)
			}
		case "mail-reader-thumb":
			if a.readerScroll.Key(key, int(mailReaderWheelStep)) {
				return a.refresh(ctx)
			}
		}
	}
	if key == "escape" && a.settings {
		a.settings = false
		return a.refresh(ctx)
	}
	return nil
}
