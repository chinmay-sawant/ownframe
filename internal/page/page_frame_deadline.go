package page

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

var _ host.FrameWanter = (*Page)(nil)

// WantFrameAt declares the next time the page needs the window to draw. The
// desktop window reads it once per frame, before it draws, and schedules its
// idle wake for that time; a time at or before now asks for the next frame at
// once and keeps a per-frame animation on the vsync loop. The window clears
// the request when it reads it, like the FrameDirty opt-in, so a tick that
// wants another frame must ask again on every frame. The zero time means the
// page does not ask: the window falls back to its keep-alive pacing. Serve
// and the mobile views do not read it. WantFrameAt may be called from the
// tick or from an event handler.
func (p *Page) WantFrameAt(at time.Time) { p.frameAt = at }

// WantFrameIn declares the next frame in d. d <= 0 asks for the next frame at
// once. It is WantFrameAt(time.Now().Add(d)) against the window's clock.
func (p *Page) WantFrameIn(d time.Duration) { p.WantFrameAt(time.Now().Add(d)) }

// NextFrameAt returns the deadline set by WantFrameAt or WantFrameIn and
// clears it. It reports the zero time when the page did not ask. The window
// calls it once per frame; the page does not.
func (p *Page) NextFrameAt() time.Time {
	at := p.frameAt
	p.frameAt = time.Time{}

	return at
}
