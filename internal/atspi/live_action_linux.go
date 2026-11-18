//go:build linux && !android && !ios && !js

package atspi

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/safego"
)

func isActionCall(msg *dbusMsg) bool {
	return (msg.iface == atspiComponent && msg.member == "GrabFocus") ||
		(msg.iface == atspiAction && msg.member == "DoAction")
}

// dispatchAction leaves the reader free to deliver publication replies while a
// page action waits for its loop. Both reply workers and page admission are bounded.
func (b *liveBus) dispatchAction(msg *dbusMsg, view *serverView, parent context.Context) {
	select {
	case b.actionSlots <- struct{}{}:
	default:
		b.reply(msg, okResult("b", encodeBoolValue(false)))
		return
	}
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		<-b.actionSlots
		return
	}
	b.wg.Add(1)
	b.mu.Unlock()
	safego.Go("atspi action", func() {
		defer b.wg.Done()
		defer func() { <-b.actionSlots }()
		ctx, cancel := context.WithTimeout(parent, liveCallTimeout)
		defer cancel()
		snapshot := *view
		snapshot.actions = b.actions
		b.reply(msg, snapshot.handle(ctx, msg))
	})
}
