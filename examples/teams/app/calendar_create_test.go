package app

import (
	"testing"
)

// TestCalendarCreate checks the modal creates the chosen meeting.
func TestCalendarCreate(t *testing.T) {
	t.Parallel()
	app := calendarApp(t)

	calClick(t, app, "calendar-new")

	if !app.View().Calendar.Compose {
		t.Fatal("Compose = false after calendar-new")
	}

	if _, ok := findBox(app.Boxes(), "calendar-compose"); !ok {
		t.Fatal("no calendar-compose box")
	}

	before := len(app.View().Calendar.Events)

	app.Page().Form().SetFormValue("cal-title", "Budget review")
	calClick(t, app, "cal-date-1")
	calClick(t, app, "cal-start-11")
	calClick(t, app, "cal-dur-2")
	calClick(t, app, "calendar-create")

	cal := app.View().Calendar
	if cal.Compose {
		t.Fatal("Compose = true after calendar-create")
	}

	if len(cal.Events) != before+1 {
		t.Fatalf("events = %d, want %d", len(cal.Events), before+1)
	}

	made := false

	for _, e := range cal.Events {
		if e.Title == "Budget review" {
			made = e.Day == 1 && e.Time == "2:30 PM" && e.Dur == "45 min" && e.Color == "violet"
		}
	}

	if !made {
		t.Fatal("no Budget review event on Tuesday at 2:30 PM")
	}
}

// TestCalendarCancel closes the modal without creating anything.
func TestCalendarCancel(t *testing.T) {
	t.Parallel()
	app := calendarApp(t)

	calClick(t, app, "calendar-new")
	calClick(t, app, "calendar-cancel")

	if app.View().Calendar.Compose {
		t.Fatal("Compose = true after calendar-cancel")
	}

	if _, ok := findBox(app.Boxes(), "calendar-compose"); ok {
		t.Fatal("calendar-compose still drawn")
	}
}
