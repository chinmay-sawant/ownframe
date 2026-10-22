package ui

import (
	"context"
	"testing"
)

func TestClickAddJob(t *testing.T) {
	app, back := newTestApp(t)
	ctx := context.Background()

	if err := app.Redraw(ctx); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	app.page.SetFormValue("url", "https://example.com/a.bin")
	app.page.SetFormValue("dest", "/tmp/downloads")
	clickVisible(t, app, "add")

	if len(back.added) != 1 || back.added[0].URL != "https://example.com/a.bin" {
		t.Fatalf("add = %+v", back.added)
	}

	if got := app.page.FormValue("url"); got != "" {
		t.Fatalf("url field = %q, want empty", got)
	}
}

func TestClickAddNeedsURL(t *testing.T) {
	app, back := newTestApp(t)

	clickAction(t, app, "add")

	if len(back.added) != 0 {
		t.Fatal("empty URL was queued")
	}

	if app.view.Notice == "" {
		t.Fatal("no notice for the missing URL")
	}
}

func TestClickControlAndSelect(t *testing.T) {
	app, back := newTestApp(t)
	ctx := context.Background()

	back.send(Update{Kind: UpdateActive, Active: []Row{running("j1", 10, 100)}})
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	clickAction(t, app, "pause-j1")

	if len(back.controls) != 1 || back.controls[0].Action != ControlPause || back.controls[0].ID != "j1" {
		t.Fatalf("controls = %+v", back.controls)
	}

	clickAction(t, app, "select-j1")

	if app.pager.Selected() != "j1" {
		t.Fatalf("selected = %q", app.pager.Selected())
	}

	if app.view.Detail == nil || app.view.Detail.ID != "j1" {
		t.Fatal("detail panel did not open")
	}
}
