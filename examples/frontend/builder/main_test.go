package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestToggleFlipsState(t *testing.T) {
	ctx := context.Background()
	page, state, err := newPage()
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if state.Pressed {
		t.Fatalf("toggle starts pressed: %+v", state)
	}
	if _, ok := pagetest.BoxByID(page, "toggle"); !ok {
		t.Fatal("toggle was not rendered")
	}
	if err := pagetest.Click(ctx, page, "toggle"); err != nil {
		t.Fatal(err)
	}
	if !state.Pressed {
		t.Fatalf("toggle did not flip on click: %+v", state)
	}
	if err := pagetest.Click(ctx, page, "toggle"); err != nil {
		t.Fatal(err)
	}
	if state.Pressed {
		t.Fatalf("toggle did not flip back: %+v", state)
	}
}

func TestSwitchAndButtonState(t *testing.T) {
	ctx := context.Background()
	page, state, err := newPage()
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, page, "notify"); err != nil {
		t.Fatal(err)
	}
	if !state.Notify {
		t.Fatalf("switch did not flip on click: %+v", state)
	}
	if err := pagetest.Click(ctx, page, "press"); err != nil {
		t.Fatal(err)
	}
	if state.Clicks != 1 {
		t.Fatalf("button click did not count: %+v", state)
	}
	if _, ok := pagetest.BoxByID(page, "press"); !ok {
		t.Fatal("button was not rendered")
	}
}
