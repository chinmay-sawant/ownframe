package app

import (
	"testing"
)

// TestInsightsVoiceTab clicks the voice tab and checks the setup panel.
func TestInsightsVoiceTab(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	openInsights(t, app)

	tab, ok := findBox(app.Boxes(), "tab-voice")
	if !ok {
		t.Fatal("no tab-voice box")
	}

	clickBox(t, app, tab)

	if got := app.View().ActiveTab; got != "voice" {
		t.Fatalf("ActiveTab = %q", got)
	}

	if _, ok := findBox(app.Boxes(), "insights-voice"); !ok {
		t.Fatal("no insights-voice box after the voice tab click")
	}

	if _, ok := findBox(app.Boxes(), "vt-setup"); !ok {
		t.Fatal("no vt-setup box")
	}

	if app.Page().Display() == nil {
		t.Fatal("the voice tab fell back to the bitmap path")
	}

	if len(app.View().VoiceTab.Checklist) != 3 {
		t.Fatalf("checklist = %d", len(app.View().VoiceTab.Checklist))
	}
}
