package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func wantExplorerRows(t *testing.T, app *App, id string, labels []string) []contextMenuRow {
	t.Helper()

	rows := app.ContextMenuFor(id)
	if len(rows) != len(labels) {
		t.Fatalf("%s rows = %+v, want %d rows", id, rows, len(labels))
	}

	for i, want := range labels {
		if rows[i].Label != want {
			t.Fatalf("%s row %d = %+v, want label %q", id, i, rows[i], want)
		}
	}

	return rows
}

func TestExplorerMenuFileRows(t *testing.T) {
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)

	labels := []string{"Cut", "Copy", "Copy Path", "Copy Absolute Path", "Paste"}
	rows := wantExplorerRows(t, app, "tree-backend/probe.go", labels)

	wantIDs := []string{
		"code-view:cut:backend/probe.go",
		"code-view:copy:backend/probe.go",
		"code-view:copy-path:backend/probe.go",
		"code-view:copy-abs:backend/probe.go",
		"code-view:paste:backend/probe.go",
	}

	for i, want := range wantIDs {
		if rows[i].ID != want {
			t.Fatalf("row %d = %+v, want id %q", i, rows[i], want)
		}

		if i < 4 && !rows[i].Enabled {
			t.Fatalf("row %d = %+v, want enabled", i, rows[i])
		}
	}

	if rows[4].Enabled {
		t.Fatal("paste is enabled with nothing staged")
	}
}

func TestExplorerMenuDirRows(t *testing.T) {
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)

	labels := []string{"Cut", "Copy", "Copy Path", "Copy Absolute Path", "Paste"}
	rows := wantExplorerRows(t, app, "tree-backend", labels)

	if rows[0].ID != "code-view:cut:backend" || rows[1].ID != "code-view:copy:backend" {
		t.Fatalf("cut/copy ids = %q %q", rows[0].ID, rows[1].ID)
	}

	if rows[2].ID != "code-view:copy-path:backend" || rows[3].ID != "code-view:copy-abs:backend" {
		t.Fatalf("path ids = %q %q", rows[2].ID, rows[3].ID)
	}

	if rows[4].ID != "code-view:paste:backend" || rows[4].Enabled {
		t.Fatalf("paste row = %+v, want disabled", rows[4])
	}
}

func TestExplorerMenuHints(t *testing.T) {
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)

	rows := app.ContextMenuFor("tree-backend/probe.go")
	if len(rows) != 5 {
		t.Fatalf("rows = %d, want 5", len(rows))
	}

	if rows[0].Hint != shortcutHint("Ctrl+X") || rows[0].Hint == "" {
		t.Fatalf("cut hint = %q", rows[0].Hint)
	}

	if rows[1].Hint != shortcutHint("Ctrl+C") || rows[1].Hint == "" {
		t.Fatalf("copy hint = %q", rows[1].Hint)
	}

	if rows[2].Hint != "" || rows[3].Hint != "" {
		t.Fatalf("path rows carry hints: %+v", rows[2:])
	}

	if rows[4].Hint != shortcutHint("Ctrl+V") || rows[4].Hint == "" {
		t.Fatalf("paste hint = %q", rows[4].Hint)
	}

	// The hint survives the public bridge the page paints from.
	bridged := app.contextMenuForPage("tree-backend/probe.go")
	if len(bridged) != 5 || bridged[1].Hint != rows[1].Hint || bridged[4].Hint != rows[4].Hint {
		t.Fatalf("bridged hints = %+v", bridged)
	}
}

func TestExplorerMenuRejectsStaleAndControlIDs(t *testing.T) {
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)

	for _, id := range []string{"tree-nope/missing.go", "tree-scroll", "tree-thumb", "tree-up", "tree-down", "tree-sticky"} {
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

		if rows[0].Hint != shortcutHint("Ctrl+V") {
			t.Fatalf("%q paste hint = %q", id, rows[0].Hint)
		}

		if rows[0].Enabled {
			t.Fatalf("%q paste is enabled with nothing staged", id)
		}
	}

	app.stageTreeCopy([]string{"backend/probe.go"}, false)

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

