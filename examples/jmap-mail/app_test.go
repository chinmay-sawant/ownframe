package main

import (
	"bytes"
	"context"
	"fmt"
	"image/color"
	"image/png"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	expectedUnreadInbox        = 2
	expectedMessagesAfterDraft = 9
)

func TestMailExampleRendersThroughPublicBuilder(t *testing.T) {
	t.Parallel()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	source := a.Page().HTML()
	for _, expected := range []string{
		"Postmark", "JMAP preview", "Design review", "Maya Patel",
		"data-action", "mail-search",
	} {
		if !strings.Contains(source, expected) {
			t.Errorf("mail source missing %q", expected)
		}
	}
	if _, err := ui.Render(a.screen()); err != nil {
		t.Fatalf("public UI builder should serialize mail example: %v", err)
	}
	if got := a.unreadCount(); got != expectedUnreadInbox {
		t.Fatalf("unread inbox: got %d, want %d", got, expectedUnreadInbox)
	}
}

func TestMailSelectionAndNativeEventRouting(t *testing.T) {
	t.Parallel()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.Page(), "message:m-102"); err != nil {
		t.Fatal(err)
	}
	if a.selectedID != "m-102" {
		t.Fatalf("selected %q, want m-102", a.selectedID)
	}
	if !a.findMessage("m-102").Keywords["$seen"] {
		t.Fatal("opening message should mark it seen")
	}
	if strings.Contains(a.Page().HTML(), "No message selected") {
		t.Fatal("reader lost selected message")
	}
}

func TestMailFoldersStarArchiveAndDraftAreLocal(t *testing.T) {
	t.Parallel()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	action := func(name string) {
		t.Helper()
		if err := a.handleClick(ctx, ownframe.Box{Action: name}); err != nil {
			t.Fatal(err)
		}
	}
	action("star")
	if a.findMessage("m-101").Keywords["$flagged"] {
		t.Fatal("star should toggle flagged keyword")
	}
	action("archive")
	if !a.findMessage("m-101").MailboxIDs["archive"] {
		t.Fatal("archive should add archive mailbox ID")
	}
	if a.findMessage("m-101").MailboxIDs["inbox"] {
		t.Fatal("archive should remove inbox membership")
	}
	action("folder:archive")
	if a.folder != "archive" || a.selectedID != "m-101" {
		t.Fatalf("folder selection: folder=%q selected=%q", a.folder, a.selectedID)
	}
	action("compose")
	if a.folder != "drafts" || a.findMessage(a.selectedID) == nil {
		t.Fatalf("composed draft missing: folder=%q selected=%q", a.folder, a.selectedID)
	}
	if !a.findMessage(a.selectedID).Keywords["$draft"] {
		t.Fatal("composed message should be a local draft")
	}
	if len(a.messages) != expectedMessagesAfterDraft {
		t.Fatalf("messages after draft: got %d want %d", len(a.messages), expectedMessagesAfterDraft)
	}
}

func TestMailListWheelScrollsAndShowsScrollbar(t *testing.T) {
	t.Parallel()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		a.messages = append(a.messages, mail{
			ID: fmt.Sprintf("extra-%d", i), ThreadID: "extra",
			From:       contact{Name: "Extra", Email: "extra@example.com"},
			To:         contact{Name: "Alex Morgan", Email: "alex@example.com"},
			Subject:    fmt.Sprintf("Extra message %d", i),
			Preview:    "Overflow fixture for the list scrollbar.",
			Body:       "Hi,\n\nOverflow fixture.\n\nBye",
			ReceivedAt: time.Now(),
			MailboxIDs: map[string]bool{"inbox": true},
			Keywords:   map[string]bool{"$seen": true},
		})
	}
	a.listViewH = 260
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(a.Page().HTML(), "mail-list-thumb") {
		t.Fatal("overflowing list should render a scrollbar thumb")
	}
	var cx, cy float64
	for _, b := range a.Page().Boxes() {
		if b.ID == "mail-list-viewport" {
			cx, cy = b.X+b.W/2, b.Y+b.H/2
		}
	}
	handled, err := a.handleWheel(ctx, cx, cy, 0, -1)
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("wheel over the list should be consumed")
	}
	if a.listOffset == 0 {
		t.Fatal("wheel down did not advance the list")
	}
	before := a.listOffset
	if _, err := a.handleWheel(ctx, cx, cy, 0, 2); err != nil {
		t.Fatal(err)
	}
	if a.listOffset >= before {
		t.Fatal("wheel up did not move the list back")
	}
}

