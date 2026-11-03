package main

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestApplicationMenuTriggersStayStill(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err = a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	ids := []string{"menu-file-trigger", "menu-view-trigger", "menu-help-trigger"}
	initial := map[string]ownframe.Box{}
	for _, id := range ids {
		initial[id] = loglensLayoutBox(t, a.page, func(b ownframe.Box) bool { return b.ID == id })
	}
	for _, id := range ids {
		b := loglensLayoutBox(t, a.page, func(b ownframe.Box) bool { return b.ID == id })
		if err = a.page.Click(ctx, b.X+b.W/2, b.Y+b.H/2); err != nil {
			t.Fatal(err)
		}
		for _, check := range ids {
			b = loglensLayoutBox(t, a.page, func(b ownframe.Box) bool { return b.ID == check })
			if !loglensSameBox(initial[check], b) {
				t.Fatalf("opening %s moved %s: before=%+v after=%+v", id, check, initial[check], b)
			}
		}
	}
}

func TestLiveAgeTicksWhileOpen(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err = a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	b := loglensLayoutBox(t, a.page, func(b ownframe.Box) bool { return b.ID == "tooltip-trigger" })
	if err = a.page.Edit().Hover(ctx, b.X+b.W/2, b.Y+b.H/2); err != nil {
		t.Fatal(err)
	}
	now := a.lastUpdatedAt.Add(25 * time.Second)
	if err = a.tickLiveAt(ctx, now); err != nil {
		t.Fatal(err)
	}
	b = loglensLayoutBox(t, a.page, func(b ownframe.Box) bool { return b.ID == "live-tooltip-content" })
	if b.Text != "Updated 25 seconds ago" {
		t.Fatalf("stale live age %q", b.Text)
	}
	a.autoRefresh = true
	if err = a.tickLiveAt(ctx, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if !a.lastUpdatedAt.Equal(now.Add(time.Second)) {
		t.Fatal("auto refresh did not advance sample time")
	}
}
