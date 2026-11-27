package app

import (
	"context"
	"math"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestBoxesCoverTheDashboard(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	openInsights(t, app)

	boxes := app.Boxes()

	wpm, ok := findBox(boxes, "card-wpm")
	if !ok {
		t.Fatal("no card-wpm box")
	}

	fixes, ok := findBox(boxes, "card-fixes")
	if !ok {
		t.Fatal("no card-fixes box")
	}

	words, ok := findBox(boxes, "card-words")
	if !ok {
		t.Fatal("no card-words box")
	}

	apps, ok := findBox(boxes, "card-apps")
	if !ok {
		t.Fatal("no card-apps box")
	}

	streak, ok := findBox(boxes, "card-streak")
	if !ok {
		t.Fatal("no card-streak box")
	}

	if wpm.Y != fixes.Y || wpm.Y != words.Y {
		t.Fatalf("top cards y = %v %v %v", wpm.Y, fixes.Y, words.Y)
	}

	if wpm.W != fixes.W || math.Abs(words.W-2*wpm.W) > 0.01 {
		t.Fatalf("top card widths = %v %v", wpm.W, words.W)
	}

	if apps.Y != streak.Y || apps.Y <= wpm.Y {
		t.Fatalf("bottom cards y = %v %v", apps.Y, streak.Y)
	}

	if apps.W < 480 || streak.W < 480 {
		t.Fatalf("bottom card widths = %v %v", apps.W, streak.W)
	}
}

func TestClickSwitchesTab(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	openInsights(t, app)

	tab, ok := findBox(app.Boxes(), "tab-voice")
	if !ok {
		t.Fatal("no tab-voice box")
	}

	if err := app.Click(context.Background(), tab.X+tab.W/2, tab.Y+tab.H/2); err != nil {
		t.Fatalf("Click: %v", err)
	}

	if got := app.View().ActiveTab; got != "voice" {
		t.Fatalf("ActiveTab = %q", got)
	}

	if _, ok := findBox(app.Boxes(), "insights-voice"); !ok {
		t.Fatal("no insights-voice box after the voice tab click")
	}
}

func findBox(boxes []ownframe.Box, id string) (ownframe.Box, bool) {
	for _, box := range boxes {
		if box.ID == id {
			return box, true
		}
	}

	return ownframe.Box{}, false
}
