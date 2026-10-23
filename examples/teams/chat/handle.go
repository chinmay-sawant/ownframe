package chat

import (
	"context"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// Handle applies one chat action and reports whether it was ours.
func Handle(_ context.Context, page *ownframe.Page, d *Data, action string) bool {
	switch {
	case action == "chat-send":
		send(page, d)
	case action == "chat-query":
		d.Query = page.Form().FormValue("chat-search")
		d.rebuild()
	case action == "chat-pin":
		toggle(d, true)
	case action == "chat-mute":
		toggle(d, false)
	case action == "chat-filter-all":
		applyFilter(d, "all")
	case action == "chat-filter-unread":
		applyFilter(d, "unread")
	case action == "chat-filter-groups":
		applyFilter(d, "groups")
	case strings.HasPrefix(action, "chat-open-"):
		open(d, strings.TrimPrefix(action, "chat-open-"))
	default:
		return false
	}

	return true
}

// send appends the compose text as an own message on the open chat.
func send(page *ownframe.Page, d *Data) {
	text := strings.TrimSpace(page.Form().FormValue("chat-compose"))
	if text == "" || d.Active == "" {
		return
	}

	d.Messages = append(d.Messages, Message{
		ID:       "sent-" + d.Active + "-" + strconv.Itoa(len(d.Messages)),
		Author:   "Robert Downey Jr.",
		Initials: "RDJ",
		Color:    "red",
		Time:     "now",
		Text:     text,
		Own:      true,
	})

	if d.threads == nil {
		d.threads = map[string][]Message{}
	}

	d.threads[d.Active] = d.Messages

	for i := range d.Chats {
		if d.Chats[i].ID == d.Active {
			d.Chats[i].Preview = text
			d.Chats[i].Time = "Now"

			break
		}
	}

	if c := d.stored(d.Active); c != nil {
		c.Preview = text
		c.Time = "Now"
	}

	page.Form().SetFormValue("chat-compose", "")
}
