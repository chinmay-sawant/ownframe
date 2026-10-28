package main

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// TestFeatureEditAndSave checks the custom editor buffer writes to disk.
func TestFeatureEditAndSave(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.openFile("main.go")
	app.toggleEdit()

	if !app.View().Editing || app.View().EditText == "" {
		t.Fatalf("edit view = %+v", app.View().Editing)
	}

	ctx := context.Background()
	if err := app.onSelectAll(ctx); err != nil {
		t.Fatal(err)
	}

	if err := app.onPaste(ctx, "package main\n\nfunc main() { println(7) }\n"); err != nil {
		t.Fatal(err)
	}

	if !app.View().EditDirty {
		t.Fatal("edit buffer not marked dirty")
	}

	app.saveEdit()

	raw, err := os.ReadFile(filepath.Join(dir, "main.go")) //nolint:gosec // path is a t.TempDir fixture
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(raw), "println(7)") {
		t.Fatalf("saved file = %q", raw)
	}

	if app.View().EditDirty || app.dirty["main.go"] {
		t.Fatal("save left the dirty flag set")
	}
}

// TestFeatureEditSelection checks the editor accepts select-all and copy.
func TestFeatureEditSelection(t *testing.T) {
	app := newTestApp(t)
	app.toggleEdit()
	ctx := context.Background()

	if err := app.onSelectAll(ctx); err != nil {
		t.Fatal(err)
	}

	text, ok, err := app.onCopy(ctx)
	if err != nil || !ok || !strings.Contains(text, "# Beacon") {
		t.Fatalf("copy = %q ok=%v err=%v", text, ok, err)
	}
}

// TestFeatureFileNav checks the bottom-center widget moves through changes.
func TestFeatureFileNav(t *testing.T) {
	app := newTestApp(t)

	view := app.View()
	if !view.FileNav.Open || view.FileNav.Total != 7 {
		t.Fatalf("file nav = %+v", view.FileNav)
	}

	first := view.FileNav.Name

	if err := app.onClick(context.Background(), ownframe.Box{Action: "file-nav-next"}); err != nil {
		t.Fatal(err)
	}

	view = app.View()
	if view.FileNav.Name == first {
		t.Fatalf("next did not move: %+v", view.FileNav)
	}

	if baseName(view.Active) != view.FileNav.Name {
		t.Fatalf("active = %q file nav = %+v", view.Active, view.FileNav)
	}
}

// TestFeatureChangedLineMarkers checks a modified open file carries gutter
// marks and that the marker opens the line card.
func TestFeatureChangedLineMarkers(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	line := 0

	for _, row := range app.View().Rows {
		if row.Mark != "" {
			line = row.ChangeLine
			break
		}
	}

	if line == 0 {
		t.Fatal("open modified file has no gutter markers")
	}

	box, ok := boxByID(app.Page().Boxes(), "gutter-"+strconv.Itoa(line))
	if !ok {
		t.Fatalf("no gutter box for line %d", line)
	}

	if err := app.Page().Click(context.Background(), box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatal(err)
	}

	card := app.View().InlineCard
	if !card.Show || card.Title != "Line "+strconv.Itoa(line) {
		t.Fatalf("inline card = %+v", card)
	}

	if card.Left <= chromeLeft {
		t.Fatalf("card left = %d, want inside the editor", card.Left)
	}
}

// TestFeatureCommitTipFloatsBesideHistory checks the hover card leaves the rail.
func TestFeatureCommitTipFloatsBesideHistory(t *testing.T) {
	app := newTestApp(t)
	clickID(t, app, "act-changes")

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	short := app.View().History.Commits[0].Short
	box, ok := boxByID(app.Page().Boxes(), "commit-"+short)
	if !ok {
		t.Fatal("no commit row")
	}

	if err := app.Page().Edit().Hover(context.Background(), box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatal(err)
	}

	tip := app.View().CommitTip
	if !tip.Show || len(tip.Lines) == 0 {
		t.Fatalf("float tip = %+v", tip)
	}

	if tip.Left <= activityW+sideWidth {
		t.Fatalf("tip left = %d, want it right of the sidebar", tip.Left)
	}
}

// TestFeatureInlineCardSteps checks the line card stops at the first and last
// changed line in the open file.
func TestFeatureInlineCardSteps(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")

	lines := inlineChangeLines(app.changedLinesFor("backend/probe.go"))
	if len(lines) < 2 {
		t.Skip("sample file has fewer than two changed lines")
	}

	app.showInlineCard(lines[0])
	card := app.View().InlineCard
	if card.HasPrev || !card.HasNext {
		t.Fatalf("first card bounds = prev %v next %v", card.HasPrev, card.HasNext)
	}

	app.inlineCardStep(-1)
	if app.View().InlineCard.Line != lines[0] {
		t.Fatal("previous moved before the first change")
	}

	for range len(lines) {
		app.inlineCardStep(1)
	}

	card = app.View().InlineCard
	if card.Line != lines[len(lines)-1] || card.HasNext {
		t.Fatalf("last card = line %d next %v, want line %d next false", card.Line, card.HasNext, lines[len(lines)-1])
	}

	if !card.HasPrev {
		t.Fatal("last card should still allow previous")
	}
}

// TestFeatureSampleReadOnly checks the bundled workspace refuses a write.
func TestFeatureSampleReadOnly(t *testing.T) {
	app := newTestApp(t)
	app.toggleEdit()
	ctx := context.Background()

	if err := app.onSelectAll(ctx); err != nil {
		t.Fatal(err)
	}

	if err := app.onPaste(ctx, "# changed\n"); err != nil {
		t.Fatal(err)
	}

	app.saveEdit()

	if !app.View().Toast.Show || app.View().Toast.Kind != "error" {
		t.Fatalf("sample save toast = %+v", app.View().Toast)
	}
}

