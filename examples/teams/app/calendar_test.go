package app

import (
	"testing"
)

func calendarApp(t *testing.T) *App {
	t.Helper()

	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	openSection(t, app, "calendar")

	return app
}

func calClick(t *testing.T, app *App, id string) {
	t.Helper()

	box, ok := findBox(app.Boxes(), id)
	if !ok {
		t.Fatalf("no %s box", id)
	}

	clickBox(t, app, box)
}

// TestCalendarWeeks checks next, prev, and Today move between weeks.
func TestCalendarWeeks(t *testing.T) {
	t.Parallel()
	app := calendarApp(t)
	first := app.View().Calendar

	calClick(t, app, "calendar-next")

	cal := app.View().Calendar
	if cal.Week != first.Week+1 || cal.WeekLabel == first.WeekLabel {
		t.Fatalf("after next: %d %q", cal.Week, cal.WeekLabel)
	}

	calClick(t, app, "calendar-today")

	cal = app.View().Calendar
	if cal.Week != first.Week || cal.WeekLabel != first.WeekLabel {
		t.Fatalf("after today: %d %q", cal.Week, cal.WeekLabel)
	}

	calClick(t, app, "calendar-prev")

	if got := app.View().Calendar.Week; got != first.Week-1 {
		t.Fatalf("after prev: %d", got)
	}
}

// TestCalendarDetail checks opening an event shows the detail card.
func TestCalendarDetail(t *testing.T) {
	t.Parallel()
	app := calendarApp(t)

	calClick(t, app, "calendar-event-e1")

	if got := app.View().Calendar.Selected; got != "e1" {
		t.Fatalf("Selected = %q", got)
	}

	if _, ok := findBox(app.Boxes(), "calendar-detail"); !ok {
		t.Fatal("no calendar-detail box")
	}
}
