package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestDiffHeaderOrderAtMinimumSize(t *testing.T) {
	a := newTestApp(t)
	ctx := context.Background()
	a.page.SetSize(minWidth, minHeight)
	if err := a.onClick(ctx, ownframe.Box{Action: "mode-changes"}); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	ids := []string{"diff-path", "diff-status", "diff-stats", "seg-unified", "seg-split", "hunk-prev", "hunk-next", "diff-wrap-toggle"}
	var right float64
	for _, id := range ids {
		box, ok := boxByID(a.page.Boxes(), id)
		if !ok {
			t.Fatalf("missing %s", id)
		}
		if box.X < right-1 {
			t.Fatalf("%s overlaps the preceding control", id)
		}
		if box.X+box.W > minWidth {
			t.Fatalf("%s extends beyond window", id)
		}
		right = box.X + box.W
	}
}

func TestWelcomeSettingsOpensDialog(t *testing.T) {
	a := newTestApp(t)
	for len(a.tabs) > 0 {
		a.closeTab(0)
	}
	if err := a.page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	clickID(t, a, "welcome-settings")
	if !a.settingsOpen || !a.View().SettingsOpen {
		t.Fatal("welcome settings did not open the dialog")
	}
}
