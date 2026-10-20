package app

import (
	"testing"
)

// TestCalendarCreatedSurvivesWeekSwitch checks a created meeting comes back
// when its week is shown again.
func TestCalendarCreatedSurvivesWeekSwitch(t *testing.T) {
	t.Parallel()
	app := calendarApp(t)

	calClick(t, app, "calendar-new")
	app.Page().SetFormValue("cal-title", "Portfolio sync")
	calClick(t, app, "calendar-create")

	calClick(t, app, "calendar-next")
	calClick(t, app, "calendar-prev")

	found := false

	for _, e := range app.View().Calendar.Events {
		if e.Title == "Portfolio sync" {
			found = true
		}
	}

	if !found {
		t.Fatal("Portfolio sync did not survive the week switch")
	}
}
