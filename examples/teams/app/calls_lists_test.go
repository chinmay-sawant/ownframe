package app

import (
	"testing"
)

// TestCallsCallAndHangup checks Call starts a call and Hangup ends it.
func TestCallsCallAndHangup(t *testing.T) {
	t.Parallel()
	app := openCallsMenu(t)

	clickCallsBox(t, app, "calls-key-7")
	clickCallsBox(t, app, "calls-call")

	calls := app.View().Calls
	if calls.Status != "calling" {
		t.Fatalf("Status = %q after call", calls.Status)
	}

	if calls.StatusName != "7" {
		t.Fatalf("StatusName = %q after call", calls.StatusName)
	}

	clickCallsBox(t, app, "calls-hangup")

	if got := app.View().Calls.Status; got != "" {
		t.Fatalf("Status = %q after hangup", got)
	}
}

// TestCallsOpenHistory checks a history row becomes Active and names the caller.
func TestCallsOpenHistory(t *testing.T) {
	t.Parallel()
	app := openCallsMenu(t)

	clickCallsBox(t, app, "calls-item-c2")

	calls := app.View().Calls
	if calls.Active != "c2" {
		t.Fatalf("Active = %q after opening c2", calls.Active)
	}

	if calls.StatusName != "Pepper Potts" {
		t.Fatalf("StatusName = %q after opening c2", calls.StatusName)
	}

	clickCallsBox(t, app, "calls-call")

	if got := app.View().Calls.StatusName; got != "Pepper Potts" {
		t.Fatalf("StatusName = %q after calling the active person", got)
	}
}

// TestCallsVoicemailTab checks the tab switch, the voicemail rows, and the transcript.
func TestCallsVoicemailTab(t *testing.T) {
	t.Parallel()
	app := openCallsMenu(t)

	clickCallsBox(t, app, "calls-tab-voicemail")

	if got := app.View().Calls.Tab; got != "voicemail" {
		t.Fatalf("Tab = %q after the voicemail tab", got)
	}

	if _, ok := findBox(app.Boxes(), "calls-vm-v1"); !ok {
		t.Fatal("no calls-vm-v1 box on the voicemail tab")
	}

	clickCallsBox(t, app, "calls-vm-v1")

	calls := app.View().Calls
	if calls.Active != "v1" {
		t.Fatalf("Active = %q after opening v1", calls.Active)
	}

	if calls.StatusName != "Happy Hogan" {
		t.Fatalf("StatusName = %q after opening v1", calls.StatusName)
	}
}
