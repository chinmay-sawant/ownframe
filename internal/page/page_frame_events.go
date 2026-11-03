package page

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

var _ host.FrameScheduler = (*Page)(nil)

// FrameEvents coalesces notifications without blocking a producer. The host
// consumes these events; no page data is read on its timer goroutine.
func (p *Page) FrameEvents() <-chan struct{} {
	p.frameOnce.Do(func() { p.frameEvents = make(chan struct{}, 1) })
	return p.frameEvents
}

func (p *Page) notifyFrame() {
	if p.tickRunning.Load() {
		return
	}
	p.FrameEvents()
	select {
	case p.frameEvents <- struct{}{}:
	default:
	}
}

// FramePolling preserves callbacks that discover changes by polling and the
// existing file reload contract. Static pages need neither keep-alive.
func (p *Page) FramePolling() bool { return p.hasFn.Load() || p.Watching() }

func (p *Page) caretDeadline() time.Time {
	c, ok := p.focusedEditable()
	fs := p.formState()
	if !ok || fs == nil || form.IsFileInput(c) {
		return time.Time{}
	}
	start, end := form.RangeOf(fs.Anchor, fs.Caret, runeLen(c.Value))
	if fs.All || start != end {
		return time.Time{}
	}
	if fs.BlinkAt.IsZero() {
		return time.Now().Add(form.BlinkInterval)
	}
	return fs.BlinkAt.Add(form.BlinkInterval)
}
