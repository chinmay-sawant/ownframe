package app

import (
	"context"
	"testing"
)

func TestNewDrawsReplayableDashboard(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	openInsights(t, app)

	if app.Page().Probe().Display() == nil {
		t.Fatal("dashboard fell back to the bitmap path")
	}
}

func TestTabSwitchKeepsDisplayList(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	openInsights(t, app)

	tab, ok := findBox(app.Boxes(), "tab-leaderboard")
	if !ok {
		t.Fatal("no tab-leaderboard box")
	}

	if err := app.Click(context.Background(), tab.X+tab.W/2, tab.Y+tab.H/2); err != nil {
		t.Fatalf("Click: %v", err)
	}

	if app.Page().Probe().Display() == nil {
		t.Fatal("the empty tab fell back to the bitmap path")
	}
}

// TestDefaultView checks the sample numbers and the landing page.
func TestDefaultView(t *testing.T) {
	v := DefaultView()

	if v.ActivePage != "dictation" {
		t.Fatalf("ActivePage = %q", v.ActivePage)
	}

	if v.ActiveTab != "usage" {
		t.Fatalf("ActiveTab = %q", v.ActiveTab)
	}

	if v.WPM.Value != "148" || v.WPM.Top != "0.2%" {
		t.Fatalf("WPM = %+v", v.WPM)
	}

	if len(v.Apps.Rows) != 6 || !v.Apps.Rows[0].Bar {
		t.Fatalf("Apps.Rows = %+v", v.Apps.Rows)
	}

	if len(v.Streak.Weeks) != 7 || len(v.Streak.Weeks[0]) != 19 {
		t.Fatalf("streak shape = %dx%d", len(v.Streak.Weeks), len(v.Streak.Weeks[0]))
	}

	if !v.Streak.Prev || v.Streak.Next {
		t.Fatalf("streak chevrons prev=%v next=%v", v.Streak.Prev, v.Streak.Next)
	}

	if len(v.Streak.Months) != 4 {
		t.Fatalf("months = %d", len(v.Streak.Months))
	}
}
