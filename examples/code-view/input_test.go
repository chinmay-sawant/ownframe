package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func clickID(t *testing.T, app *App, id string) {
	t.Helper()

	box, ok := boxByID(app.Page().Boxes(), id)
	if !ok {
		t.Fatalf("no box %s", id)
	}

	if err := app.Page().Click(context.Background(), box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatalf("click %s: %v", id, err)
	}
}

func TestPointerControls(t *testing.T) {
	app := newTestApp(t)

	clickID(t, app, "act-changes")

	if app.View().Mode != "changes" {
		t.Fatalf("changes mode = %q", app.View().Mode)
	}

	clickID(t, app, "act-explorer")

	if app.View().Mode != "explorer" {
		t.Fatalf("explorer mode = %q", app.View().Mode)
	}

	clickID(t, app, "tree-backend/probe.go")

	if app.View().Active != "backend/probe.go" || len(app.View().Tabs) != 2 {
		t.Fatalf("file click active=%q tabs=%d", app.View().Active, len(app.View().Tabs))
	}

	clickID(t, app, "act-collapse")

	if len(app.folders) != 0 {
		t.Fatalf("folders = %v", app.folders)
	}

	clickID(t, app, "act-details")

	if !app.View().Details {
		t.Fatal("details not open")
	}

	clickID(t, app, "det-close")

	if app.View().Details {
		t.Fatal("details not closed")
	}

	clickID(t, app, "act-changes")

	if app.View().Mode != "changes" {
		t.Fatalf("changes mode = %q", app.View().Mode)
	}

	clickID(t, app, "change-web/theme.css")

	if app.View().Diff.Selected != "web/theme.css" {
		t.Fatalf("selected = %q", app.View().Diff.Selected)
	}

	clickID(t, app, "seg-split")

	if !app.View().Diff.Split {
		t.Fatal("split not on")
	}

	clickID(t, app, "sb-devtools")

	if app.devOn {
		t.Fatal("devtools still on")
	}

	clickID(t, app, "act-explorer")
	clickID(t, app, "tab-close-1")

	if len(app.View().Tabs) != 1 {
		t.Fatalf("tabs = %d", len(app.View().Tabs))
	}
}

func TestQuickOpenPalette(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	clickID(t, app, "cmdbar")

	if !app.View().Palette.Open {
		t.Fatal("palette did not open")
	}

	if app.Page().Form().FocusID() != "palette" {
		t.Fatalf("focus = %q", app.Page().Form().FocusID())
	}

	if len(app.View().Palette.Rows) != paletteDefaultRows {
		t.Fatalf("default rows = %d", len(app.View().Palette.Rows))
	}

	app.view.PaletteQuery = "probe"

	if err := app.onChange(ctx, ownframe.Box{ID: "palette"}); err != nil {
		t.Fatal(err)
	}

	rows := app.View().Palette.Rows
	if len(rows) == 0 || rows[0].Path != "backend/probe.go" {
		t.Fatalf("query rows = %+v", rows)
	}

	clickID(t, app, "palette-backend/probe.go")

	if app.View().Palette.Open {
		t.Fatal("palette stayed open")
	}

	clickID(t, app, "cmdbar")

	if !app.View().Palette.Open {
		t.Fatal("palette did not reopen")
	}

	back, ok := boxByID(app.Page().Boxes(), "palette-back")
	if !ok {
		t.Fatal("no palette backdrop")
	}

	if err := app.Page().Click(ctx, back.X+10, back.Y+back.H/2); err != nil {
		t.Fatal(err)
	}

	if app.View().Palette.Open {
		t.Fatal("backdrop did not close the palette")
	}

	if app.View().Active != "backend/probe.go" {
		t.Fatalf("active = %q", app.View().Active)
	}

	if err := app.onKeyDown(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "p"); err != nil {
		t.Fatal(err)
	}

	if !app.View().Palette.Open {
		t.Fatal("ctrl+p did not open the palette")
	}

	if err := app.onKeyUp(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "escape"); err != nil {
		t.Fatal(err)
	}

	if app.View().Palette.Open {
		t.Fatal("escape did not close the palette")
	}
}