func TestMailReaderWheelScrollsWhenOverflowing(t *testing.T) {
	t.Parallel()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	m := a.findMessage(a.selectedID)
	if m == nil {
		t.Fatal("no selected message")
	}
	m.Body = "Hi,\n\n" + strings.Repeat("A long paragraph keeps the reader tall enough to overflow. ", 40) + "\n\nBye"
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(a.Page().HTML(), "mail-reader-thumb") {
		t.Fatal("overflowing reader should render a scrollbar thumb")
	}
	var cx, cy float64
	for _, b := range a.Page().Boxes() {
		if b.ID == "mail-reader-viewport" {
			cx, cy = b.X+b.W/2, b.Y+b.H/2
		}
	}
	handled, err := a.handleWheel(ctx, cx, cy, 0, -2)
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("wheel over the reader should be consumed")
	}
	if a.readerScroll == 0 {
		t.Fatal("wheel down did not scroll the reader")
	}
}

func TestMailSendMovesDraftToSentWithToast(t *testing.T) {
	t.Parallel()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a.folder = "drafts"
	a.ensureSelection()
	draft := a.findMessage(a.selectedID)
	if draft == nil || !draft.Keywords["$draft"] {
		t.Fatalf("expected an open draft, selected %q", a.selectedID)
	}
	sentID := a.selectedID
	if err := a.handleClick(ctx, ownframe.Box{Action: "send"}); err != nil {
		t.Fatal(err)
	}
	if a.folder != "sent" {
		t.Fatalf("folder after send = %q, want sent", a.folder)
	}
	sent := a.findMessage(sentID)
	if sent == nil || !sent.MailboxIDs["sent"] || sent.MailboxIDs["drafts"] {
		t.Fatal("sent message should leave drafts for sent")
	}
	if sent.Keywords["$draft"] {
		t.Fatal("sent message should lose the draft keyword")
	}
	if a.toast == "" {
		t.Fatal("send should raise the toast")
	}
	html := a.Page().HTML()
	if !strings.Contains(html, "mail-toast") || !strings.Contains(html, "toast-close") {
		t.Fatal("toast card with a close control should render")
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "toast-close"}); err != nil {
		t.Fatal(err)
	}
	if a.toast != "" || strings.Contains(a.Page().HTML(), "mail-toast") {
		t.Fatal("toast close should dismiss the card immediately")
	}
}

func TestMailToastExpiresAfterTenSeconds(t *testing.T) {
	t.Parallel()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a.showToast("Message sent.")
	if time.Until(a.toastUntil) < 9*time.Second {
		t.Fatal("toast should linger for ten seconds")
	}
	a.toastUntil = time.Now().Add(-time.Second)
	if err := a.tick(ctx); err != nil {
		t.Fatal(err)
	}
	if a.toast != "" {
		t.Fatal("expired toast should clear on tick")
	}
}

func TestMailSearchAndEscaping(t *testing.T) {
	t.Parallel()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	a.query = "working group"
	if got := a.visibleMessages(); len(got) != 1 || got[0].ID != "m-102" {
		t.Fatalf("search results: got %#v", got)
	}
	a.query = "{{.Private}} <tag>"
	markup, err := ui.Render(a.screen())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(markup, "{{") || strings.Contains(markup, "<tag>") {
		t.Fatalf("unsafe search data in rendered source: %s", markup)
	}
}