// TestFeatureWorkspaceSearch opens a content hit and moves the editor there.
func TestFeatureWorkspaceSearch(t *testing.T) {
	app := newTestApp(t)

	if err := app.openWorkspaceSearch(context.Background()); err != nil {
		t.Fatal(err)
	}

	app.searchQuery = "Beacon"
	app.runWorkspaceSearch(context.Background())

	deadline := time.Now().Add(3 * time.Second)

	for time.Now().Before(deadline) {
		if err := app.tickWatch(context.Background()); err != nil {
			t.Fatal(err)
		}

		if !app.View().Search.Busy && len(app.View().Search.Rows) > 0 {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	if len(app.View().Search.Rows) == 0 {
		t.Fatal("workspace search returned no rows")
	}

	if got := app.View().Search.Rows[0].Path; got != "README.md" {
		t.Fatalf("first hit = %q, want README.md", got)
	}

	app.openSearchHit(0)

	if app.View().Active != "README.md" || app.View().Search.Open {
		t.Fatalf("after open active=%q searchOpen=%v", app.View().Active, app.View().Search.Open)
	}
}

// TestFeatureSplitEditor checks two independent tab lists and panes.
func TestFeatureSplitEditor(t *testing.T) {
	app := newTestApp(t)
	app.toggleSplit()

	view := app.View()
	if !view.Split || !view.Second.Open || view.Second.Active != "README.md" {
		t.Fatalf("split view = %+v second = %+v", view.Split, view.Second)
	}

	app.openFileIn(1, "web/theme.css")

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	view = app.View()
	if view.Second.Active != "web/theme.css" || len(view.Second.Tabs) != 2 {
		t.Fatalf("second pane = %+v", view.Second)
	}

	if _, ok := boxByID(app.Page().Boxes(), "editor-window-2"); !ok {
		t.Fatal("no second editor box")
	}

	app.closeTabIn(1, app.active2)

	if view = app.View(); !view.Split || view.Second.Active != "README.md" {
		t.Fatalf("closing one tab closed the split: %+v", view.Second)
	}

	app.closeTabIn(1, app.active2)

	if app.View().Split {
		t.Fatal("closing the last second-group tab left the split open")
	}
}

// TestFeatureScanBudget checks the visible budget string and status box.
func TestFeatureScanBudget(t *testing.T) {
	dir := t.TempDir()
	writeScanFiles(t, dir, 5)

	app, err := NewWithLimit(dir, 2)
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if app.View().ScanBudget == "" {
		t.Fatal("scan budget is empty")
	}

	if _, ok := boxByID(app.Page().Boxes(), "sb-scan"); !ok {
		t.Fatal("no scan budget status item")
	}
}

// TestFeatureDiffWrapIntegration checks the changes view follows the wrap
// toggle instead of clipping long rows.
func TestFeatureDiffWrapIntegration(t *testing.T) {
	app := newTestApp(t)
	app.setMode(context.Background(), "changes")
	app.setSplit(true)

	unified, split := app.diffColumns()
	cols := unified
	if app.View().Diff.Split {
		cols = split
	}

	for _, row := range app.View().Diff.Rows {
		if row.Kind == "split" {
			if len([]rune(lineText(row.Left))) > cols || len([]rune(lineText(row.Right))) > cols {
				t.Fatalf("split row exceeds %d columns: %+v", cols, row)
			}

			continue
		}

		if len([]rune(lineText(row.Left))) > cols || len([]rune(row.Hunk)) > cols {
			t.Fatalf("unified row exceeds %d columns: %+v", cols, row)
		}
	}

	app.wrap = false
	app.refresh()

	if len(app.View().Diff.Rows) == 0 {
		t.Fatal("diff disappeared with wrap off")
	}
}

// TestFeatureSessionRestoresLayout checks a second App picks up the first one.
func TestFeatureSessionRestoresLayout(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")
	writeFile(t, dir, "other.go", "package main\n")

	first, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	first.EnableSession()
	first.openFileIn(0, "other.go")
	first.toggleDetails()
	first.toggleSplit()
	first.openFileIn(1, "main.go")
	first.refresh()

	second, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	second.EnableSession()

	view := second.View()
	// The strip windows full-width tabs, so assert the restored tab state
	// rather than the rendered window.
	if len(second.tabs) != 2 || second.tabs[1] != "other.go" {
		t.Fatalf("restored tabs = %+v", second.tabs)
	}

	if !view.Details || !view.Split || view.Second.Active != "main.go" {
		t.Fatalf("restored layout details=%v split=%v second=%q", view.Details, view.Split, view.Second.Active)
	}
}

// TestFeatureModalFocusTrap checks Tab cannot leave an open overlay.
func TestFeatureModalFocusTrap(t *testing.T) {
	app := newTestApp(t)

	if err := app.openPalette(context.Background()); err != nil {
		t.Fatal(err)
	}

	if !app.View().Modal {
		t.Fatal("palette did not set Modal")
	}

	seen := map[string]bool{}

	for range 16 {
		if err := app.Page().Form().FocusNext(context.Background()); err != nil {
			t.Fatal(err)
		}

		id := app.Page().Form().FocusID()
		if id == "" {
			continue
		}

		seen[id] = true

		if !strings.HasPrefix(id, "palette") {
			t.Fatalf("tab escaped the palette to %q", id)
		}
	}

	if len(seen) == 0 {
		t.Fatal("no palette control took focus")
	}
}
