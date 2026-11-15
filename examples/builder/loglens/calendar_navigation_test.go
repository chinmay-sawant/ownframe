package main

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestOnCallCalendarPreviousMonthAfterNext(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	a.calendarMonth = time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
	a.page.SetSize(windowWidth, windowHeight)
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	calendar := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
		return box.ID == "on-call-calendar"
	})
	for _, step := range []struct {
		glyph     string
		wantMonth time.Month
	}{
		{glyph: "›", wantMonth: time.December},
		{glyph: "‹", wantMonth: time.November},
	} {
		button := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
			return box.Tag == "button" && box.Text == step.glyph &&
				box.X >= calendar.X && box.X+box.W <= calendar.X+calendar.W &&
				box.Y >= calendar.Y && box.Y < calendar.Y+48
		})
		cx, cy := button.X+button.W/2, button.Y+button.H/2
		// The -web server selects at the point before it clicks.
		if err := a.page.Testing().SelectAt(ctx, cx, cy); err != nil {
			t.Fatalf("selecting %q month navigation: %v", step.glyph, err)
		}
		if err := a.page.Click(ctx, cx, cy); err != nil {
			t.Fatalf("clicking %q month navigation: %v", step.glyph, err)
		}
		if a.calendarMonth.Month() != step.wantMonth {
			t.Fatalf("clicking %q showed %s, want %s", step.glyph, a.calendarMonth.Format("January"), step.wantMonth)
		}
	}
}
