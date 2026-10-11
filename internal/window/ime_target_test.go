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
