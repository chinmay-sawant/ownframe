package main

import (
	"context"
	"strings"
	"testing"
)

func TestContextMenuTabRows(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")

	rows := app.ContextMenuFor("tab-0")
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(rows))
	}

	labels := []string{"Close", "Close Others", "Close All"}
	for i, label := range labels {
		if rows[i].Label != label {
			t.Errorf("row %d label = %q, want %q", i, rows[i].Label, label)
		}

		if !strings.HasPrefix(rows[i].ID, "code-view:") {
			t.Errorf("row %d id = %q, want the code-view prefix", i, rows[i].ID)
		}

		if !rows[i].Enabled {
			t.Errorf("row %d %q is disabled with two tabs", i, rows[i].Label)
		}
	}

	if rows := app.ContextMenuFor("tab-details"); len(rows) != 0 {
		t.Fatalf("tab-details rows = %d, want none", len(rows))
	}

	if rows := app.ContextMenuFor("editor"); len(rows) != 0 {
		t.Fatalf("editor rows = %d, want none", len(rows))
	}
}

func TestContextMenuCloseOthersNeedsTwoTabs(t *testing.T) {
	app := newTestApp(t)

	rows := app.ContextMenuFor("tab-0")
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(rows))
	}

	if rows[1].Enabled {
		t.Fatal("Close Others is enabled with one tab")
	}
}

func TestContextMenuCloseAll(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")

	rows := app.ContextMenuFor("tab-0")
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(rows))
	}

	if err := app.MenuAction(context.Background(), rows[2].ID); err != nil {
		t.Fatalf("MenuAction: %v", err)
	}

	if len(app.tabs) != 0 {
		t.Fatalf("tabs = %v, want empty", app.tabs)
	}

	if app.active != 0 {
		t.Fatalf("active = %d, want 0", app.active)
	}

	if app.View().Tabs != nil && len(app.View().Tabs) != 0 {
		t.Fatalf("view tabs = %v, want empty", app.View().Tabs)
	}
}
