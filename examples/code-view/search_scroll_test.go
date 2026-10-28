package main

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestWorkspaceSearchScroll(t *testing.T) {
	ctx := context.Background()
	app := newTestApp(t)
	clickID(t, app, "act-search")
	if !app.searchOpen {
		t.Fatal("search icon did not open the search panel")
	}
	for i := 0; i < 24; i++ {
		app.searchHits = append(app.searchHits, workspaceHit{Path: "README.md", Line: i + 1, Col: 1, Match: fmt.Sprint(i)})
	}
	app.refresh()
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if !app.View().Search.Scroll || len(app.View().Search.Rows) != searchWindowRows {
		t.Fatalf("search view = %+v", app.View().Search)
	}
	if os.Getenv("OWNFRAME_CODE_VIEW_SCREENSHOT") == "1" {
		writeShot(t, app, "search-scroll")
	}
	box, ok := boxByID(app.Page().Boxes(), "search-results")
	if !ok {
		t.Fatal("missing search results box")
	}
	claimed, err := app.onWheel(ctx, box.X+10, box.Y+10, 0, -1)
	if err != nil || !claimed || app.View().Search.Rows[0].Index == 0 {
		t.Fatalf("wheel claimed=%v err=%v rows=%+v", claimed, err, app.View().Search.Rows)
	}
	thumb, ok := boxByID(app.Page().Boxes(), "search-thumb")
	if !ok {
		t.Fatal("missing search scrollbar thumb")
	}
	claimed, err = app.onDragStart(ctx, thumb)
	if err != nil || !claimed {
		t.Fatalf("thumb claimed=%v err=%v", claimed, err)
	}
	if err := app.onDragMove(ctx, 0, 1000); err != nil {
		t.Fatal(err)
	}
	if err := app.onDragEnd(ctx); err != nil {
		t.Fatal(err)
	}
	rows := app.View().Search.Rows
	if rows[len(rows)-1].Index != 23 || app.searchIndex < rows[0].Index {
		t.Fatalf("last window = %+v index=%d", rows, app.searchIndex)
	}
	app.searchIndex = 23
	app.searchStep(1)
	if app.searchIndex != 0 || app.View().Search.Rows[0].Index != 0 {
		t.Fatalf("keyboard wrap index=%d rows=%+v", app.searchIndex, app.View().Search.Rows)
	}
	app.openSearchHit(0)
	if app.searchOpen || app.View().Active != "README.md" {
		t.Fatalf("result did not open: search=%v active=%s", app.searchOpen, app.View().Active)
	}
}
