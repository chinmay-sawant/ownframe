package app

import (
	"testing"
)

// TestChatPeterThread checks the Spider-Man chat opens and reads as a
// two-person discussion with Iron Man about the web shooter field test.
func TestChatPeterThread(t *testing.T) {
	t.Parallel()
	app := newChatApp(t)
	clickChat(t, app, "chat-item-peter")

	c := app.View().Chat
	if c.Active != "peter" || c.Header != "Peter Parker" {
		t.Fatalf("Active=%q Header=%q", c.Active, c.Header)
	}

	if len(c.Messages) < 6 {
		t.Fatalf("Messages = %d", len(c.Messages))
	}

	own := false

	for _, m := range c.Messages {
		if m.Author != "Peter Parker" && m.Author != "Robert Downey Jr." {
			t.Fatalf("unexpected author = %q", m.Author)
		}

		if m.Own && m.Author != "Robert Downey Jr." {
			t.Fatalf("own author = %q", m.Author)
		}

		if m.Own {
			own = true
		}
	}

	if !own {
		t.Fatal("no own message in the peter thread")
	}
}
