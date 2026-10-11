package page

import "sync/atomic"

// CancelTouches queues gesture cancellation for the window loop. A native
// host may call it from its UI thread before clearing platform touch IDs.
func (p *Page) CancelTouches() { atomic.StoreUint32(&p.touchCancel, 1) }

// TakeTouchCancel consumes a queued cancellation. Custom screens can expose
// this method to let the window drop active gestures without delivering a tap.
func (p *Page) TakeTouchCancel() bool { return atomic.SwapUint32(&p.touchCancel, 0) != 0 }
