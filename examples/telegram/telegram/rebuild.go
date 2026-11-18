package telegram

import "strings"

// rebuild recomputes the template fields from the app state and the current
// query. Call it after every change and before SetData.
func (a *App) rebuild() {
	a.view.Chats = a.filtered()
	a.view.Contacts = a.contacts
	a.view.Unread = unreadTotal(a.chats)
	a.view.Thread = a.threads[a.view.Active]

	if (a.web || a.view.Web) && a.view.Active != "" {
		full := a.threads[a.view.Active]
		if a.threadLimit <= 0 {
			a.threadLimit = webThreadPage
		}
		if len(full) > a.threadLimit {
			window := full[len(full)-a.threadLimit:]
			reversed := make([]Message, len(window))
			for i, m := range window {
				reversed[len(window)-1-i] = m
			}
			a.view.Thread = reversed
			a.view.ThreadMore = true
		} else {
			reversed := make([]Message, len(full))
			for i, m := range full {
				reversed[len(full)-1-i] = m
			}
			a.view.Thread = reversed
			a.view.ThreadMore = false
		}
	} else {
		a.view.ThreadMore = false
	}

	if i := a.chatIndex(a.view.Active); i >= 0 {
		c := a.chats[i]
		a.view.Header = c.Name
		a.view.Initials = c.Initials
		a.view.Color = c.Color
		a.view.Group = c.Group
	}

	a.view.CanGift = a.view.Active == giftChat
	a.onList.Store(a.view.Active == "")
	a.dark.Store(a.view.Dark)
}

// filtered returns the chats that match the search query, in list order.
func (a *App) filtered() []Chat {
	q := strings.ToLower(strings.TrimSpace(a.view.Query))
	out := make([]Chat, 0, len(a.chats))

	for _, c := range a.chats {
		if q != "" && !strings.Contains(strings.ToLower(c.Name), q) &&
			!strings.Contains(strings.ToLower(c.Preview), q) {
			continue
		}

		out = append(out, c)
	}

	return out
}

// unreadTotal is the badge on the Chats tab.
func unreadTotal(chats []Chat) int {
	n := 0

	for _, c := range chats {
		n += c.Unread
	}

	return n
}

// chatIndex finds a chat by id, or -1.
func (a *App) chatIndex(id string) int {
	for i := range a.chats {
		if a.chats[i].ID == id {
			return i
		}
	}

	return -1
}
