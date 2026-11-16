package main

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// A live session makes one tab react to the pointer, the keys and the clock.
// The page tick calls Frame once per frame with one liveInput; the session
// patches the page through the OwnFrame element API. A session acts only
// while its own tab is the open document and sits idle otherwise, so any tab
// can plug one in without the others noticing.

// liveInput is one frame of input in page coordinates (the window cursor plus
// the scroll offset).
type liveInput struct {
	CursorX, CursorY float64
	// Click is true on the frame the left button went down; Down while it is
	// held.
	Click, Down bool
	// Keys are the non-text keys that went down this frame, named as ebiten
	// names them: ArrowUp, Backspace, Enter.
	Keys []string
	// Text is what the person typed this frame.
	Text string
}

// liveSession is one tab's behavior.
type liveSession interface {
	// TabID is the tab the session belongs to.
	TabID() string
	// Frame runs one frame. It returns when it next wants a frame (zero for
	// never).
	Frame(ctx context.Context, page *ownframe.Page, in liveInput, now time.Time) (time.Time, error)
}

// liveClock is the time source. The window uses the wall clock; tests and the
// scripted demo use a manual one.
type liveClock interface{ Now() time.Time }

type wallClock struct{}

func (wallClock) Now() time.Time { return time.Now() }

// manualLiveClock is a clock the caller moves.
type manualLiveClock struct{ t time.Time }

func (c *manualLiveClock) Now() time.Time { return c.t }

// advance moves the clock forward.
func (c *manualLiveClock) advance(d time.Duration) { c.t = c.t.Add(d) }

// liveSource reads one frame of input. scroll is the page scroll offset in
// CSS pixels.
type liveSource struct {
	cursor func() (x, y float64)
	click  func() bool
	down   func() bool
	keys   func() []string
	text   func() string
}

// liveRunner owns the sessions of one page and drives them from its tick.
type liveRunner struct {
	page     *ownframe.Page
	sessions []liveSession
	clock    liveClock
	src      liveSource
	// reportErr gets the first refused patch. The window keeps running after
	// it, because an error from a tick ends the loop.
	reportErr func(error)
	reported  bool
	// beforeFrame runs at the start of every tick (the scripted demo moves
	// its script here).
	beforeFrame func()
}

// attach registers the per-frame callback on the page.
func (r *liveRunner) attach() {
	r.page.SetTick(func(ctx context.Context) error {
		if r.beforeFrame != nil {
			r.beforeFrame()
		}

		return r.tick(ctx)
	})
}

// tick reads the input once and offers it to every session.
func (r *liveRunner) tick(ctx context.Context) error {
	_, sy := r.page.Scroll().ScrollOffset()
	cx, cy := r.src.cursor()
	in := liveInput{CursorX: cx, CursorY: cy + float64(sy), Click: r.src.click(), Down: r.src.down(), Keys: r.src.keys(), Text: r.src.text()}
	now := r.clock.Now()

	for _, s := range r.sessions {
		next, err := s.Frame(ctx, r.page, in, now)
		if err != nil {
			if r.reportErr == nil {
				return err
			}

			if !r.reported {
				r.reported = true
				r.reportErr(err)
			}
		}

		if !next.IsZero() {
			r.page.WantFrameAt(next)
		}
	}

	return nil
}

// liveSessions lists the tabs that have a live session.
func liveSessions() []liveSession {
	return []liveSession{newNavSession()}
}
