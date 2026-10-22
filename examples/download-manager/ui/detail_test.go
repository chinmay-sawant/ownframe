package ui

import (
	"context"
	"testing"
)

func TestDetailUpdatesOnProgress(t *testing.T) {
	app, back := newTestApp(t)
	ctx := context.Background()

	back.send(Update{Kind: UpdateActive, Active: []Row{running("j1", 10, 100)}})
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	clickAction(t, app, "select-j1")

	row := running("j1", 55, 100)
	back.send(Update{Kind: UpdateProgress, Row: &row})
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if app.view.Detail == nil || app.view.Detail.Done != 55 {
		t.Fatalf("detail = %+v", app.view.Detail)
	}
}

func TestDetailSurvivesPageChange(t *testing.T) {
	app, back := newTestApp(t)
	ctx := context.Background()

	back.send(Update{Kind: UpdateActive, Active: []Row{running("j1", 10, 100)}})
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	clickAction(t, app, "select-j1")

	req := back.pages[len(back.pages)-1]
	back.send(Update{Kind: UpdateHistory, Page: &PageResponse{
		Gen:   req.Gen,
		Rows:  []Row{{ID: "other", Name: "other"}},
		Total: 1,
	}})
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if app.view.Detail == nil || app.view.Detail.ID != "j1" {
		t.Fatal("detail panel closed on a page change")
	}

	if app.pager.Selected() != "j1" {
		t.Fatalf("selection = %q", app.pager.Selected())
	}
}

func TestDetailClearsOnRemove(t *testing.T) {
	app, back := newTestApp(t)
	ctx := context.Background()

	failed := Row{ID: "j1", Name: "j1", State: StateFailed, Done: 10, Total: 100}
	back.send(Update{Kind: UpdateActive, Active: []Row{failed}})
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	clickAction(t, app, "select-j1")
	clickAction(t, app, "remove-j1")

	if app.view.Detail != nil {
		t.Fatal("detail panel stayed open after Remove")
	}

	if len(back.controls) != 1 || back.controls[0].Action != ControlRemove {
		t.Fatalf("controls = %+v", back.controls)
	}
}
