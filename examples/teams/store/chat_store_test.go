package store

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/teams/chat"
)

// TestChatRoundTrip checks sent messages, the open chat, and the pin survive.
func TestChatRoundTrip(t *testing.T) {
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}

	st, err := Open(Memory)
	must(err)
	defer func() { _ = st.Close() }()

	must(st.Seed())

	d, ok, err := st.Load()
	must(err)

	if !ok {
		t.Fatal("Load: not seeded")
	}

	page, err := ownframe.New(ownframe.Config{HTML: `<input id="chat-compose">`, Width: 800, Height: 600})
	must(err)

	ctx := context.Background()
	must(page.Redraw(ctx))

	pbefore := len(d.Chat.AllThreads()["pepper"])
	before := len(d.Chat.AllThreads()["peter"])

	chat.Handle(ctx, page, &d.Chat, "chat-open-peter")
	page.SetFormValue("chat-compose", "First message")
	chat.Handle(ctx, page, &d.Chat, "chat-send")
	chat.Handle(ctx, page, &d.Chat, "chat-open-pepper")
	page.SetFormValue("chat-compose", "Second message")
	chat.Handle(ctx, page, &d.Chat, "chat-send")
	page.SetFormValue("chat-compose", "Third message")
	chat.Handle(ctx, page, &d.Chat, "chat-send")
	chat.Handle(ctx, page, &d.Chat, "chat-open-peter")
	chat.Handle(ctx, page, &d.Chat, "chat-pin")

	must(st.Save(d))

	c, ok, err := st.Load()
	must(err)

	if !ok {
		t.Fatal("Load: not seeded")
	}

	if c.Chat.Active != "peter" || len(c.Chat.Messages) != before+1 || !c.Chat.Pinned() {
		t.Fatalf("Active=%q Messages=%d", c.Chat.Active, len(c.Chat.Messages))
	}

	if last := c.Chat.Messages[before]; !last.Own || last.Text != "First message" {
		t.Fatalf("last = %+v", last)
	}

	pepper := c.Chat.AllThreads()["pepper"]

	if len(pepper) != pbefore+2 || pepper[pbefore+1].Text != "Third message" {
		t.Fatalf("pepper = %+v", pepper)
	}

	if peter := c.Chat.AllChats()[1]; peter.Preview != "First message" || peter.Unread != 0 {
		t.Fatalf("peter row = %+v", peter)
	}
}
