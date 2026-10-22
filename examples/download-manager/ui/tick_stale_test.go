package ui

import (
	"context"
	"testing"
)

func TestTickDiscardsStaleSummary(t *testing.T) {
	app, back := newTestApp(t)
	ctx := context.Background()

	back.send(Update{Kind: UpdateActive, Active: []Row{running("j1", 1, 100)}})
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if len(back.summary) == 0 {
		t.Fatal("no summary request")
	}

	gen := back.summary[len(back.summary)-1]
	fresh := Summary{Running: 1}

	back.send(Update{Kind: UpdateSummary, Gen: gen - 1, Summary: &fresh})
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if app.view.Summary.Running != 0 {
		t.Fatal("a stale summary was applied")
	}

	back.send(Update{Kind: UpdateSummary, Gen: gen, Summary: &fresh})
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if app.view.Summary.Running != 1 {
		t.Fatal("the current summary was discarded")
	}
}

func TestTickDiscardsStalePage(t *testing.T) {
	app, back := newTestApp(t)
	ctx := context.Background()

	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if len(back.pages) != 1 {
		t.Fatalf("page requests = %d, want 1", len(back.pages))
	}

	gen := back.pages[0].Gen
	back.send(Update{Kind: UpdateHistory, Page: &PageResponse{Gen: gen - 1, Rows: []Row{{ID: "old"}}}})
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if len(app.view.History) != 0 {
		t.Fatal("a stale page was applied")
	}

	back.send(Update{Kind: UpdateHistory, Page: &PageResponse{Gen: gen, Rows: []Row{{ID: "new"}}, Total: 1}})
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if len(app.view.History) != 1 {
		t.Fatal("the current page was discarded")
	}
}
