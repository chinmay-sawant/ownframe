package page

import "sync/atomic"

// frameState is the page's per-frame opt-ins: the dirty-rect opt-in the
// window reads, and the next-frame deadline a page sets to pace an idle
// window. Both clear when the window reads them. Page embeds this state to
// group the atomic handoff fields.
// Writer: the page loop sets, the window take-clears (page_tick.go).
// Lock: none, sync/atomic carries the handoff; no mutex needed.
//
// frameAt is stored as Unix nanoseconds and accessed with sync/atomic so a
// frame deadline can be handed off between the tick, handlers, and the game
// loop without a lock.
type frameState struct {
	frameDirty atomic.Bool
	frameAt    int64 // atomic: Unix nanoseconds
}
