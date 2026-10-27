package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestFileNavChangeArrows(t *testing.T) {
	app := newTestApp(t)
	path := "backend/probe.go"
	app.changes = changeSet{Files: []changedFile{{Path: path, Status: "M", Add: 2, Hunks: []hunk{
		{NewStart: 3, Lines: []diffLine{{Kind: '+', NewNo: 3, Text: "first"}}},
		{NewStart: 12, Lines: []diffLine{{Kind: '+', NewNo: 12, Text: "second"}}},
	}}}}
	app.syncFileNav()
	app.openFile(path)
	app.wrap = false
	ctx := context.Background()
	for _, step := range []struct {
		action string
		row    int
	}{
		{"file-nav-change-next", 2},
		{"file-nav-change-next", 11},
		{"file-nav-change-prev", 2},
	} {
		if err := app.onClick(ctx, ownframe.Box{Action: step.action}); err != nil {
			t.Fatal(err)
		}
		if !app.changeNav.valid || app.changeNav.row != step.row {
			t.Fatalf("%s: navigation = %+v, want row %d", step.action, app.changeNav, step.row)
		}
	}
	app.toggleSplit()
	app.openFileIn(1, path)
	if err := app.onClick(ctx, ownframe.Box{Action: "file-nav-change-next"}); err != nil {
		t.Fatal(err)
	}
	if app.changeNav.group != 1 || app.changeNav.row != 2 || app.scrollGroup[1] != app.changeNav.scroll {
		t.Fatalf("split navigation = %+v, scroll = %v", app.changeNav, app.scrollGroup)
	}
	app.selectDiff(path)
	rows := app.fileNavChangeRows(path)
	if len(rows) != 2 {
		t.Fatalf("diff change rows = %v", rows)
	}
	for _, row := range rows {
		if err := app.onClick(ctx, ownframe.Box{Action: "file-nav-change-next"}); err != nil {
			t.Fatal(err)
		}
		if app.changeNav.row != row || app.changeNav.mode != "changes" {
			t.Fatalf("diff navigation = %+v, want row %d", app.changeNav, row)
		}
	}
}

func TestFileNavChangeArrowLayout(t *testing.T) {
	app := newTestApp(t)
	for _, width := range []int{1200, 800, 350} {
		app.page.SetSize(width, 700)
		if err := app.page.Redraw(context.Background()); err != nil {
			t.Fatal(err)
		}
		boxes := app.page.Boxes()
		bar, ok := findBoxID(boxes, "file-nav")
		if !ok {
			t.Fatalf("width %d: missing navigator", width)
		}
		var end float64
		for _, id := range []string{"file-nav-change-prev", "file-nav-change-next", "file-nav-next"} {
			box, found := findBoxID(boxes, id)
			if !found || box.X < end || box.X+box.W > bar.X+bar.W {
				t.Fatalf("width %d: %s = %+v, bar = %+v, previous end = %v", width, id, box, bar, end)
			}
			end = box.X + box.W
		}
	}
}
