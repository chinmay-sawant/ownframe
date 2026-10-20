package app

import (
	"context"
	"testing"
)

// TestChatEnterSends checks Enter in the composer sends the message.
func TestChatEnterSends(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)
	before := len(app.View().Chat.Messages)

	if err := app.Page().Focus(context.Background(), "chat-compose"); err != nil {
		t.Fatalf("Focus: %v", err)
	}

	app.Page().SetFormValue("chat-compose", "Avengers assemble")

	if err := app.Page().Submit(context.Background()); err != nil {
		t.Fatalf("Submit: %v", err)
	}

	c := app.View().Chat
	if len(c.Messages) != before+1 {
		t.Fatalf("messages = %d, want %d", len(c.Messages), before+1)
	}

	last := c.Messages[before]
	if !last.Own || last.Text != "Avengers assemble" || last.Author != "Robert Downey Jr." {
		t.Fatalf("last = %+v", last)
	}

	if app.Page().FormValue("chat-compose") != "" {
		t.Fatal("compose input not cleared")
	}
}
