package gesture

import "sync/atomic"

const (
	touchCancelQueued uint32 = 1
	touchCancelClear  uint32 = 0
)

// QueueCancel stores a gesture cancellation for the window loop.
// A native host may call it from its UI thread.
func QueueCancel(flag *uint32) {
	atomic.StoreUint32(flag, touchCancelQueued)
}

// TakeCancel consumes a queued cancellation.
func TakeCancel(flag *uint32) bool {
	return atomic.SwapUint32(flag, touchCancelClear) != touchCancelClear
}

// TouchZoom reports whether two-finger touch may zoom.
// A nil page allows zoom, which the caller passes as missing.
func TouchZoom(missing, allowed bool) bool {
	return missing || allowed
}
