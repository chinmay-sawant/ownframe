package ipc

import "context"

// Handle registers fn as the reply for channel.
// A later Handle on that channel replaces fn.
// Cancel removes fn only when it is still the handler.
// An empty channel does nothing.
//
// Handlers run synchronously in the requesting goroutine. They must not block
// indefinitely; there is no per-call timeout, so a blocking handler blocks
// Request. IPC is an in-process convenience API, not an isolation boundary.
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
