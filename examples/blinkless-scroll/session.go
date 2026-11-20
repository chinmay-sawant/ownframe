package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// innerScrollStep is how far one J or K press moves a scroll container, and
// wheelStep is how far one wheel notch does. wheelStep matches the 48 pixels
// the window moves the page for the same notch.
const (
	innerScrollStep = 40
	wheelStep       = 48
)

// wheel offers one wheel event to the scroll container under the pointer. dy is
// the wheel delta the window reports, positive when the wheel rolls up. It
// reports whether a container took the event. A container with no room left in
// that direction hands the event back, so the page scrolls on, as a browser
// chains scrolling.
func (s *session) wheel(px, py, dy float64) bool {
	id, ok := s.host.ScrollerAt(px, py)
	if !ok {
		return false
	}

	return s.host.ScrollBy(id, -dy*wheelStep)
}

// session ties one page to one host. frame is the single entry point for
// input: window.go calls it with what ebiten reports, and the demo calls it
// with scripted values. Nothing else moves the host.
type session struct {
	page    *ownframe.Page
	host    *Host
	applier *Applier
	html    string
	width   int
	height  int
	// reportErr gets the first patch error. The window keeps running after a
	// refused patch instead of closing, because a tick error ends the loop.
	reportErr func(error)
	reported  bool
	last      FrameResult // result of the latest tick, read by the demo
	// beforeFrame, when set, runs at the start of every tick, before the
	// frame reads input. The windowed demo uses it to advance its script.
	beforeFrame func()
}

// FrameInput is one frame of input in the shape the window sees it: the
// scroll offset the page reports, the cursor in window coordinates, whether
// the left button went down this frame, and the keys that went down.
type FrameInput struct {
	ScrollY float64
	CursorX float64
	CursorY float64
	Click   bool
	Keys    []string
}

// FrameResult is what a frame decided, for the window to ignore and the demo
// to log.
type FrameResult struct {
	Frame Frame
	// Hit is the id of the box a click landed on, "" when it hit nothing.
	// It is only set on a frame with a click.
	Hit      string
	Clicked  bool
	Events   []string // event triggers that changed, by owner id and name
	Deadline time.Time
}

func (s *session) viewport(scrollY float64) Viewport {
	w, h := s.page.Size()
	content := float64(h)

	if d := s.page.Probe().Display(); d != nil {
		content = float64(d.Height)
	}

	max := content - float64(h)
	if max < 0 {
		max = 0
	}

	if scrollY > max {
		scrollY = max
	}

	if scrollY < 0 {
		scrollY = 0
	}

	return Viewport{ScrollY: scrollY, W: float64(w), H: float64(h), ContentW: float64(w), ContentH: content}
}

// syncGeometry reloads the box rectangles when the page size changed, because
// a resize lays the page out again and moves every box.
func (s *session) syncGeometry() {
	w, h := s.page.Size()
	if w == s.width && h == s.height {
		return
	}

	s.width, s.height = w, h

	if d := s.page.Probe().Display(); d != nil && len(d.Boxes) == len(s.host.Nodes) {
		for i, n := range Build(d).Nodes {
			s.host.Nodes[i].Rect = n.Rect
		}
	}
}

// frame runs one frame of input through the host and patches the page.
func (s *session) frame(ctx context.Context, in FrameInput) (FrameResult, error) {
	s.syncGeometry()

	scroll := s.viewport(in.ScrollY).ScrollY
	px, py := in.CursorX, in.CursorY+scroll // page coordinates

	var res FrameResult

	s.host.Hover(px, py)

	if in.Click {
		res.Clicked = true
		hit := HitTest(s.host.Nodes, px, py)
		s.host.SetClicked(hit)

		if hit >= 0 {
			res.Hit = s.host.Nodes[hit].ID
		}

		res.Events = append(res.Events, s.eventNames(s.host.Dispatch("click", px, py))...)
	}

	for _, key := range in.Keys {
		res.Events = append(res.Events, s.eventNames(s.host.Dispatch("keydown("+key+")", px, py))...)

		// OwnFrame reports no scroll offset for an inner box, so J and K move
		// whichever scroll container the pointer is over.
		if id, ok := s.host.ScrollerAt(px, py); ok {
			switch key {
			case "J":
				s.host.ScrollBy(id, innerScrollStep)
			case "K":
				s.host.ScrollBy(id, -innerScrollStep)
			}
		}
	}

	res.Frame = s.host.Compute(s.viewport(in.ScrollY), px, py)

	if at, ok := s.host.NextDeadline(); ok {
		res.Deadline = at
	}

	if err := s.applier.Apply(ctx, s.page, s.host.Styles(res.Frame, s.applier.base)); err != nil {
		if s.reportErr == nil {
			return res, err
		}

		if !s.reported {
			s.reported = true
			s.reportErr(err)
		}
	}

	return res, nil
}

func (s *session) eventNames(idx []int) []string {
	out := make([]string, len(idx))
	for i, k := range idx {
		e := s.host.Events[k]
		out[i] = fmt.Sprintf("%s:%s", s.host.Nodes[e.Owner].ID, e.Name)
	}

	return out
}

// inputSource is where a frame reads its input. The window supplies ebiten;
// the demo supplies a script. The scroll offset always comes from the page,
// which is where the window reports it.
type inputSource struct {
	cursor func() (x, y float64)
	click  func() bool
	keys   func() []string
}

// attach registers the per-frame callback after the window input stages.
func (s *session) attach(src inputSource) {
	s.last = FrameResult{}

	s.page.SetTick(func(ctx context.Context) error {
		if s.beforeFrame != nil {
			s.beforeFrame()
		}

		_, y := s.page.Scroll().ScrollOffset()
		cx, cy := src.cursor()

		res, err := s.frame(ctx, FrameInput{ScrollY: float64(y), CursorX: cx, CursorY: cy, Click: src.click(), Keys: src.keys()})
		s.last = res

		if !res.Deadline.IsZero() {
			// An interest timer is pending. Ask for a frame when it fires so an
			// idle window wakes for it.
			s.page.WantFrameAt(res.Deadline)
		}

		return err
	})
}
