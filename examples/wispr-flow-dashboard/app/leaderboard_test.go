package app

import (
	"strings"
	"testing"
)

// TestInsightsLeaderboardTab clicks the leaderboard tab and checks the rows.
func TestInsightsLeaderboardTab(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	openInsights(t, app)

	tab, ok := findBox(app.Boxes(), "tab-leaderboard")
	if !ok {
		t.Fatal("no tab-leaderboard box")
	}

	clickBox(t, app, tab)

	if got := app.View().ActiveTab; got != "leaderboard" {
		t.Fatalf("ActiveTab = %q", got)
	}

	if _, ok := findBox(app.Boxes(), "insights-leaderboard"); !ok {
		t.Fatal("no insights-leaderboard box after the leaderboard tab click")
	}

	you, ok := findBox(app.Boxes(), "lead-you")
	if !ok {
		t.Fatal("no lead-you box")
	}

	if !strings.Contains(you.Text, "You") {
		t.Fatalf("lead-you text = %q", you.Text)
	}

	if app.Page().Display() == nil {
		t.Fatal("the leaderboard tab fell back to the bitmap path")
	}

	rows := DefaultView().Leaderboard.Rows

	if len(rows) != 8 {
		t.Fatalf("rows = %d", len(rows))
	}

	if !rows[3].You || rows[3].Name != "You" || rows[3].Words != "29,804" {
		t.Fatalf("row 4 = %+v", rows[3])
	}
}
