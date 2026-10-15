package ipc

import (
	"context"
	"errors"
)

// ErrHandlerPanicked is returned by Request when the registered handler
// panics. Its panic is recovered so it cannot propagate to the caller.
var ErrHandlerPanicked = errors.New("ownframe: handler panicked")

// Aliases for ErrHandlerPanicked.
var (
	ErrHandlerPanic    = ErrHandlerPanicked
	ErrPanicked        = ErrHandlerPanicked
	ErrPanic           = ErrHandlerPanicked
	ErrHandlePanicked  = ErrHandlerPanicked
	ErrHandlePanic     = ErrHandlerPanicked
	ErrRequestPanicked = ErrHandlerPanicked
	ErrRequestPanic    = ErrHandlerPanicked
)

// Handle registers fn as the reply for channel.
// A later Handle on that channel replaces fn.
// Cancel removes fn only when it is still the handler.
// An empty channel does nothing.
func Handle(
	channel string,
	fn func(context.Context, string) (string, error),
) (cancel func()) {
	if channel == "" || fn == nil {
		return noop
	}

	mu.Lock()
	defer mu.Unlock()

	seq++
	id := seq
	asks[channel] = ask{id: id, fn: fn}

	return func() { removeAsk(channel, id) }
}

// removeAsk drops the handler only when id is still current.
func removeAsk(channel string, id int64) {
	mu.Lock()
	defer mu.Unlock()

	item, ok := asks[channel]
	if ok && item.id == id {
		delete(asks, channel)
	}
}

// Request calls the Handle callback and returns its reply.
// An empty channel, or no handler, returns ErrNoHandler.
// A nil or canceled ctx does not call the handler.
func Request(ctx context.Context, channel, payload string) (string, error) {
	if channel == "" {
		return "", ErrNoHandler
	}

	if ctx == nil {
		return "", errNilContext
	}

	if err := ctx.Err(); err != nil {
		return "", err
	}

	mu.Lock()
	item, ok := asks[channel]
	mu.Unlock()

	if !ok || item.fn == nil {
		return "", ErrNoHandler
	}

	wrap := func() (reply string, err error) {
		defer func() {
			if r := recover(); r != nil {
				reply = ""
				err = ErrHandlerPanicked
			}
		}()
		return item.fn(ctx, payload)
	}
	return wrap()
}
