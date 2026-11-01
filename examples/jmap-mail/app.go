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
	mailSidebarWidth = "226px"
	mailListWidth    = "394px"
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
	mailScrollThumbMin  = 24
	mailToastLifetime   = 10 * time.Second
	// mailRefreshPasses bounds the measure-then-paint loop in refresh.
	mailRefreshPasses = 2
	// mailDefaultVisibleRows is the list window before the first box
	// measurement lands in the scroll cache.
	mailDefaultVisibleRows = 5
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
	// listOffset is the first visible message row. Rows are fixed height,
	// so the visible count derives from the viewport box and the thumb
	// shows only when rows overflow it.
	listOffset int
	listAccum  float64
	listViewH  int
	// readerScroll is the pixel offset of the reader body. The body has
	// no click targets, so a 2D transform moves it without disturbing
	// hit-testing; the thumb shows only when content exceeds the viewport.
	readerScroll   int
	readerAccum    float64
	readerViewH    int
	readerContentH int
	toast          string
	toastUntil     time.Time
	lastW          int
	lastH          int
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
	a.lastW, a.lastH = p.Size()
	// This example intentionally uses only OwnFrame's exported Page,
	// Handlers, and ui builder APIs. No internal packages are imported.
	p.Handle(ownframe.Handlers{Click: a.handleClick, Wheel: a.handleWheel})
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
			a.listOffset, a.listAccum = 0, 0
			a.readerScroll, a.readerAccum = 0, 0
			a.ensureSelection()
		case "clear-search":
			a.query = ""
			a.selectedID = ""
			a.listOffset, a.listAccum = 0, 0
			a.readerScroll, a.readerAccum = 0, 0
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
	// Render twice at most: the scroll cache (viewport and content boxes)
	// is measured after each Load, so a second pass paints the thumbs the
	// fresh measurements call for. Heights are layout-stable, so it
	// converges immediately.
	for i := 0; i < mailRefreshPasses; i++ {
		fragment, err := ui.Render(a.screen())
		if err != nil {
			return fmt.Errorf("render mail view: %w", err)
		}
		// The builder emits a normal HTML fragment, and the public Load API
		// gives it back to OwnFrame's established Blinkless pipeline.
		if err := a.page.Load(ctx, a.renderMailDocument(fragment)); err != nil {
			return fmt.Errorf("reload mail view: %w", err)
		}
		before := a.scrollCacheKey()
		a.updateScrollCache()
		if a.scrollCacheKey() == before {
			break
		}
	}
	a.lastW, a.lastH = a.page.Size()
	if a.toast != "" {
		if remaining := time.Until(a.toastUntil); remaining > 0 {
			a.page.WantFrameIn(remaining)
		}
	}
	return nil
}

func (a *app) scrollCacheKey() [4]int {
	return [4]int{a.listViewH, a.listOffset, a.readerViewH, a.readerContentH}
}

func (a *app) renderMailDocument(fragment string) string {
	// Mirror ui-builder: reset the default 8px body margin or it shows as a
	// white border on all four sides, and paint html/body with the active
	// theme so the edge matches the app instead of flashing white.
	return "<!doctype html><html><head><meta charset=\"utf-8\"><style>html,body{margin:0;padding:0;background-color:" +
		a.palette("#f5f7fb", "#111827") + "}</style></head><body>" +
		fragment + "</body></html>"
}

func (a *app) selectFolder(id string) {
	if !a.hasFolder(id) {
		return
	}
	a.folder = id
	a.selectedID = ""
	a.listOffset, a.listAccum = 0, 0
	a.readerScroll, a.readerAccum = 0, 0
	a.ensureSelection()
}

func (a *app) selectMessage(id string) {
	m := a.findMessage(id)
	if m == nil || !a.isVisible(m) {
		return
	}
	a.selectedID = id
	a.readerScroll, a.readerAccum = 0, 0
	a.ensureKeywords(m)
	m.Keywords["$seen"] = true
}

