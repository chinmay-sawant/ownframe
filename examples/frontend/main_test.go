package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestLaunchBoard(t *testing.T) {
	page, err := newPage()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	click := func(id string) {
		t.Helper()
		if err := pagetest.Click(ctx, page, id); err != nil {
			t.Fatalf("click %s: %v", id, err)
		}
	}
	visible := func(id string, want bool) {
		t.Helper()
		_, got := pagetest.BoxByID(page, id)
		if got != want {
			t.Fatalf("%s visible = %v, want %v", id, got, want)
		}
	}

	click("open")
	visible("design", false)
	visible("build", true)
	click("build")
	visible("build", false)
	click("done")
	visible("design", true)
	visible("build", true)
	visible("test", false)
	click("build")
	visible("build", false)
	click("all")
	click("build")
	click("test")
	click("ship")
	if !strings.Contains(page.HTML(), "Ready") {
		t.Fatal("completing every task did not mark the release ready")
	}
	click("open")
	if !strings.Contains(page.HTML(), "No tasks here") {
		t.Fatal("empty filter did not show its empty state")
	}
	click("theme")
	if !strings.Contains(page.HTML(), "Dark theme") {
		t.Fatal("theme switch did not update its label")
	}
	click("theme")
	if !strings.Contains(page.HTML(), "Light theme") {
		t.Fatal("theme switch did not restore the dark theme label")
	}
}
