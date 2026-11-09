package main

import (
	"context"
	"strings"
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
	if !strings.Contains(page.HTML(), "Run action") {
		t.Fatal("button label was not Run action")
	}
	if !strings.Contains(page.HTML(), statusText(state)) {
		t.Fatalf("status = %q missing from %s", statusText(state), page.HTML())
	}
	if err := pagetest.Click(ctx, page, "toggle"); err != nil {
		t.Fatal(err)
	}
	if !state.Pressed {
		t.Fatalf("toggle did not flip on click: %+v", state)
	}
	if !strings.Contains(page.HTML(), statusText(state)) || !strings.Contains(page.HTML(), "Run action") {
		t.Fatalf("toggle redraw mismatch: %s", page.HTML())
	}
	if err := pagetest.Click(ctx, page, "toggle"); err != nil {
		t.Fatal(err)
	}
	if state.Pressed {
		t.Fatalf("toggle did not flip back: %+v", state)
	}
	if !strings.Contains(page.HTML(), statusText(state)) {
		t.Fatalf("status = %q missing from %s", statusText(state), page.HTML())
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
	if !strings.Contains(page.HTML(), statusText(state)) {
		t.Fatalf("status = %q missing from %s", statusText(state), page.HTML())
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
	if !strings.Contains(page.HTML(), "Ran 1 time") || strings.Contains(page.HTML(), "Ran 1 times") {
		t.Fatalf("button label mismatch: %s", page.HTML())
	}
	if !strings.Contains(page.HTML(), "Button ran 1 times. Toggle false. Switch true.") {
		t.Fatalf("status mismatch: %s", page.HTML())
	}
}