func TestMailSettingsThemeAndPaneSpacing(t *testing.T) {
	t.Parallel()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	initial := a.Page().HTML()
	if !strings.Contains(initial, "gap:14px;") ||
		!strings.Contains(initial, "height:36px;") {
		t.Fatal("mail panes or search field missing layout sizing")
	}
	if err := pagetest.ClickAction(ctx, a.Page(), "settings"); err != nil {
		t.Fatal(err)
	}
	if !a.settings || !strings.Contains(a.Page().HTML(), "settings-card") {
		t.Fatal("settings button should open the settings card")
	}
	for _, word := range []string{"Settings", "Appearance", "theme:light", "theme:dark"} {
		if !strings.Contains(a.Page().HTML(), word) {
			t.Fatalf("settings card missing %q", word)
		}
	}
	if err := pagetest.ClickAction(ctx, a.Page(), "theme:dark"); err != nil {
		t.Fatal(err)
	}
	if a.appearance != mailThemeDark {
		t.Fatal("theme option did not change state")
	}
	if got := mailToolbarPixel(t, ctx, a.Page()); got != (color.RGBA{R: 28, G: 39, B: 56, A: 255}) {
		t.Fatalf("dark toolbar surface = %v", got)
	}
	if err := pagetest.ClickAction(ctx, a.Page(), "theme:light"); err != nil {
		t.Fatal(err)
	}
	if a.appearance != mailThemeLight || mailToolbarPixel(t, ctx, a.Page()) != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatal("light theme not restored")
	}
	if err := pagetest.Click(ctx, a.Page(), "settings-close"); err != nil {
		t.Fatal(err)
	}
	if a.settings || strings.Contains(a.Page().HTML(), "settings-card") {
		t.Fatal("card close button did not close the settings card")
	}
	if err := pagetest.ClickAction(ctx, a.Page(), "settings"); err != nil {
		t.Fatal(err)
	}
	if !a.settings {
		t.Fatal("settings button should reopen the card")
	}
	if err := a.Page().Click(ctx, 30, 400); err != nil {
		t.Fatal(err)
	}
	if a.settings || strings.Contains(a.Page().HTML(), "settings-card") {
		t.Fatal("backdrop click did not close the settings card")
	}
}

func mailToolbarPixel(t *testing.T, ctx context.Context, page *ownframe.Page) color.Color {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(page.Testing().PNGWithContext(ctx)))
	if err != nil {
		t.Fatal(err)
	}
	return color.RGBAModel.Convert(img.At(1, 1))
}

func TestMailCustomThemeAndModalLifecycle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.SetTheme(`:root{--of-surface:#123456;--of-text:#ffffff;--of-background:#102030}`); err != nil {
		t.Fatal(err)
	}
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if got := mailToolbarPixel(t, ctx, a.page); got != (color.RGBA{R: 18, G: 52, B: 86, A: 255}) {
		t.Fatalf("custom surface=%v", got)
	}
	a.page.Form().SetFormValue("mail-search", "Old query")
	if err := pagetest.ClickAction(ctx, a.page, "clear-search"); err != nil {
		t.Fatal(err)
	}
	if got := a.page.Form().FormValue("mail-search"); got != "" {
		t.Fatalf("search was not cleared: %q", got)
	}
	if err := pagetest.ClickAction(ctx, a.page, "settings"); err != nil {
		t.Fatal(err)
	}
	if got := a.page.Form().FocusID(); got != "settings-close" {
		t.Fatalf("modal focus=%q", got)
	}
	for _, want := range []string{"theme-light", "theme-dark", "theme-custom", "settings-close"} {
		if err := a.page.Form().FocusNext(ctx); err != nil {
			t.Fatal(err)
		}
		if got := a.page.Form().FocusID(); got != want {
			t.Fatalf("modal Tab=%q want %q", got, want)
		}
	}
	if err := a.page.Edit().KeyDown(ctx, "escape"); err != nil {
		t.Fatal(err)
	}
	if a.settings {
		t.Fatal("Escape did not close settings")
	}
	if got := mailToolbarPixel(t, ctx, a.page); got != (color.RGBA{R: 18, G: 52, B: 86, A: 255}) {
		t.Fatalf("updates lost custom theme: %v", got)
	}
}

func TestMailBuiltInAndCustomThemeSelection(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, a.page, "settings-open"); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name string
		want color.RGBA
	}{
		{"custom", color.RGBA{R: 255, G: 253, B: 246, A: 255}},
		{"dark", color.RGBA{R: 28, G: 39, B: 56, A: 255}},
		{"light", color.RGBA{R: 255, G: 255, B: 255, A: 255}},
	} {
		if err := pagetest.Click(ctx, a.page, "theme-"+tt.name); err != nil {
			t.Fatal(err)
		}
		if a.appearance != mailTheme(tt.name) {
			t.Fatalf("appearance=%s", a.appearance)
		}
		if got := mailToolbarPixel(t, ctx, a.page); got != tt.want {
			t.Fatalf("%s surface=%v", tt.name, got)
		}
		if !strings.Contains(a.page.HTML(), "--of-radius:18px;") {
			t.Fatal("theme switch lost Compose override")
		}
	}
	before := a.appearance
	a.selectTheme("unknown")
	if a.appearance != before {
		t.Fatal("invalid theme changed state")
	}
}