func (a *app) selectTheme(name string) {
	switch name {
	case "light":
		a.dark = false
	case "dark":
		a.dark = true
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
		a.readerScroll, a.readerAccum = 0, 0
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
	a.readerScroll, a.readerAccum = 0, 0
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
	a.listOffset, a.listAccum = 0, 0
	a.readerScroll, a.readerAccum = 0, 0
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
	a.listOffset, a.listAccum = 0, 0
	a.readerScroll, a.readerAccum = 0, 0
	a.showToast("Message sent — moved to Sent.")
}

func (a *app) showToast(message string) {
	a.toast = message
	a.toastUntil = time.Now().Add(mailToastLifetime)
}

// tick clears an expired toast and refreshes once when the window size
// changes, so the visible row counts track the live resolution.
func (a *app) tick(ctx context.Context) error {
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

// scrollListByWheel moves by whole rows, keeping the trackpad fraction in
// Accum. It reports whether the first visible row changed.
func (a *app) scrollListByWheel(dy float64) bool {
	a.listAccum += -dy * mailListWheelRows
	rows := int(a.listAccum)
	a.listAccum -= float64(rows)
	if rows == 0 {
		return false
	}
	before := a.listOffset
	a.listOffset = clampListOffset(a.listOffset+rows, len(a.visibleMessages()), a.listVisibleCount())
	return a.listOffset != before
}

// scrollReaderByWheel moves the reader body by pixels. It reports whether
// the offset changed.
func (a *app) scrollReaderByWheel(dy float64) bool {
	a.readerAccum += -dy * mailReaderWheelStep
	step := int(a.readerAccum)
	a.readerAccum -= float64(step)
	if step == 0 {
		return false
	}
	before := a.readerScroll
	maxOffset := max(a.readerContentH-a.readerViewH, 0)
	a.readerScroll = min(max(a.readerScroll+step, 0), maxOffset)
	return a.readerScroll != before
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
	a.listOffset = clampListOffset(a.listOffset, total, visible)
	end := min(a.listOffset+visible, total)
	if end < a.listOffset {
		end = a.listOffset
	}
	return a.listOffset, end
}

func clampListOffset(offset, total, visible int) int {
	maxOffset := total - visible
	if maxOffset < 0 {
		maxOffset = 0
	}
	return max(0, min(offset, maxOffset))
}

// listThumb lays the list scrollbar thumb over the viewport height. It
// reports show false when every row fits, leaving the track empty.
func (a *app) listThumb(total, visible int) (top, height int, show bool) {
	track := a.listViewH
	if track < 1 || visible < 1 || total <= visible {
		return 0, 0, false
	}
	height = track * visible / total
	height = min(max(height, mailScrollThumbMin, 1), track)
	offset := clampListOffset(a.listOffset, total, visible)
	travel := track - height
	top = travel * offset / max(total-visible, 1)
	return top, height, true
}

// readerThumb lays the reader scrollbar thumb out the same way in pixels.
func (a *app) readerThumb() (top, height int, show bool) {
	track, content, view := a.readerViewH, a.readerContentH, a.readerViewH
	if track < 1 || content <= view {
		return 0, 0, false
	}
	height = track * view / content
	height = min(max(height, mailScrollThumbMin, 1), track)
	maxOffset := max(content-view, 1)
	travel := track - height
	top = travel * min(max(a.readerScroll, 0), maxOffset) / maxOffset
	return top, height, true
}

// updateScrollCache rereads the viewport and content boxes after a redraw
// and pins both offsets to the live range.
func (a *app) updateScrollCache() {
	for _, b := range a.page.Boxes() {
		switch b.ID {
		case "mail-list-viewport":
			if h := int(b.H); h > 0 {
				a.listViewH = h
			}
		case "mail-reader-viewport":
			if h := int(b.H); h > 0 {
				a.readerViewH = h
			}
		case "mail-reader-content":
			if h := int(b.H); h > 0 {
				a.readerContentH = h
			}
		}
	}
	a.listOffset = clampListOffset(a.listOffset, len(a.visibleMessages()), a.listVisibleCount())
	maxReader := max(a.readerContentH-a.readerViewH, 0)
	a.readerScroll = min(max(a.readerScroll, 0), maxReader)
}