func TestExplorerCopyAbsPathClipboard(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)

	if err := app.MenuAction(context.Background(), "code-view:copy-abs:backend/probe.go"); err != nil {
		t.Fatalf("MenuAction: %v", err)
	}

	got, err := ownframe.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if !filepath.IsAbs(got) {
		t.Fatalf("clipboard = %q, want an absolute path", got)
	}

	if !strings.HasSuffix(filepath.ToSlash(got), "backend/probe.go") {
		t.Fatalf("clipboard = %q, want it to end in backend/probe.go", got)
	}
}

func TestExplorerCopyStagesSingleFile(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)

	if err := app.MenuAction(context.Background(), "code-view:copy:backend/probe.go"); err != nil {
		t.Fatalf("MenuAction: %v", err)
	}

	staged := stagedCopy()
	if staged == nil || len(staged.SrcPaths) != 1 || staged.SrcPath != "backend/probe.go" {
		t.Fatalf("staged = %+v, want backend/probe.go", staged)
	}

	if staged.IsCut {
		t.Fatal("copy staged a cut")
	}

	got, err := ownframe.Read()
	if err != nil || got != "backend/probe.go" {
		t.Fatalf("clipboard = %q (%v), want the staged path", got, err)
	}
}

func TestExplorerCutStagesMove(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)

	if err := app.MenuAction(context.Background(), "code-view:cut:backend/probe.go"); err != nil {
		t.Fatalf("MenuAction: %v", err)
	}

	staged := stagedCopy()
	if staged == nil || staged.SrcPath != "backend/probe.go" || !staged.IsCut {
		t.Fatalf("staged = %+v, want a cut of backend/probe.go", staged)
	}
}

func TestExplorerCopyStagesMultiSelect(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)
	app.explSel = map[string]bool{"backend/probe.go": true, "backend/store.go": true}

	if err := app.MenuAction(context.Background(), "code-view:copy:backend/probe.go"); err != nil {
		t.Fatalf("MenuAction: %v", err)
	}

	staged := stagedCopy()
	if staged == nil || len(staged.SrcPaths) != 2 {
		t.Fatalf("staged = %+v, want both selected files", staged)
	}

	got, err := ownframe.Read()
	if err != nil || !strings.Contains(got, "backend/probe.go") || !strings.Contains(got, "backend/store.go") {
		t.Fatalf("clipboard = %q (%v), want both paths", got, err)
	}
}

func TestExplorerRightClickReplacesOutsideSelection(t *testing.T) {
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)
	app.explSel = map[string]bool{"backend/probe.go": true}
	app.explAnchor = "backend/probe.go"

	rows := app.ContextMenuFor("tree-backend/store.go")
	if len(rows) != 5 {
		t.Fatalf("rows = %d, want 5", len(rows))
	}

	if len(app.explSel) != 1 || !app.explSel["backend/store.go"] {
		t.Fatalf("selection = %+v, want only backend/store.go", app.explSel)
	}
}

func TestExplorerCutPasteMovesFile(t *testing.T) {
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

	if err := app.MenuAction(ctx, "code-view:cut:a.go"); err != nil {
		t.Fatalf("cut: %v", err)
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

	if _, err := os.Stat(filepath.Join(dir, "a.go")); !os.IsNotExist(err) {
		t.Fatalf("cut source still on disk: %v", err)
	}

	if stagedCopy() != nil {
		t.Fatal("cut staging survived the paste")
	}

	if len(app.explSel) != 0 {
		t.Fatalf("selection survived the move: %+v", app.explSel)
	}
}

func TestExplorerMultiPasteDiskCopies(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)
	clearStagedCopy()
	defer clearStagedCopy()

	dir := t.TempDir()
	writeFile(t, dir, "a.go", "package main\n")
	writeFile(t, dir, "b.go", "package b\n")

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatalf("redraw: %v", err)
	}

	ctx := context.Background()
	app.explSel = map[string]bool{"a.go": true, "b.go": true}

	if err := app.MenuAction(ctx, "code-view:copy:a.go"); err != nil {
		t.Fatalf("copy: %v", err)
	}

	rels, err := app.PastePendingCopies("")
	if err != nil {
		t.Fatalf("paste: %v", err)
	}

	if len(rels) != 2 {
		t.Fatalf("pasted = %+v, want 2 files", rels)
	}

	for _, rel := range []string{"a (copy).go", "b (copy).go"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}
}

