package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestSplitDistributesEveryCent(t *testing.T) {
	v := view{Amount: "100.01", Friends: []friend{{Name: "A", Selected: true}, {Name: "B", Selected: true}, {Name: "C", Selected: true}, {Name: "D"}}}
	v.split()
	if !v.HasSplit || v.Error != "" || v.Selected != 3 {
		t.Fatalf("unexpected split: %+v", v)
	}
	for i, want := range []string{"$33.34", "$33.34", "$33.33", ""} {
		if v.Friends[i].Share != want {
			t.Errorf("friend %d: got %q, want %q", i, v.Friends[i].Share, want)
		}
	}
	if len(v.Segments) != 3 {
		t.Fatalf("pie has %d segments", len(v.Segments))
	}
}

func TestInvalidSplit(t *testing.T) {
	for _, input := range []string{"", "-1", "0", "1.234", "1x", "1000001"} {
		if _, err := parseAmount(input); err == nil {
			t.Errorf("accepted %q", input)
		}
	}
	v := view{Amount: "10", Friends: []friend{{Selected: true}}}
	v.split()
	if v.HasSplit || v.Error == "" {
		t.Fatalf("accepted one friend: %+v", v)
	}
}

func TestAddFriendAndSplit(t *testing.T) {
	ctx := context.Background()
	page, state, err := newPage()
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := page.Form().Focus(ctx, "friend-name"); err != nil {
		t.Fatal(err)
	}
	if err := page.Edit().Type(ctx, "Maya"); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, page, "add"); err != nil {
		t.Fatal(err)
	}
	if len(state.Friends) != 5 || state.Friends[4].Name != "Maya" {
		t.Fatalf("added friend is missing: %+v", state.Friends)
	}
	if _, ok := pagetest.BoxByID(page, "friend-4"); !ok {
		t.Fatal("added friend was not rendered")
	}
	if err := pagetest.Click(ctx, page, "split"); err != nil {
		t.Fatal(err)
	}
	for _, index := range []int{0, 1, 2, 4} {
		if got := state.Friends[index].Share; got != "$30.00" {
			t.Fatalf("friend %d: got %s, want $30.00", index, got)
		}
	}
}

func TestSettlementAndSavedActivity(t *testing.T) {
	ctx := context.Background()
	page, state, err := newPage()
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"screen-settle", "payer-1", "screen-split", "save", "screen-activity"} {
		if err := pagetest.Click(ctx, page, action); err != nil {
			t.Fatalf("click %s: %v", action, err)
		}
	}
	if state.Screen != "activity" || len(state.History) != 1 {
		t.Fatalf("saved activity missing: %+v", state)
	}
	entry := state.History[0]
	if entry.Payer != "Sam" || entry.Total != "$120.00" || len(entry.Shares) != 3 || entry.Shares[0].Share != "$40.00" {
		t.Fatalf("incorrect saved split: %+v", entry)
	}
	for _, action := range []string{"screen-settle", "payer-0", "screen-split", "save"} {
		if err := pagetest.Click(ctx, page, action); err != nil {
			t.Fatalf("click %s: %v", action, err)
		}
	}
	if len(state.History) != 2 || state.History[0].Payer != "Alex" || state.History[1].Payer != "Sam" {
		t.Fatalf("payer change did not produce a new snapshot: %+v", state.History)
	}
	state.Friends[0].Share = "$0.00"
	if entry.Shares[0].Share != "$40.00" {
		t.Fatal("saved split changed when the current split changed")
	}
}
