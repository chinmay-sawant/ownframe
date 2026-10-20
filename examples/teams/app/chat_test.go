package app

import (
	"context"
	"testing"
)

// newChatApp opens the app and draws the chat menu.
func newChatApp(t *testing.T) *App {
	t.Helper()

	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := app.Redraw(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	return app
}

// clickChat clicks the control with the id.
func clickChat(t *testing.T, app *App, id string) {
	box, ok := findBox(app.Boxes(), id)
	if !ok {
		t.Fatalf("no %s box", id)
	}

	clickBox(t, app, box)
}

func TestChatDefault(t *testing.T) {
	t.Parallel()
	app := newChatApp(t)

	c := app.View().Chat
	if app.View().Section != "chat" || c.Active != c.Chats[0].ID || c.Header != c.Chats[0].Name {
		t.Fatalf("Section=%q Active=%q Header=%q", app.View().Section, c.Active, c.Header)
	}

	if len(c.Messages) < 6 {
		t.Fatalf("Messages = %d", len(c.Messages))
	}
}

func TestChatOpenSwitchesThread(t *testing.T) {
	t.Parallel()
	app := newChatApp(t)
	clickChat(t, app, "chat-item-rhodey")

	c := app.View().Chat
	if c.Active != "rhodey" || c.Header != "James Rhodes" || len(c.Messages) == 0 {
		t.Fatalf("Active=%q Header=%q Messages=%d", c.Active, c.Header, len(c.Messages))
	}
}

func TestChatSendAppendsMessage(t *testing.T) {
	t.Parallel()
	app := newChatApp(t)

	before := len(app.View().Chat.Messages)
	app.Page().SetFormValue("chat-compose", "  Hello team  ")
	clickChat(t, app, "chat-send")

	c := app.View().Chat
	m := c.Messages[before]
	if len(c.Messages) != before+1 || !m.Own || m.Author != "Robert Downey Jr." || m.Text != "Hello team" {
		t.Fatalf("message = %+v", m)
	}

	if app.Page().FormValue("chat-compose") != "" {
		t.Fatal("compose input not cleared")
	}
}

func TestChatPinToggles(t *testing.T) {
	t.Parallel()
	app := newChatApp(t)

	if !app.View().Chat.Pinned() {
		t.Fatal("pepper should start pinned")
	}

	clickChat(t, app, "chat-pin")

	if app.View().Chat.Pinned() {
		t.Fatal("still pinned after the click")
	}
}
