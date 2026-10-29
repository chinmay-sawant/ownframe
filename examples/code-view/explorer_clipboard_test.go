package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestExplorerMenuFileRows(t *testing.T) {
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)

	rows := app.ContextMenuFor("tree-backend/probe.go")
	if len(rows) != 2 {
		t.Fatalf("file rows = %d, want 2", len(rows))
	}

	if rows[0].Label != "Copy Path" || rows[0].ID != "code-view:copy-path:backend/probe.go" || !rows[0].Enabled {
		t.Fatalf("copy-path row = %+v", rows[0])
	}

	if rows[1].Label != "Copy File" || rows[1].ID != "code-view:copy-file:backend/probe.go" || !rows[1].Enabled {
		t.Fatalf("copy-file row = %+v", rows[1])
	}
}

func TestExplorerMenuDirRows(t *testing.T) {
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)

	rows := app.ContextMenuFor("tree-backend")
	if len(rows) != 2 {
		t.Fatalf("dir rows = %d, want 2", len(rows))
	}

	if rows[0].Label != "Copy Path" || rows[0].ID != "code-view:copy-path:backend" || !rows[0].Enabled {
		t.Fatalf("copy-path row = %+v", rows[0])
	}

	if rows[1].Label != "Paste" || rows[1].ID != "code-view:paste:backend" {
		t.Fatalf("paste row = %+v", rows[1])
	}

	if rows[1].Enabled {
		t.Fatal("paste is enabled with nothing staged")
	}
}

func TestExplorerMenuRejectsStaleAndControlIDs(t *testing.T) {
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)

	for _, id := range []string{"tree-nope/missing.go", "tree-scroll", "tree-thumb", "tree-up", "tree-down"} {
		if rows := app.ContextMenuFor(id); len(rows) != 0 {
			t.Fatalf("%s rows = %+v, want none", id, rows)
		}
	}
}

func TestExplorerMenuRootPasteOnly(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)

	for _, id := range []string{"tree-pane", "sidebar", ""} {
		rows := app.ContextMenuFor(id)
		if len(rows) != 1 || rows[0].Label != "Paste" || rows[0].ID != "code-view:paste:" {
			t.Fatalf("%q rows = %+v, want one paste row", id, rows)
		}

		if rows[0].Enabled {
			t.Fatalf("%q paste is enabled with nothing staged", id)
		}
	}

	app.copyTreeFile("backend/probe.go")

	for _, id := range []string{"tree-pane", "sidebar", ""} {
		rows := app.ContextMenuFor(id)
		if len(rows) != 1 || !rows[0].Enabled {
			t.Fatalf("%q paste stays disabled after a copy: %+v", id, rows)
		}
	}
}

func TestExplorerCopyPathClipboard(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)

	if err := app.MenuAction(context.Background(), "code-view:copy-path:backend/probe.go"); err != nil {
		t.Fatalf("MenuAction: %v", err)
	}

	got, err := ownframe.Read()
	if err != nil || got != "backend/probe.go" {
		t.Fatalf("clipboard = %q (%v), want the slash-relative path", got, err)
	}
}

func TestExplorerCopyFileClipboard(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)

	if err := app.MenuAction(context.Background(), "code-view:copy-file:backend/probe.go"); err != nil {
		t.Fatalf("MenuAction: %v", err)
	}

	want, err := app.ws.readFile("backend/probe.go")
	if err != nil {
		t.Fatalf("readFile: %v", err)
	}

	got, err := ownframe.Read()
	if err != nil || got != string(want) {
		t.Fatalf("clipboard = %d bytes (%v), want %d", len(got), err, len(want))
	}

	if staged := stagedCopy(); staged == nil || staged.SrcPath != "backend/probe.go" {
		t.Fatalf("staged = %+v, want backend/probe.go", staged)
	}
}

func TestExplorerPasteDiskCopy(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)
	clearStagedCopy()
	defer clearStagedCopy()

	dir := t.TempDir()
	writeFile(t, dir, "a.go", "package main\n")
	writeFile(t, dir, "sub/b.go", "package sub\n")

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatalf("redraw: %v", err)
	}

	ctx := context.Background()

	if err := app.MenuAction(ctx, "code-view:copy-file:a.go"); err != nil {
		t.Fatalf("copy: %v", err)
	}

	if err := app.MenuAction(ctx, "code-view:paste:sub"); err != nil {
		t.Fatalf("paste: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "sub", "a.go")) //nolint:gosec // path is a t.TempDir fixture
	if err != nil {
		t.Fatalf("pasted file: %v", err)
	}

	if string(got) != "package main\n" {
		t.Fatalf("pasted bytes = %q", got)
	}

	if app.ws.ByPath["sub/a.go"] == nil {
		t.Fatal("sub/a.go missing from the workspace")
	}

	if err := app.MenuAction(ctx, "code-view:paste:sub"); err != nil {
		t.Fatalf("second paste: %v", err)
	}

	dup, err := os.ReadFile(filepath.Join(dir, "sub", "a copy.go")) //nolint:gosec // path is a t.TempDir fixture
	if err != nil {
		t.Fatalf("duplicate paste: %v", err)
	}

	if string(dup) != "package main\n" {
		t.Fatalf("duplicate bytes = %q", dup)
	}

	if err := app.MenuAction(ctx, "code-view:paste:"); err != nil {
		t.Fatalf("root paste: %v", err)
	}

	root, err := os.ReadFile(filepath.Join(dir, "a copy.go")) //nolint:gosec // path is a t.TempDir fixture
	if err != nil {
		t.Fatalf("root paste: %v", err)
	}

	if string(root) != "package main\n" {
		t.Fatalf("root bytes = %q", root)
	}
}

func TestExplorerPasteSampleGuard(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)
	before := len(app.ws.Files)

	if err := app.MenuAction(context.Background(), "code-view:copy-file:backend/probe.go"); err != nil {
		t.Fatalf("copy: %v", err)
	}

	if err := app.MenuAction(context.Background(), "code-view:paste:"); err != nil {
		t.Fatalf("paste: %v", err)
	}

	if toast := app.View().Toast; !toast.Show || toast.Kind != "error" {
		t.Fatalf("toast = %+v, want a visible error", toast)
	}

	if len(app.ws.Files) != before {
		t.Fatal("sample workspace changed after a guarded paste")
	}
}

func TestExplorerMenuPaintsInPage(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)
	ctx := context.Background()

	rows := app.contextMenuForPage("tree-backend/probe.go")
	if len(rows) != 2 {
		t.Fatalf("menu rows = %d, want 2", len(rows))
	}

	if !app.presentMenu(40, 50, rows) || !app.View().MenuOpen {
		t.Fatal("page menu did not open")
	}

	box, ok := boxByID(app.Page().Boxes(), "ctx-0")
	if !ok {
		t.Fatal("menu row was not painted")
	}

	if !strings.Contains(box.Action, "code-view:copy-path:backend/probe.go") {
		t.Fatalf("ctx-0 action = %q, want the copy-path id", box.Action)
	}

	if err := app.onClick(ctx, ownframe.Box{ID: box.ID, Action: box.Action}); err != nil {
		t.Fatalf("click: %v", err)
	}

	if app.menuOpen || app.View().MenuOpen {
		t.Fatal("menu stayed open after the row ran")
	}

	got, err := ownframe.Read()
	if err != nil || got != "backend/probe.go" {
		t.Fatalf("clipboard = %q (%v), want the slash-relative path", got, err)
	}
}