func TestMarkdownPreview(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	if app.View().Lang != langMarkdown {
		t.Fatalf("initial language = %s", app.View().Lang)
	}

	app.togglePreview()

	if !app.View().Preview || len(app.View().PreviewRows) == 0 {
		t.Fatalf("preview = %v rows = %d", app.View().Preview, len(app.View().PreviewRows))
	}

	first := lineText(app.View().PreviewRows[0].Segments)
	if first != "Beacon" {
		t.Fatalf("first preview row = %q", first)
	}

	app.togglePreview()

	if app.View().Preview || len(app.View().Rows) == 0 {
		t.Fatalf("source restore preview=%v rows=%d", app.View().Preview, len(app.View().Rows))
	}

	if lineText(app.View().Rows[0].Segments) != "# Beacon" {
		t.Fatalf("first source row = %q", lineText(app.View().Rows[0].Segments))
	}

	if err := app.onKeyDown(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "shift"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "v"); err != nil {
		t.Fatal(err)
	}

	if !app.View().Preview {
		t.Fatal("ctrl+shift+v did not turn preview on")
	}

	if err := app.onKeyUp(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyUp(ctx, "shift"); err != nil {
		t.Fatal(err)
	}

	app.openFile("backend/probe.go")

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if app.View().Previewable || app.View().Preview {
		t.Fatal("preview leaked onto a Go file")
	}

	if _, ok := boxByID(app.Page().Boxes(), "preview-toggle"); ok {
		t.Fatal("preview button shown for a Go file")
	}
}

func TestPreviewButton(t *testing.T) {
	app := newTestApp(t)

	clickID(t, app, "preview-toggle")

	if !app.View().Preview {
		t.Fatal("preview button did not turn preview on")
	}

	clickID(t, app, "preview-toggle")

	if app.View().Preview {
		t.Fatal("preview button did not turn preview off")
	}
}

func TestActionControlsHaveIDs(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	check := func() {
		t.Helper()

		for _, box := range app.Page().Boxes() {
			if box.Action != "" && box.ID == "" {
				t.Errorf("action %q has no id at (%.0f,%.0f)", box.Action, box.X, box.Y)
			}
		}
	}

	check()

	if err := app.onClick(ctx, ownframe.Box{Action: "mode-changes"}); err != nil {
		t.Fatal(err)
	}

	check()

	app.togglePreview()

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	check()

	if err := app.openPalette(ctx); err != nil {
		t.Fatal(err)
	}

	check()

	if err := app.closePalette(ctx); err != nil {
		t.Fatal(err)
	}

	app.toggleDetails()

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	check()

	app.showToast("info", "Check", "The toast has an id.")
	app.refresh()

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	check()
}

func TestFixedChromeClickAfterScroll(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")
	app.scrollTo(400)
	app.Page().Scroll().SetScrollOffset(0, 400)

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	tab, ok := boxByID(app.Page().Boxes(), "tab-0")
	if !ok {
		t.Fatal("no first tab")
	}

	// The window adds the page offset before hit testing a fixed box.
	if err := app.Page().Click(context.Background(), tab.X+tab.W/2, tab.Y+tab.H/2+400); err != nil {
		t.Fatal(err)
	}

	if app.active != 0 {
		t.Fatalf("tab click after scroll left active=%d", app.active)
	}
}

func TestOutlineClickScrolls(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")
	app.toggleDetails()

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	target := -1

	for _, symbol := range app.View().Outline {
		if symbol.Name == "Probe" {
			target = symbol.Line
		}
	}

	if target < 0 {
		t.Fatal("no Probe outline row")
	}

	clickID(t, app, fmt.Sprintf("outline-%d", target))

	if app.scrollY == 0 {
		t.Fatal("outline click did not scroll")
	}

	if len(app.View().Rows) == 0 || app.View().Rows[0].No == 1 {
		t.Fatalf("first row = %d", app.View().Rows[0].No)
	}
}

func TestDetailsMetadata(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")
	app.toggleDetails()

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if view.LangLabel != "Go" || view.LineCount != 170 || view.Encoding != "UTF-8" || view.Indent != "Tabs" {
		t.Fatalf("probe metadata = %+v", view)
	}

	if view.Status != "M" || len(view.Outline) == 0 {
		t.Fatalf("probe status = %q outline = %d", view.Status, len(view.Outline))
	}

	app.openFile("README.md")

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	view = app.View()
	if view.LangLabel != "Markdown" || view.Indent == "Tabs" || view.Status != "" {
		t.Fatalf("README metadata = %+v", view)
	}

	if len(view.Outline) == 0 || view.Outline[0].Name != "Beacon" {
		t.Fatalf("README outline = %+v", view.Outline)
	}
}

func TestDiffScroll(t *testing.T) {
	app := newTestApp(t)
	app.mode = "changes"
	app.changes = changeSet{Branch: "test", Note: "test"}
	visible := app.changeVisibleRows()

	for i := 0; i < visible+2; i++ {
		app.changes.Files = append(app.changes.Files, changedFile{Path: fmt.Sprintf("file-%02d.go", i), Status: "M"})
	}

	app.statuses = statusMap(app.changes)
	app.selectFirstDiff()
	app.changesScroll.SetOffset(0, len(app.changes.Files), visible)
	app.refresh()

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if !app.View().Diff.Scroll || len(app.View().Diff.Files) != visible || !app.View().Diff.AtTop {
		t.Fatalf("top = %+v", app.View().Diff)
	}

	app.scrollChanges(visible)

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if app.View().Diff.AtTop || !app.View().Diff.AtEnd || len(app.View().Diff.Files) == 0 {
		t.Fatalf("bottom = %+v", app.View().Diff)
	}

	if got := app.View().Diff.Files[0].Path; got != fmt.Sprintf("file-%02d.go", len(app.changes.Files)-visible) {
		t.Fatalf("first visible = %q", got)
	}
}

func TestHistoryScrollLoadsMore(t *testing.T) {
	dir, _ := historyRepo(t)

	for i := 0; i < 24; i++ {
		writeFile(t, dir, "app.go", fmt.Sprintf("package main\n\nfunc main() { println(%d) }\n", i))
		runGit(t, dir, "add", "-A")
		runGit(t, dir, "commit", "-q", "-m", fmt.Sprintf("commit %02d", i))
	}

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	log, err := loadCommitLogLimit(context.Background(), dir, "", maxCommits)
	if err != nil {
		t.Fatal(err)
	}

	if len(log.Commits) == 0 {
		t.Fatalf("direct log = %+v", log)
	}

	app.reloadChanges(context.Background(), false)
	app.setMode(context.Background(), "changes")

	if len(app.commits.Commits) != maxCommits {
		t.Fatalf("initial commits = %d, want %d", len(app.commits.Commits), maxCommits)
	}

	if !app.View().History.Scroll {
		t.Fatal("history does not expose a scroll rail")
	}

	app.scrollHistory(context.Background(), 100)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) && len(app.commits.Commits) <= maxCommits {
		if err := app.tickWatch(context.Background()); err != nil {
			t.Fatal(err)
		}

		time.Sleep(10 * time.Millisecond)
	}

	if len(app.commits.Commits) <= maxCommits {
		t.Fatalf("history did not load more: %d", len(app.commits.Commits))
	}

	if len(app.View().History.Commits[0].Graph) == 0 {
		t.Fatal("history row has no graph cells")
	}
}

