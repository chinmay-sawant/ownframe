//go:build !js

package window

import "testing"

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
