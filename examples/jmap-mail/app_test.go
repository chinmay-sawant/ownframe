package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	expectedUnreadInbox = 2
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
		t.Fatalf("unread inbox: got %d, want 2", got)
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
		t.Fatalf("messages after draft: got %d want 9", len(a.messages))
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
