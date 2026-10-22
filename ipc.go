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
//
// Channels share one process-global namespace with no authentication:
// prefix names per owner (for example "player:play"), keep secrets out
// of payloads, and never bridge a channel to Serve or the notify port
// without a bearer token plus an allowlist.
func Send(channel, payload string) {
	ipc.Send(channel, payload)
}

// Listen registers fn for channel. Prefix the channel with the owner
// name so two packages cannot collide on one victim channel.
// Cancel removes that listener only. An empty channel does nothing.
func Listen(channel string, fn func(string)) (cancel func()) {
	return ipc.Listen(channel, fn)
}

// Handle registers fn as the reply for channel. Prefix the channel with
// the owner name: a later Handle replaces fn, so an unprefixed channel
// lets any linked code steal the reply.
// Cancel removes fn only if it is still current.
func Handle(
	channel string,
	fn func(context.Context, string) (string, error),
) (cancel func()) {
	return ipc.Handle(channel, fn)
}

// Request calls the Handle callback and returns its reply.
// An empty channel, or no handler, returns ErrNoHandler.
// A nil or canceled ctx does not call the handler.
// Keep secrets out of channel payloads; the bus has no auth boundary.
func Request(ctx context.Context, channel, payload string) (string, error) {
	return ipc.Request(ctx, channel, payload)
}
