package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

// Profile popover follows the search backdrop pattern: a viewport-unit
// dismiss layer sits below the card, so clicks outside close it while
// clicks inside keep it open. Navigation also drops the card.
func TestProfileDismissClosesOutside(t *testing.T) {
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, a.page, "dc-member-u-mira"); err != nil {
		t.Fatalf("open profile: %v", err)
	}
	if a.profileOpen == "" {
		t.Fatal("profile did not open")
	}
	dismiss, ok := pagetest.BoxByID(a.page, "dc-dismiss")
	if !ok {
		t.Fatal("missing dc-dismiss while profile open")
	}
	if dismiss.H < 600 {
		t.Fatalf("dismiss height %.0f too small, want viewport height", dismiss.H)
	}
	if _, ok := pagetest.BoxByID(a.page, "dc-profile"); !ok {
		t.Fatal("missing dc-profile")
	}
	if err := pagetest.Click(ctx, a.page, "dc-dismiss"); err != nil {
		t.Fatalf("click dismiss: %v", err)
	}
	if a.profileOpen != "" {
		t.Fatal("dismiss click did not close profile")
	}
}

func TestProfileNavigationClosesCard(t *testing.T) {
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, a.page, "dc-member-u-ada"); err != nil {
		t.Fatal(err)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "chan:select:ownframe-general"}); err != nil {
		t.Fatal(err)
	}
	if a.profileOpen != "" {
		t.Fatal("channel select left profile open")
	}
}
