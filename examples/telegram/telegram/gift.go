package telegram

import "strconv"

// giftChat is the only chat that allows sending a gift.
const giftChat = "anna"

// gift appends a gift bubble to the open chat.
func (a *App) gift() {
	id := a.view.Active
	if id != giftChat {
		return
	}

	a.threads[id] = append(a.threads[id], Message{
		ID:   "gift-" + strconv.Itoa(len(a.threads[id])),
		Time: "now", Own: true, Read: true, Gift: true,
	})

	if i := a.chatIndex(id); i >= 0 {
		a.chats[i].Preview = "You: a gift"
		a.chats[i].Time = "now"
	}

	a.page.ScrollTo(0, scrollBottom)
}
