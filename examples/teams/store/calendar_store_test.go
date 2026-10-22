package store

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/teams/calendar"
)

// TestCalendarRoundTrip seeds the calendar from the SQL files, creates a
// meeting, and checks it survives Save, Load, and a next/prev week switch.
func TestCalendarRoundTrip(t *testing.T) {
	st, err := Open(Memory)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	defer func() { _ = st.Close() }()

	if err := st.Seed(); err != nil {
		t.Fatalf("Seed: %v", err)
	}

	d, ok, err := st.Load()
	if err != nil || !ok {
		t.Fatalf("Load: %v ok=%v", err, ok)
	}

	before := len(d.Calendar.AllEvents())
	if before == 0 {
		t.Fatal("seed loaded no calendar events")
	}

	page, err := ownframe.New(ownframe.Config{HTML: "<p>cal</p>", Width: 800, Height: 600})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx := context.Background()

	if !calendar.Handle(ctx, page, &d.Calendar, "calendar-create") {
		t.Fatal("calendar-create not handled")
	}

	if err := st.Save(d); err != nil {
		t.Fatalf("Save: %v", err)
	}

	d, ok, err = st.Load()
	if err != nil || !ok {
		t.Fatalf("Load: %v ok=%v", err, ok)
	}

	cal := d.Calendar

	if got := len(cal.AllEvents()); got != before+1 {
		t.Fatalf("events after reload = %d, want %d", got, before+1)
	}

	found := false

	for _, e := range cal.Events {
		if e.Title == "Untitled meeting" {
			found = e.Week == cal.Week
		}
	}

	if !found {
		t.Fatal("created meeting missing from its week")
	}

	if !calendar.Handle(ctx, page, &cal, "calendar-next") ||
		!calendar.Handle(ctx, page, &cal, "calendar-prev") {
		t.Fatal("week switch not handled")
	}

	found = false

	for _, e := range cal.Events {
		if e.Title == "Untitled meeting" {
			found = true
		}
	}

	if !found {
		t.Fatal("created meeting lost after a week switch")
	}
}
