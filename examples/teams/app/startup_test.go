package app

import (
	"context"
	"path/filepath"
	"testing"
)

// TestStartupMatchesTheSavedState checks a reopen on the same database shows
// the saved theme and the whole chat list: the profile toggle works and a
// search box left over from a previous visit cannot hide every chat.
func TestStartupMatchesTheSavedState(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "teams.db")
	ctx := context.Background()

	first, err := New(WithDB(path))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := first.Redraw(ctx); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	clickID(t, first, "chat-search")

	if err := first.Page().Edit().Type(ctx, "nothing matches this"); err != nil {
		t.Fatalf("Type: %v", err)
	}

	if len(first.View().Chat.Chats) != 0 {
		t.Fatalf("the query kept %d chats", len(first.View().Chat.Chats))
	}

	clickID(t, first, "me")
	clickID(t, first, "theme-light")

	if first.View().Dark {
		t.Fatal("the light theme did not apply")
	}

	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	second, err := New(WithDB(path))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	defer func() { _ = second.Close() }()

	if second.View().Dark {
		t.Fatal("the saved light state started dark")
	}

	if second.View().Chat.Query != "" {
		t.Fatalf("query = %q, want empty", second.View().Chat.Query)
	}

	if len(second.View().Chat.Chats) == 0 {
		t.Fatal("the saved search query came back and hid every chat")
	}

	clickID(t, second, "me")
	clickID(t, second, "theme-dark")

	if !second.View().Dark {
		t.Fatal("the dark theme did not apply from a light start")
	}
}
