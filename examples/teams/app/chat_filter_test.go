package app

import (
	"context"
	"strings"
	"testing"
)

// TestChatFilterNarrowsData checks the filter chips change the real list.
func TestChatFilterNarrowsData(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)
	all := len(app.View().Chat.Chats)

	clickID(t, app, "chat-filter-unread")

	unread := app.View().Chat.Chats
	if len(unread) == 0 || len(unread) >= all {
		t.Fatalf("unread list = %d of %d", len(unread), all)
	}

	for _, c := range unread {
		if c.Unread == 0 {
			t.Fatalf("%s has no unread count", c.Name)
		}
	}

	clickID(t, app, "chat-filter-groups")

	for _, c := range app.View().Chat.Chats {
		if !c.Group {
			t.Fatalf("%s is not a group", c.Name)
		}
	}

	clickID(t, app, "chat-filter-all")

	if got := len(app.View().Chat.Chats); got != all {
		t.Fatalf("all list = %d, want %d", got, all)
	}
}

// TestChatSearchNarrowsData checks typing in the search box filters the list.
func TestChatSearchNarrowsData(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)
	all := len(app.View().Chat.Chats)

	clickID(t, app, "chat-search")

	if err := app.Page().Type(context.Background(), "peter"); err != nil {
		t.Fatalf("Type: %v", err)
	}

	chats := app.View().Chat.Chats
	if len(chats) == 0 || len(chats) >= all {
		t.Fatalf("search list = %d of %d", len(chats), all)
	}

	for _, c := range chats {
		text := strings.ToLower(c.Name + " " + c.Preview)
		if !strings.Contains(text, "peter") {
			t.Fatalf("%s does not match the query", c.Name)
		}
	}
}