func TestExplorerDirCopyPaste(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)
	clearStagedCopy()
	defer clearStagedCopy()

	dir := t.TempDir()
	writeFile(t, dir, "pkg/x.go", "package pkg\n")

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatalf("redraw: %v", err)
	}

	ctx := context.Background()

	if err := app.MenuAction(ctx, "code-view:copy:pkg"); err != nil {
		t.Fatalf("copy: %v", err)
	}

	rel, err := app.PastePendingCopy("")
	if err != nil {
		t.Fatalf("paste: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, rel, "x.go")) //nolint:gosec // path is a t.TempDir fixture
	if err != nil || string(got) != "package pkg\n" {
		t.Fatalf("pasted dir file = %q (%v)", got, err)
	}
}

func TestExplorerSelectToggleAndRange(t *testing.T) {
	clearStagedCopy()
	defer clearStagedCopy()

	dir := t.TempDir()
	writeFile(t, dir, "a.go", "package main\n")
	writeFile(t, dir, "b.go", "package b\n")
	writeFile(t, dir, "c.go", "package c\n")

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx := context.Background()

	if err := app.clickTreeSelect(ctx, "a.go"); err != nil {
		t.Fatalf("toggle a: %v", err)
	}

	if err := app.clickTreeSelect(ctx, "c.go"); err != nil {
		t.Fatalf("toggle c: %v", err)
	}

	if len(app.explSel) != 2 || !app.explSel["a.go"] || !app.explSel["c.go"] {
		t.Fatalf("selection = %+v, want a.go and c.go", app.explSel)
	}

	app.clearExplorerSel()
	app.explAnchor = "a.go"

	app.shift = true
	if err := app.clickTreeSelect(ctx, "c.go"); err != nil {
		t.Fatalf("range: %v", err)
	}
	app.shift = false

	for _, want := range []string{"a.go", "b.go", "c.go"} {
		if !app.explSel[want] {
			t.Fatalf("selection = %+v, want the a..c range", app.explSel)
		}
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

	if err := app.MenuAction(ctx, "code-view:copy:a.go"); err != nil {
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

	dup, err := os.ReadFile(filepath.Join(dir, "sub", "a (copy).go")) //nolint:gosec // path is a t.TempDir fixture
	if err != nil {
		t.Fatalf("duplicate paste: %v", err)
	}

	if string(dup) != "package main\n" {
		t.Fatalf("duplicate bytes = %q", dup)
	}

	if err := app.MenuAction(ctx, "code-view:paste:"); err != nil {
		t.Fatalf("root paste: %v", err)
	}

	root, err := os.ReadFile(filepath.Join(dir, "a (copy).go")) //nolint:gosec // path is a t.TempDir fixture
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

	if err := app.MenuAction(context.Background(), "code-view:copy:backend/probe.go"); err != nil {
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
	if len(rows) != 5 {
		t.Fatalf("menu rows = %d, want 5", len(rows))
	}

	if !app.presentMenu(40, 50, rows) || !app.View().MenuOpen {
		t.Fatal("page menu did not open")
	}

	box, ok := boxByID(app.Page().Boxes(), "ctx-0")
	if !ok {
		t.Fatal("menu row was not painted")
	}

	if !strings.Contains(box.Action, "code-view:cut:backend/probe.go") {
		t.Fatalf("ctx-0 action = %q, want the cut id", box.Action)
	}

	if err := app.onClick(ctx, ownframe.Box{ID: box.ID, Action: box.Action}); err != nil {
		t.Fatalf("click: %v", err)
	}

	if app.menuOpen || app.View().MenuOpen {
		t.Fatal("menu stayed open after the row ran")
	}

	staged := stagedCopy()
	if staged == nil || !staged.IsCut || staged.SrcPath != "backend/probe.go" {
		t.Fatalf("staged = %+v, want a cut of backend/probe.go", staged)
	}
}

func TestWslMntToWindows(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want string
		ok   bool
	}{
		{"/mnt/c/Users/a/b.go", `C:\Users\a\b.go`, true},
		{"/mnt/d", `D:\`, true},
		{"/mnt/c/", `C:\`, true},
		{"/home/u/b.go", "", false},
		{"/mnt", "", false},
		{"/mnt/cd/x", "", false},
		{"/mnt/1/x", "", false},
		{"C:\\x", "", false},
		{"", "", false},
	} {
		got, ok := wslMntToWindows(tc.in)
		if got != tc.want || ok != tc.ok {
			t.Fatalf("wslMntToWindows(%q) = %q, %v; want %q, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}
