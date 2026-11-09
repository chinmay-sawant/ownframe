package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/pagetest"
)

// Settings docks below the 48px topbar so the brand, search, and window
// controls stay visible and clickable while settings is open.
func TestSettingsKeepsTopbarVisible(t *testing.T) {
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, a.page, "dc-top-settings"); err != nil {
		t.Fatalf("open settings: %v", err)
	}
	setBox, ok := pagetest.BoxByID(a.page, "dc-settings")
	if !ok {
		t.Fatal("missing dc-settings")
	}
	topBox, ok := pagetest.BoxByID(a.page, "dc-topbar")
	if !ok {
		t.Fatal("missing dc-topbar while settings open")
	}
	if setBox.Y < topBox.Y+topBox.H-2 {
		t.Fatalf("settings Y %.0f overlaps topbar bottom %.0f", setBox.Y, topBox.Y+topBox.H)
	}
}

// Header search is a wide quick-switcher affordance, not a content-sized
// button.
func TestTopSearchHasWideAffordance(t *testing.T) {
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	searchBox, ok := pagetest.BoxByID(a.page, "dc-top-search")
	if !ok {
		t.Fatal("missing dc-top-search")
	}
	if searchBox.W < 300 {
		t.Fatalf("search width %.0f too narrow, want >=300", searchBox.W)
	}
}
