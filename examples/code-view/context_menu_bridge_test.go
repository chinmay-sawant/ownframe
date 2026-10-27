package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// pageMenuBridge is the public menu surface the window needs. The host
// interfaces live in an internal package, so the example names the two
// methods it expects the public page to carry.
type pageMenuBridge interface {
	ContextMenuFor(id string) []ownframe.MenuItem
	MenuAction(ctx context.Context, id string) error
}

// TestContextMenuBridgeTabRows reads the app's rows through the handler
// conversion, then asks the public page for the same tabs. The public page
// has to expose the callbacks registered through Handle for the second half
// to be reachable from an example test.
func TestContextMenuBridgeTabRows(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")
	app.openFile("web/index.html")

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatalf("redraw: %v", err)
	}

	rows := app.contextMenuForPage("tab-1")
	if len(rows) != 3 {
		t.Fatalf("handler rows = %d, want 3", len(rows))
	}

	if !strings.HasPrefix(rows[2].ID, "code-view:tab-close-all") {
		t.Fatalf("Close All id = %q, want the code-view:tab-close-all prefix", rows[2].ID)
	}

	bridge, ok := any(app.Page()).(pageMenuBridge)
	if !ok {
		t.Fatalf("ownframe.Page has no ContextMenuFor/MenuAction: the tab menu registered through Page.Handle is unreachable through the public page bridge")
	}

	rows = bridge.ContextMenuFor("tab-1")
	if len(rows) != 3 {
		t.Fatalf("page rows = %d, want 3", len(rows))
	}

	if !strings.HasPrefix(rows[2].ID, "code-view:tab-close-all") {
		t.Fatalf("page Close All id = %q, want the code-view:tab-close-all prefix", rows[2].ID)
	}
}

// TestContextMenuBridgeCloseAllGroupOne closes the second editor group
// through the public menu bridge. An empty second group has to end the split
// and return focus to the left pane.
func TestContextMenuBridgeCloseAllGroupOne(t *testing.T) {
	ctx := context.Background()
	app := newTestApp(t)
	app.openFile("backend/probe.go")
	app.toggleSplit()
	app.openFileIn(1, "README.md")
	app.openFileIn(1, "backend/main.go")

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("redraw: %v", err)
	}

	if !app.split || len(app.tabs2) != 3 {
		t.Fatalf("second group = split %v with %d tabs, want an open split with 3 tabs", app.split, len(app.tabs2))
	}

	rows := app.contextMenuForPage("tab2-0")
	if len(rows) != 3 {
		t.Fatalf("handler rows = %d, want 3", len(rows))
	}

	closeAll := rows[2]
	if !strings.HasPrefix(closeAll.ID, "code-view:tab-close-all") {
		t.Fatalf("Close All id = %q, want the code-view:tab-close-all prefix", closeAll.ID)
	}

	if !closeAll.Enabled {
		t.Fatal("Close All is disabled with three tabs")
	}

	bridge, ok := any(app.Page()).(pageMenuBridge)
	if !ok {
		t.Errorf("ownframe.Page has no ContextMenuFor/MenuAction: Close All cannot be driven through the public page bridge")
	}

	switch {
	case ok:
		if err := bridge.MenuAction(ctx, closeAll.ID); err != nil {
			t.Fatalf("page MenuAction: %v", err)
		}
	default:
		// The handler conversion already proved the row shape. Run the same
		// callback directly so this test still reports whether Close All
		// clears the second group.
		if err := app.MenuAction(ctx, closeAll.ID); err != nil {
			t.Fatalf("app MenuAction: %v", err)
		}
	}

	if len(app.tabs2) != 0 {
		t.Fatalf("second group tabs = %v, want empty", app.tabs2)
	}

	if app.active2 != 0 {
		t.Fatalf("second group active = %d, want 0", app.active2)
	}

	if app.split {
		t.Fatal("split is still open after Close All")
	}

	if app.focusGroup != 0 {
		t.Fatalf("focus group = %d, want 0", app.focusGroup)
	}

	if app.View().Split {
		t.Fatal("view still reports a split after Close All")
	}

	if len(app.tabs) == 0 {
		t.Fatal("Close All on group 1 also cleared group 0")
	}
}
