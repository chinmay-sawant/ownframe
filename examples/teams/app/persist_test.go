package app

import (
	"context"
	"path/filepath"
	"testing"
)

// TestPersistAcrossReopen checks a sent message survives closing the app
// and opening it again on the same database file.
func TestPersistAcrossReopen(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "teams.db")

	first, err := New(WithDB(path))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := first.Redraw(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	if err := first.Page().Focus(context.Background(), "chat-compose"); err != nil {
		t.Fatalf("Focus: %v", err)
	}

	first.Page().SetFormValue("chat-compose", "Persisted across restart")

	if err := first.Page().Submit(context.Background()); err != nil {
		t.Fatalf("Submit: %v", err)
	}

	sent := false

	for _, m := range first.View().Chat.Messages {
		if m.Text == "Persisted across restart" {
			sent = true
		}
	}

	if !sent {
		t.Fatal("the message was not appended before the save")
	}

	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	second, err := New(WithDB(path))
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}

	defer second.Close()

	found := false

	for _, m := range second.View().Chat.Messages {
		if m.Text == "Persisted across restart" {
			found = true
		}
	}

	if !found {
		t.Fatal("the sent message did not survive the reopen")
	}
}
