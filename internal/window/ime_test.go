//go:build !js

package window

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"image"
	"testing"
)

func TestIMEBoundsMatchZoomedScroll(t *testing.T) {
	app := &fakeScreen{width: 200, height: 200}
	s := &shell{app: app, screenW: 200, screenH: 200, scrollX: 20, scrollY: 40, pageZoom: 2}
	box := layout.Box{X: 10, Y: 20, W: 80, H: 20}
	if got := s.imeScreenRect(box); got != image.Rect(40, 80, 200, 120) {
		t.Fatal(got)
	}
}

func TestIMEBoundsIncludeCustomizedHeader(t *testing.T) {
	s := &shell{app: &fakeScreen{width: 200, height: 200}, screenW: 200, screenH: 200, titlebar: titlebarState{enabled: true, style: headerStyle{height: 48}}}
	if got := s.imeScreenRect(layout.Box{X: 10, Y: 20, W: 80, H: 20}); got != image.Rect(10, 68, 90, 88) {
		t.Fatal(got)
	}
}

func TestIMEBoundsMatchPendingResize(t *testing.T) {
	app := &fakeScreen{width: 200, height: 200}
	s := &shell{app: app, screenW: 400, screenH: 100}
	box := layout.Box{X: 10, Y: 20, W: 80, H: 20}
	if got := s.imeScreenRect(box); got != image.Rect(20, 10, 180, 20) {
		t.Fatal(got)
	}
}

type imeAnchorScreen struct {
	*fakeScreen
	anchor layout.Box
}

func (s *imeAnchorScreen) IMEAnchor() (layout.Box, bool) { return s.anchor, true }

func TestIMEScreenRectUsesGlyphAnchor(t *testing.T) {
	app := &imeAnchorScreen{fakeScreen: &fakeScreen{width: 200, height: 200},
		anchor: layout.Box{X: 35, Y: 22, W: 1, H: 12}}
	s := &shell{app: app, screenW: 200, screenH: 200, scrollX: 10, scrollY: 3, pageZoom: 2}
	if got := s.imeCaretScreenRect(layout.Box{X: 10, Y: 10, W: 80, H: 20}); got != image.Rect(80, 47, 82, 71) {
		t.Fatalf("candidate bounds = %v", got)
	}
}

func TestIMERejectsCallbacksAfterFocusChange(t *testing.T) {
	app := &focusScreen{fakeScreen: &fakeScreen{}, focus: "first"}
	s := &shell{app: app}
	if !s.imeTargetCurrent("first") {
		t.Fatal("current session rejected")
	}
	app.focus = "second"
	if s.imeTargetCurrent("first") {
		t.Fatal("old session would edit second field")
	}
	app.focus = ""
	if s.imeTargetCurrent("first") || s.imeTargetCurrent("") {
		t.Fatal("blurred session accepted")
	}
}

func TestIMECancelInvalidatesSessionBeforeNextFocus(t *testing.T) {
	app := &focusScreen{fakeScreen: &fakeScreen{}, focus: "first"}
	s := &shell{app: app}
	s.ime.field = "first"
	if err := s.imeCancelSession(); err != nil {
		t.Fatal(err)
	}
	if s.ime.field != "" || s.imeTargetCurrent(s.ime.field) {
		t.Fatalf("cancelled field still active: %q", s.ime.field)
	}
	app.focus = "second"
	if s.imeTargetCurrent("first") {
		t.Fatal("callback from blurred field would target the new field")
	}
}

func TestEscapeCancelsIMEBeforeBlur(t *testing.T) {
	app := &focusScreen{fakeScreen: &fakeScreen{}, focus: "first"}
	s := &shell{app: app}
	s.ime.field = "first"
	if err := s.escape(); err != nil {
		t.Fatal(err)
	}
	if app.focus != "" || s.ime.field != "" {
		t.Fatalf("focus=%q session=%q", app.focus, s.ime.field)
	}
}

func TestIMEShutdownInvalidatesSession(t *testing.T) {
	s := &shell{}
	s.ime.field, s.ime.last = "first", "draft"
	s.imeShutdown()
	if s.ime.field != "" || s.ime.last != "" {
		t.Fatalf("shutdown left session state: field=%q draft=%q", s.ime.field, s.ime.last)
	}
}
