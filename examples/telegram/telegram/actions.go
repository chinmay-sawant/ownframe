package telegram

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// onClick runs the control under the tap and rebuilds the view. A tap that
// is not a reaction closes an open reaction bar first.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	open := a.view.ReactID != ""
	if open && !strings.HasPrefix(box.Action, "react-") {
		a.view.ReactID = ""
	}

	switch {
	case box.Action == "tab-chats", box.Action == "tab-contacts", box.Action == "tab-settings":
		a.view.Tab = strings.TrimPrefix(box.Action, "tab-")
		a.page.ScrollTo(0, 0)
	case box.Action == "chat-back":
		a.view.Active = ""
		a.view.Status = ""
		a.view.AttachOpen = false
		a.page.ScrollTo(0, 0)
	case strings.HasPrefix(box.Action, "open-"):
		a.open(strings.TrimPrefix(box.Action, "open-"))
	case strings.HasPrefix(box.Action, "contact-"):
		a.openContact(strings.TrimPrefix(box.Action, "contact-"))
	case box.Action == "send":
		a.send()
		if err := a.refocusComposer(ctx); err != nil {
			return err
		}
	case box.Action == "gift":
		a.gift()
	case box.Action == "attach":
		a.view.AttachOpen = !a.view.AttachOpen
	case box.Action == "attach-camera":
		a.view.AttachOpen = false
		a.RequestAttach("camera")
	case box.Action == "attach-gallery":
		a.view.AttachOpen = false
		a.RequestAttach("gallery")
	case strings.HasPrefix(box.Action, "react-"):
		a.react(strings.TrimPrefix(box.Action, "react-"))
	default:
		if !open {
			return nil
		}
	}

	a.rebuild()
	a.page.SetData(&a.view)

	return nil
}
