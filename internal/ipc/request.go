package ipc

import (
	"context"
	"errors"
)

// ErrHandlerPanicked is returned by Request when the registered handler
// panics. Its panic is recovered so it cannot propagate to the caller.
var ErrHandlerPanicked = errors.New("ownframe: handler panicked")

// Aliases for ErrHandlerPanicked; use ErrHandlerPanicked in new code.
var (
	// Deprecated: use ErrHandlerPanicked.
	ErrHandlerPanic = ErrHandlerPanicked
	// Deprecated: use ErrHandlerPanicked.
	ErrPanicked = ErrHandlerPanicked
	// Deprecated: use ErrHandlerPanicked.
	ErrPanic = ErrHandlerPanicked
	// Deprecated: use ErrHandlerPanicked.
	ErrHandlePanicked = ErrHandlerPanicked
	// Deprecated: use ErrHandlerPanicked.
	ErrHandlePanic = ErrHandlerPanicked
	// Deprecated: use ErrHandlerPanicked.
	ErrRequestPanicked = ErrHandlerPanicked
	// Deprecated: use ErrHandlerPanicked.
	ErrRequestPanic = ErrHandlerPanicked
)

// Request calls the Handle callback and returns its reply.
// An empty channel, or no handler, returns ErrNoHandler.
// A nil or canceled ctx does not call the handler.
//
// A panicking handler is recovered and reported as ErrHandlerPanicked. Serve
// has no equivalent recovery; only this handler call is guarded.
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
