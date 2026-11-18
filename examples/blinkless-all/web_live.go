package main

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// The window reads the pointer and keys from Ebiten. A served page has no
// Ebiten, so webInput collects the same input from the page handlers that the
// -web routes call (/hover, /click, /key, /type) and hands it to the live
// sessions one frame at a time.
//
// What the routes cannot carry: /click reports the box under the point, not
// the point, so the cursor is the center of that box. A slider click lands on
// the middle of its track. There is no pointer-down or pointer-up route, so
// dragging a slider or the mock window is a window-only test.

// boxCenterDivisor halves a box's width and height to find its center.
const boxCenterDivisor = 2

// webInput is the input waiting for the next tick.
type webInput struct {
	mu    sync.Mutex
	x, y  float64
	click bool
	keys  []string
	text  strings.Builder
}

// at points the cursor at the center of b. The live runner adds the scroll
// offset to the cursor, and a box is in page coordinates, so the offset comes
// off here.
func (w *webInput) at(b ownframe.Box, scrollY int) {
	w.x, w.y = b.X+b.W/boxCenterDivisor, b.Y+b.H/boxCenterDivisor-float64(scrollY)
}

// source returns the liveSource that drains w.
func (w *webInput) source() liveSource {
	return liveSource{
		cursor: func() (float64, float64) {
			w.mu.Lock()
			defer w.mu.Unlock()

			return w.x, w.y
		},
		click: func() bool {
			w.mu.Lock()
			defer w.mu.Unlock()

			c := w.click
			w.click = false

			return c
		},
		// A route click is a press and a release in one request, so the button
		// is never reported as held between frames.
		down: func() bool { return false },
		keys: func() []string {
			w.mu.Lock()
			defer w.mu.Unlock()

			k := w.keys
			w.keys = nil

			return k
		},
		text: func() string {
			w.mu.Lock()
			defer w.mu.Unlock()

			t := w.text.String()
			w.text.Reset()

			return t
		},
	}
}

// webKeyName maps the lowercase names a route key arrives with to the names
// the live sessions read. ok is false for a key the live sessions do not read.
func webKeyName(key string) (name string, ok bool) {
	switch key {
	case "up", "arrowup":
		return "ArrowUp", true
	case "down", "arrowdown":
		return "ArrowDown", true
	case "left", "arrowleft":
		return "ArrowLeft", true
	case "right", "arrowright":
		return "ArrowRight", true
	case "backspace":
		return "Backspace", true
	case "enter":
		return "Enter", true
	}

	return "", false
}

// attachWebLive installs the live sessions on a served page and routes the
// page handlers into them.
func attachWebLive(page *ownframe.Page) {
	in := &webInput{}
	r := &liveRunner{page: page, sessions: liveSessions(), clock: wallClock{}, src: in.source()}
	r.attach()

	page.Handle(webHandlers(page, in))
}

// webHandlers routes the page handlers the web routes call into in.
func webHandlers(page *ownframe.Page, in *webInput) ownframe.Handlers {
	scrollY := func() int {
		_, sy := page.Scroll().ScrollOffset()

		return sy
	}

	return ownframe.Handlers{
		Wheel: sideWheel(page),
		Hover: func(_ context.Context, _, next ownframe.Box) (bool, error) {
			in.mu.Lock()
			in.at(next, scrollY())
			in.mu.Unlock()

			return false, nil
		},
		Click: func(_ context.Context, b ownframe.Box) error {
			in.mu.Lock()
			in.at(b, scrollY())
			in.click = true
			in.mu.Unlock()
			page.WantFrameAt(time.Now())

			return nil
		},
		KeyDown: func(_ context.Context, key string) error {
			if name, ok := webKeyName(strings.ToLower(key)); ok {
				in.mu.Lock()
				in.keys = append(in.keys, name)
				in.mu.Unlock()
				page.WantFrameAt(time.Now())
			}

			return nil
		},
		Type: func(_ context.Context, text string) error {
			in.mu.Lock()
			in.text.WriteString(text)
			in.mu.Unlock()
			page.WantFrameAt(time.Now())

			return nil
		},
		Backspace: func(_ context.Context) error {
			in.mu.Lock()
			in.keys = append(in.keys, "Backspace")
			in.mu.Unlock()
			page.WantFrameAt(time.Now())

			return nil
		},
	}
}
