//go:build !android && !ios && !js

package window

import (
	"context"
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// imeCompositionScreen records what the window forwards to a screen.
type imeCompositionScreen struct {
	*focusScreen
	box        layout.Box
	anchor     layout.Box
	text       string
	start, end int
	calls      int
}

func (s *imeCompositionScreen) IMEContext() (layout.Box, int, string, string, bool) {
	return s.box, 0, "", "", true
}

func (s *imeCompositionScreen) IMEReplace(context.Context, int, int, string, int) error {
	return nil
}

func (s *imeCompositionScreen) IMEAnchor() (layout.Box, bool) { return s.anchor, true }

func (s *imeCompositionScreen) SetIMEComposition(_ context.Context, text string, start, end int) error {
	s.text, s.start, s.end = text, start, end
	s.calls++

	return nil
}

func newIMEShell(focus string) (*shell, *imeCompositionScreen) {
	app := &imeCompositionScreen{
		focusScreen: &focusScreen{fakeScreen: &fakeScreen{width: 200, height: 200}, focus: focus},
		box:         layout.Box{X: 10, Y: 10, W: 100, H: 20},
		anchor:      layout.Box{X: 10, Y: 10, W: 1, H: 16},
	}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200}
	s.ime.field = focus
	s.ime.caret = s.imeCaretScreenRect(app.box)

	return s, app
}

func TestIMEPreeditReachesScreenAndTracksComposition(t *testing.T) {
	s, app := newIMEShell("f")

	s.imePreedit("かな", 6, 6)
	if app.text != "かな" || app.start != 6 || app.end != 6 || !s.ime.composing {
		t.Fatalf("preedit = %q [%d,%d] composing=%v", app.text, app.start, app.end, s.ime.composing)
	}

	// Cancel and commit both clear the preedit before the value changes.
	s.imeDrop()
	if app.text != "" || s.ime.composing {
		t.Fatalf("drop left preedit %q composing=%v", app.text, s.ime.composing)
	}
}

func TestIMEPreeditIgnoredAfterFocusMoves(t *testing.T) {
	s, app := newIMEShell("f")
	app.focus = "other"

	s.imePreedit("か", 3, 3)
	if app.calls != 0 || s.ime.composing {
		t.Fatalf("stale preedit reached the new field: calls=%d composing=%v", app.calls, s.ime.composing)
	}

	// An empty composition must still clear a stale preedit.
	s.imePreedit("", 0, 0)
	if app.calls != 1 {
		t.Fatalf("clear calls = %d", app.calls)
	}
}

// Windows reports the clause cursor at the end of a fresh preedit, which
// moves the glyph anchor. The window must not restart (and so commit) the
// session while a composition is live.
func TestIMEResizeKeepsLiveComposition(t *testing.T) {
	s, app := newIMEShell("f")
	s.imePreedit("かな", 6, 6)
	s.ime.last = "live"
	app.anchor = layout.Box{X: 60, Y: 10, W: 1, H: 16}

	s.imeResize()
	if s.ime.last != "live" {
		t.Fatal("session restarted during composition")
	}
}

func TestIMEResizeRestartsIdleSessionWhenCaretMoves(t *testing.T) {
	s, app := newIMEShell("f")
	s.ime.last = "idle"
	app.anchor = layout.Box{X: 60, Y: 10, W: 1, H: 16}

	s.imeResize()
	if s.ime.last != "" {
		t.Fatal("idle session kept stale candidate bounds")
	}
}

func TestIMEResizeIgnoresCaretJitter(t *testing.T) {
	s, app := newIMEShell("f")
	s.ime.last = "idle"
	app.anchor.X += 2

	s.imeResize()
	if s.ime.last != "idle" {
		t.Fatal("jitter restarted the session")
	}
	if !imeNear(image.Rect(0, 0, 4, 4), image.Rect(4, 4, 8, 8)) {
		t.Fatal("slack boundary changed")
	}
}
