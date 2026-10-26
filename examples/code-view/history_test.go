package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.CommandContext(context.Background(), "git", append([]string{"-C", dir}, args...)...) //nolint:gosec // test helper runs the fixed git binary on a temp repo
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Code View Test",
		"GIT_AUTHOR_EMAIL=code-view@example.com",
		"GIT_COMMITTER_NAME=Code View Test",
		"GIT_COMMITTER_EMAIL=code-view@example.com",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}

	return strings.TrimSpace(string(out))
}

func writeFile(t *testing.T, dir, name, text string) {
	t.Helper()

	path := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func historyRepo(t *testing.T) (string, []string) {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.name", "Code View Test")
	runGit(t, dir, "config", "user.email", "code-view@example.com")

	writeFile(t, dir, "README.md", "# Base\n\nA fixture repository.\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "base")

	writeFile(t, dir, "app.go", "package main\n\nfunc main() {}\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "add app")

	writeFile(t, dir, "app.go", "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "print hello", "-m", "owner: code-view team")

	writeFile(t, dir, "README.md", "# Base\n\nA fixture repository.\n\nworking change\n")

	hashes := strings.Fields(runGit(t, dir, "log", "-3", "--format=%H"))

	return dir, hashes
}

func newHistoryApp(t *testing.T) (*App, []string) {
	t.Helper()

	dir, hashes := historyRepo(t)
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.reloadChanges(context.Background(), false)

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := app.onClick(context.Background(), ownframe.Box{Action: "mode-changes"}); err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return app, hashes
}

func TestHistoryPaneAndCommitExpand(t *testing.T) {
	app, hashes := newHistoryApp(t)
	view := app.View()

	if !view.History.Visible || len(view.History.Commits) != 3 {
		t.Fatalf("history = %+v", view.History)
	}

	if view.History.Branch != "main" {
		t.Fatalf("history branch = %q", view.History.Branch)
	}

	if view.History.Commits[0].Subject != "print hello" || view.History.Commits[0].Hash != hashes[0] {
		t.Fatalf("newest = %+v", view.History.Commits[0])
	}

	if !view.Diff.HasChanges || !view.Diff.HasFiles {
		t.Fatalf("working diff = %+v", view.Diff)
	}

	short := view.History.Commits[0].Short
	clickID(t, app, "commit-"+short)

	view = app.View()
	if !view.History.Commits[0].Expanded || view.Diff.Source != "commit" || !view.Diff.HasFiles {
		t.Fatalf("expanded = %+v diff = %+v", view.History.Commits[0], view.Diff)
	}

	text := ""

	for _, row := range view.Diff.Rows {
		text += lineText(row.Left) + lineText(row.Right)
	}

	if !strings.Contains(text, "println") {
		t.Fatalf("commit diff rows do not contain the change:\n%s", text)
	}

	app.setSplit(true)

	for _, row := range app.View().History.Commits[0].Preview {
		if row.Kind == "split" {
			t.Fatal("history preview used split rows")
		}
	}

	app.setSplit(false)

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	fileBox, ok := boxByID(app.Page().Boxes(), "commit-file-"+hashes[0]+"-app.go")
	if !ok {
		t.Fatal("no commit file row")
	}

	if err := app.Page().Click(context.Background(), fileBox.X+fileBox.W/2, fileBox.Y+fileBox.H/2); err != nil {
		t.Fatal(err)
	}

	if app.View().Diff.Path != "app.go" {
		t.Fatalf("selected commit file = %q", app.View().Diff.Path)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	for _, box := range app.Page().Boxes() {
		if box.Action != "" && box.ID == "" {
			t.Fatalf("history action %q has no id", box.Action)
		}
	}

	clickID(t, app, "history-working")

	if app.View().Diff.Source != "working" {
		t.Fatalf("working diff source = %q", app.View().Diff.Source)
	}

	clickID(t, app, "commit-"+short)

	if app.View().Diff.Source != "commit" {
		t.Fatalf("commit diff source after return = %q", app.View().Diff.Source)
	}
}

func TestHistoryRefreshAddsCommit(t *testing.T) {
	app, _ := newHistoryApp(t)
	dir := app.ws.Root

	writeFile(t, dir, "app.go", "package main\n\nfunc main() {\n\tprintln(\"again\")\n}\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "fourth commit")

	clickID(t, app, "history-refresh")

	view := app.View()
	if len(view.History.Commits) != 4 || view.History.Commits[0].Subject != "fourth commit" {
		t.Fatalf("history after refresh = %+v", view.History.Commits)
	}
}

func TestHistoryHoverAndSplitDrag(t *testing.T) {
	app, hashes := newHistoryApp(t)
	view := app.View()
	short := view.History.Commits[0].Short

	box, ok := boxByID(app.Page().Boxes(), "commit-"+short)
	if !ok {
		t.Fatal("no commit row")
	}

	if err := app.Page().Edit().Hover(context.Background(), box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatal(err)
	}

	if !app.View().History.Tip.Show || !strings.Contains(app.View().History.Tip.Meta, short) {
		t.Fatalf("tip = %+v", app.View().History.Tip)
	}

	divider, ok := boxByID(app.Page().Boxes(), "history-divider")
	if !ok {
		t.Fatal("no divider")
	}

	before := app.View().History.Height
	claimed, err := app.Page().Testing().BeginDrag(context.Background(), divider.X+divider.W/2, divider.Y+divider.H/2)
	if err != nil || !claimed {
		t.Fatalf("drag claim = %v %v", claimed, err)
	}

	if err := app.Page().Testing().MoveDrag(context.Background(), divider.X+divider.W/2, divider.Y+divider.H/2+90); err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Testing().EndDrag(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if app.View().History.Height <= before {
		t.Fatalf("history height = %d, before %d", app.View().History.Height, before)
	}

	_ = hashes
}

func TestWorkingTreeTabClose(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")
	clickID(t, app, "act-changes")
	clickID(t, app, "tab-close-working")

	view := app.View()
	if view.Mode != "explorer" || view.Active != "backend/probe.go" || len(view.Tabs) != 2 {
		t.Fatalf("mode=%q active=%q tabs=%d", view.Mode, view.Active, len(view.Tabs))
	}

	clickID(t, app, "act-changes")

	if err := app.onKeyDown(context.Background(), "w"); err != nil {
		t.Fatal(err)
	}

	if app.View().Mode != "explorer" {
		t.Fatalf("W left mode %q", app.View().Mode)
	}

	app.closeTab(0)
	app.closeTab(0)

	if !app.View().Welcome {
		t.Fatal("expected the welcome panel")
	}

	clickID(t, app, "act-changes")
	clickID(t, app, "tab-close-working")

	view = app.View()
	if view.Mode != "explorer" || view.Active != "README.md" || len(view.Tabs) != 1 {
		t.Fatalf("no-tab close: mode=%q active=%q tabs=%d", view.Mode, view.Active, len(view.Tabs))
	}
}

func TestDiffLinesClip(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.name", "Code View Test")
	runGit(t, dir, "config", "user.email", "code-view@example.com")

	writeFile(t, dir, "long.go", "package main\n\nvar x = \""+strings.Repeat("a", 300)+"\"\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "long line")
	writeFile(t, dir, "long.go", "package main\n\nvar x = \""+strings.Repeat("b", 300)+"\"\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.reloadChanges(context.Background(), false)
	if err := app.onClick(context.Background(), ownframe.Box{Action: "mode-changes"}); err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	unified, split := app.diffColumns()
	max := unified

	for _, row := range app.View().Diff.Rows {
		if text := lineText(row.Left); len([]rune(text)) > max+1 {
			t.Fatalf("unified row is %d runes, max %d", len([]rune(text)), max)
		}
	}

	app.setSplit(true)

	for _, row := range app.View().Diff.Rows {
		if row.Kind != "split" {
			continue
		}

		for _, text := range []string{lineText(row.Left), lineText(row.Right)} {
			if len([]rune(text)) > split+1 {
				t.Fatalf("split row is %d runes, max %d", len([]rune(text)), split)
			}
		}
	}
}

func TestWrapLongCodeLine(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Fixture\n")
	writeFile(t, dir, "AGENTS.md", strings.Repeat("x", 240)+"\nsecond line\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.openFile("AGENTS.md")

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if !view.Wrap {
		t.Fatal("wrap is off by default")
	}

	continuations := 0
	firstLine := ""

	for _, row := range view.Rows {
		if row.Line == 0 {
			firstLine += lineText(row.Segments)
		}

		if row.No == 0 {
			continuations++
		}
	}

	if continuations == 0 {
		t.Fatalf("no wrapped rows: %+v", view.Rows)
	}

	if firstLine != strings.Repeat("x", 240) {
		t.Fatalf("wrapped text changed: %q", firstLine)
	}

	if view.SpaceHeight <= editorBodyTop+2*rowH+chromeBottom {
		t.Fatalf("space height = %d", view.SpaceHeight)
	}

	clickID(t, app, "wrap-toggle")

	if app.View().Wrap {
		t.Fatal("wrap did not toggle off")
	}
}

func TestBodyDoesNotClipTheScrollSpacer(t *testing.T) {
	// The engine maps overflow-x:hidden to a two-axis clip. On a body that is
	// only as tall as the viewport that clips the scroll-space spacer and makes
	// every wheel or ScrollTo clamp to zero.
	if strings.Contains(buildHTML(), "overflow-x: hidden") {
		t.Fatal("body overflow-x hidden clips the page scroll")
	}
}

func TestWrapReflowsOnResize(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Fixture\n")
	writeFile(t, dir, "AGENTS.md", strings.Repeat("word ", 100)+"\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.openFile("AGENTS.md")
	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	doc := app.docs["AGENTS.md"]
	wideRows := len(doc.Wrap.Rows)
	wideCols := doc.Wrap.Cols

	app.Page().SetSize(840, 700)
	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	narrowRows := len(doc.Wrap.Rows)
	if doc.Wrap.Cols >= wideCols || narrowRows <= wideRows {
		t.Fatalf("narrow cols=%d rows=%d, wide cols=%d rows=%d", doc.Wrap.Cols, narrowRows, wideCols, wideRows)
	}

	app.Page().SetSize(1500, 900)
	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if len(doc.Wrap.Rows) != wideRows {
		t.Fatalf("wide rows after restore = %d, want %d", len(doc.Wrap.Rows), wideRows)
	}
}

func TestPaletteOpenClearsFocus(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	if err := app.openPalette(ctx); err != nil {
		t.Fatal(err)
	}

	app.view.PaletteQuery = "probe"

	if err := app.onChange(ctx, ownframe.Box{ID: "palette"}); err != nil {
		t.Fatal(err)
	}

	if err := app.onSubmit(ctx); err != nil {
		t.Fatal(err)
	}

	if app.View().Active != "backend/probe.go" {
		t.Fatalf("active = %q", app.View().Active)
	}

	if focus := app.Page().Form().FocusID(); focus != "" {
		t.Fatalf("focus stayed on %q", focus)
	}

	if err := app.onKeyDown(ctx, "pagedown"); err != nil {
		t.Fatal(err)
	}

	if app.scrollY == 0 {
		t.Fatal("keyboard scroll did not run after palette open")
	}
}

func TestScrollAfterOpeningLongFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Fixture\n")
	writeFile(t, dir, "AGENTS.md", strings.Repeat("line\n", 200))

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.openFile("AGENTS.md")

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	space := float64(app.View().SpaceHeight)
	found := false

	for _, box := range app.Page().Boxes() {
		if box.W <= 1.5 && box.H >= space-1 {
			found = true

			break
		}
	}

	if !found {
		t.Fatalf("no scroll spacer box with height %.0f", space)
	}

	if !app.onScroll(200, defaultHeight) {
		t.Fatal("scroll callback reported no change")
	}

	if app.scrollY != 200 || len(app.View().Rows) == 0 || app.View().Rows[0].No <= 1 {
		t.Fatalf("scrollY=%d first=%d", app.scrollY, app.View().Rows[0].No)
	}
}
