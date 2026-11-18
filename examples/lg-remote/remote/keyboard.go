package remote

import (
	"context"
	"strconv"

	"github.com/chinmay-sawant/ownframe/examples/lg-remote/tv"
)

const (
	typingField  = "typing"
	focusRetries = 30
)

// openKeys shows the keyboard panel and focuses the field so the phone's
// keyboard opens with it.
func (a *App) openKeys(ctx context.Context) {
	a.show("keys", "")
	a.typed = ""
	a.page.Form().SetFormValue(typingField, "")
	// The click that opened this tab focuses its own button afterwards, so
	// focus the field on the next tick or the phone keyboard never opens.
	a.focusNext = typingField
	host := a.host
	a.later(func() { _, _ = a.use().Exec(host, "ime:watch:") })
}

// keyboardNote tells the user whether the TV is ready for text.
func (a *App) keyboardNote() string {
	if a.view.Mode == "Bluetooth" {
		return "Typing needs Wi-Fi."
	}

	reader, ok := a.use().(interface{ KeyboardState() tv.KeyboardState })
	state := tv.KeyboardUnknown
	if ok {
		state = reader.KeyboardState()
	}

	switch state {
	case tv.KeyboardOpen:
		return "The TV keyboard is open. Type here."
	case tv.KeyboardUnreported:
		return "Open a text box on the TV, then type here."
	default:
		return "No LG text box is open. Press OK on a search box to open the TV keyboard, then type here. Apps with their own keyboard (YouTube, Disney+) cannot take typing."
	}
}

// refreshKeys updates the note and moves focus once the click has settled.
func (a *App) refreshKeys(ctx context.Context) bool {
	changed := false
	if id := a.focusNext; id != "" {
		// The field may not be laid out on the first frame, and Focus ignores
		// an unknown id, so keep asking until the focus holds.
		_ = a.page.Form().Focus(ctx, id)
		a.focusTries++
		if a.page.Form().FocusID() == id || a.focusTries >= focusRetries {
			a.focusNext = ""
			a.focusTries = 0
		}
		changed = true
	}

	if note := a.keyboardNote(); note != a.view.KeyboardNote {
		a.view.KeyboardNote = note
		changed = true
	}

	return changed
}

// syncTyping sends what changed in the field to the TV. It reports whether
// the page needs a redraw.
func (a *App) syncTyping() bool {
	now := a.page.Form().FormValue(typingField)
	if now == a.typed {
		return false
	}

	old := []rune(a.typed)
	cur := []rune(now)
	a.typed = now

	if a.view.Mode == "Bluetooth" {
		a.view.Status = "Typing needs Wi-Fi"
		return true
	}

	keep := 0
	for keep < len(old) && keep < len(cur) && old[keep] == cur[keep] {
		keep++
	}

	host := a.host
	if gone := len(old) - keep; gone > 0 {
		a.later(func() { a.doExec(host, "ime:delete:"+strconv.Itoa(gone)) })
	}
	if add := string(cur[keep:]); add != "" {
		a.later(func() { a.doExec(host, "ime:text:"+add) })
	}

	return true
}

func (a *App) keysPress(id string) bool {
	host := a.host

	switch id {
	case "kb-delete":
		if a.view.Mode == "Bluetooth" {
			a.view.Status = "Typing needs Wi-Fi"
			return true
		}
		a.view.Status = "Sending"
		a.later(func() { a.doExec(host, "ime:delete:1") })
		runes := []rune(a.typed)
		if len(runes) > 0 {
			a.typed = string(runes[:len(runes)-1])
			a.page.Form().SetFormValue(typingField, a.typed)
		}
	case "kb-enter":
		if a.view.Mode == "Bluetooth" {
			a.view.Status = "Typing needs Wi-Fi"
			return true
		}
		a.view.Status = "Sending"
		a.later(func() { a.doExec(host, "ime:enter:") })
	case "kb-clear":
		if a.view.Mode == "Bluetooth" {
			a.view.Status = "Typing needs Wi-Fi"
			return true
		}
		a.later(func() { a.doExec(host, "ime:clear:") })
		a.typed = ""
		a.page.Form().SetFormValue(typingField, "")
		a.view.Status = "Sending"
	default:
		return false
	}

	return true
}
