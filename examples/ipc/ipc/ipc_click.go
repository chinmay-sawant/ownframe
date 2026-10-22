package ipc

import (
	"context"
	"errors"

	"github.com/chinmay-sawant/ownframe"
)

// onClick runs the action for the clicked control.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	switch box.ID {
	case "send":
		a.onSend()
	case "request":
		a.onRequest(ctx)
	case "missing":
		a.onMissing(ctx)
	case "cancel":
		a.onToggle()
	}

	return nil
}

// onSend fires one payload at demo.log and expects no reply.
func (a *App) onSend() {
	payload := a.page.Form().FormValue("payload")
	ownframe.Send("demo.log", payload)

	if a.view.Live {
		a.say(false, "Send demo.log "+q(payload)+" -> 2 listeners ran, no reply")
		return
	}

	a.say(false, "Send demo.log "+q(payload)+" -> no listeners, dropped")
}

// onRequest asks demo.double for twice the number in the input.
func (a *App) onRequest(ctx context.Context) {
	payload := a.page.Form().FormValue("num")
	reply, err := ownframe.Request(ctx, "demo.double", payload)
	if err != nil {
		a.say(true, "Request demo.double "+q(payload)+" -> error: "+err.Error())
		return
	}

	a.say(false, "Request demo.double "+q(payload)+" -> "+q(reply))
}

// onMissing asks a channel that has no handler, so Request fails closed.
func (a *App) onMissing(ctx context.Context) {
	_, err := ownframe.Request(ctx, "demo.none", "x")
	if errors.Is(err, ownframe.ErrNoHandler) {
		a.say(true, "Request demo.none -> ErrNoHandler")
		return
	}

	a.say(true, "Request demo.none -> unexpected reply")
}

// onToggle cancels the registrations, or adds them again.
func (a *App) onToggle() {
	if a.view.Live {
		a.Cancel()
		a.say(false, "Cancel -> listeners and handler removed")
		return
	}

	a.register()
	a.say(false, "Register -> listeners and handler added again")
}

// q quotes a payload so the tape shows which part of a line is data.
func q(s string) string {
	return `"` + s + `"`
}
