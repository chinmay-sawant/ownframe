package telegram

import "strconv"
import "strings"

// open shows the thread for a chat, clears its unread count, and asks the
// window to jump to the newest message.
func (a *App) open(id string) {
	i := a.chatIndex(id)
	if i < 0 {
		return
	}

	a.chats[i].Unread = 0
	a.view.Active = id
	a.view.Status = "online"
	a.view.AttachOpen = false

	if a.chats[i].Group {
		a.view.Status = "5 members"
	}

	a.page.ScrollTo(0, scrollBottom)
}

// send appends the draft as an own message and updates the list row.
func (a *App) send() {
	text := strings.TrimSpace(a.view.Draft)
	if text == "" || a.view.Active == "" {
		return
	}

	id := a.view.Active
	a.threads[id] = append(a.threads[id], Message{
		ID:   "own-" + strconv.Itoa(len(a.threads[id])),
		Text: text, Time: "now", Own: true, Read: true,
	})
	a.view.Draft = ""

	if i := a.chatIndex(id); i >= 0 {
		a.chats[i].Preview = "You: " + text
		a.chats[i].Time = "now"
	}

	a.page.ScrollTo(0, scrollBottom)
}

// openContact opens the contact's chat. A contact with no thread gets a new,
// empty chat at the end of the list, remembered on the contact.
func (a *App) openContact(id string) {
	for i := range a.contacts {
		c := &a.contacts[i]
		if c.ID != id {
			continue
		}

		if c.ChatID != "" && a.chatIndex(c.ChatID) >= 0 {
			a.open(c.ChatID)

			return
		}

		c.ChatID = c.ID
		a.chats = append(a.chats, Chat{
			ID: c.ID, Name: c.Name, Initials: c.Initials, Color: c.Color,
			Preview: "No messages yet", Time: "now",
		})
		a.open(c.ID)

		return
	}
}