func TestSampleHistoryGraph(t *testing.T) {
	app := newTestApp(t)
	clickID(t, app, "act-changes")

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if len(view.History.Commits) == 0 || len(view.History.Commits[0].Graph) == 0 {
		t.Fatalf("sample graph = %+v", view.History.Commits)
	}

	for _, cell := range view.History.Commits[0].Graph {
		t.Logf("graph cell: %+v", cell)
	}
}

func TestTabsStayInsideStrip(t *testing.T) {
	app := newTestApp(t)

	opened := 0

	for _, file := range app.ws.Files {
		if file.Path == "README.md" {
			continue
		}

		app.openFile(file.Path)
		opened++

		if opened == maxTabs {
			break
		}
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	editor, ok := boxByID(app.Page().Boxes(), "editor-window")
	if !ok {
		t.Fatal("no editor")
	}

	for i := 0; i < maxTabs; i++ {
		box, ok := boxByID(app.Page().Boxes(), fmt.Sprintf("tab-%d", i))
		if !ok {
			t.Fatalf("missing tab-%d", i)
		}

		if box.X < editor.X-1 || box.X+box.W > editor.X+editor.W+1 {
			t.Fatalf("tab-%d at %.0f..%.0f outside %.0f..%.0f", i, box.X, box.X+box.W, editor.X, editor.X+editor.W)
		}
	}
}
