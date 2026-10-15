package ownframe

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/ipc"
)

// ErrNoHandler means Request found no Handle for that channel.
var ErrNoHandler = ipc.ErrNoHandler

// ErrHandlerPanicked means a handler panicked during Request; it is recovered.
var ErrHandlerPanicked = ipc.ErrHandlerPanicked

// Send calls each Listen callback for channel in this process.
// A callback that panics does not stop the others.
// An empty channel does nothing.
func Send(channel, payload string) {
	ipc.Send(channel, payload)
}

// Listen registers fn for channel.
// Cancel removes that listener only. An empty channel does nothing.
func Listen(channel string, fn func(string)) (cancel func()) {
	return ipc.Listen(channel, fn)
}

// Handle registers fn as the reply for channel.
// A later Handle replaces fn. Cancel removes fn only if it is still current.
func Handle(
	channel string,
	fn func(context.Context, string) (string, error),
) (cancel func()) {
	return ipc.Handle(channel, fn)
}

// Request calls the Handle callback and returns its reply.
// An empty channel, or no handler, returns ErrNoHandler.
// A nil or canceled ctx does not call the handler.
func Request(ctx context.Context, channel, payload string) (string, error) {
	return ipc.Request(ctx, channel, payload)
}
