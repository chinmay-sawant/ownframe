package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

// waitForCommitPatch drains the asynchronous patch worker in tests.
func waitForCommitPatch(t *testing.T, app *App) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		if err := app.tickWatch(context.Background()); err != nil {
			t.Fatal(err)
		}

		if app.View().Diff.HasFiles && app.View().Diff.Source == "commit" {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("commit patch did not arrive: %+v", app.View().Diff)
}

func TestSampleHistoryRail(t *testing.T) {
	app := newTestApp(t)
	clickID(t, app, "act-changes")

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if !view.History.Visible || !view.History.Sample || len(view.History.Commits) != 3 {
		t.Fatalf("sample history = %+v", view.History)
	}

	if view.History.Branch != "feat/probe-retries" {
		t.Fatalf("sample branch = %q", view.History.Branch)
	}

	add, del := 0, 0
	for _, commit := range view.History.Commits {
		add += commit.Add
		del += commit.Del
	}

	if add != 562 || del != 109 {
		t.Fatalf("sample totals = +%d -%d", add, del)
	}

	short := view.History.Commits[0].Short
	clickID(t, app, "commit-"+short)

	view = app.View()
	if view.Diff.Source != "commit" || !view.Diff.HasFiles || view.Diff.CommitShort != short {
		t.Fatalf("sample commit diff = %+v", view.Diff)
	}

	if len(view.Diff.Rows) == 0 {
		t.Fatal("sample commit has no diff rows")
	}

	clickID(t, app, "history-working")

	if app.View().Diff.Source != "working" {
		t.Fatalf("sample working source = %q", app.View().Diff.Source)
	}
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
	waitForCommitPatch(t, app)

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

	if err := app.Page().Testing().MoveDrag(context.Background(), divider.X+divider.W/2, divider.Y+divider.H/2-90); err != nil {
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

func TestTreeExpandedFolderScrolls(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Fixture\n")

	for i := 0; i < 60; i++ {
		writeFile(t, dir, fmt.Sprintf("many/file%02d.go", i), "package many\n")
	}

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.Page().SetSize(1280, 720)

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	clickID(t, app, "tree-many")

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if !view.TreeScroll || view.TreeTotal <= view.TreeVisible || len(view.Tree) >= view.TreeTotal {
		t.Fatalf("tree total=%d visible=%d rows=%d scroll=%v", view.TreeTotal, view.TreeVisible, len(view.Tree), view.TreeScroll)
	}

	treeBox, ok := boxByID(app.Page().Boxes(), "tree-pane")
	if !ok {
		t.Fatal("no tree pane")
	}

	handled, err := app.onWheel(context.Background(), treeBox.X+treeBox.W/2, treeBox.Y+treeBox.H/2, 0, -3)
	if err != nil || !handled {
		t.Fatalf("tree wheel = %v %v", handled, err)
	}

	if app.treeOffset == 0 {
		t.Fatal("tree wheel did not scroll")
	}

	app.treeOffset = 0
	app.refresh()

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	first := view.Tree[0].Path
	thumb, ok := boxByID(app.Page().Boxes(), "tree-thumb")
	if !ok {
		t.Fatal("no tree thumb")
	}

	if _, ok := boxByID(app.Page().Boxes(), "tree-up"); !ok {
		t.Fatal("no tree up button")
	}

	if _, ok := boxByID(app.Page().Boxes(), "tree-down"); !ok {
		t.Fatal("no tree down button")
	}

	claimed, err := app.Page().Testing().BeginDrag(context.Background(), thumb.X+thumb.W/2, thumb.Y+thumb.H/2)
	if err != nil || !claimed {
		t.Fatalf("tree drag claim = %v %v", claimed, err)
	}

	if err := app.Page().Testing().MoveDrag(context.Background(), thumb.X+thumb.W/2, thumb.Y+thumb.H/2+120); err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Testing().EndDrag(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if app.treeOffset == 0 || app.View().Tree[0].Path == first {
		t.Fatalf("tree offset=%d first=%q", app.treeOffset, app.View().Tree[0].Path)
	}
}

func TestBranchPickerLoadsHistoryWithoutCheckout(t *testing.T) {
	app, _ := newHistoryApp(t)
	dir := app.ws.Root

	runGit(t, dir, "checkout", "-q", "-b", "feature")
	writeFile(t, dir, "feature.go", "package main\n\nconst feature = true\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "feature work")
	runGit(t, dir, "checkout", "-q", "main")

	app.reloadChanges(context.Background(), false)
	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	clickID(t, app, "branch-picker")

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if !view.History.BranchOpen {
		t.Fatal("branch picker did not open")
	}

	index := -1

	for i, branch := range view.History.Branches {
		if branch.Name == "feature" {
			index = i
		}
	}

	if index < 0 {
		t.Fatalf("feature branch missing from %+v", view.History.Branches)
	}

	clickID(t, app, fmt.Sprintf("branch-%d", index))

	view = app.View()
	if view.History.Branch != "feature" || len(view.History.Commits) != 4 {
		t.Fatalf("feature history = %+v", view.History)
	}

	if view.Branch != "main" {
		t.Fatalf("status branch = %q, want main", view.Branch)
	}

	if view.Diff.Source != "working" {
		t.Fatalf("diff source = %q", view.Diff.Source)
	}

	if head := runGit(t, dir, "rev-parse", "--abbrev-ref", "HEAD"); head != "main" {
		t.Fatalf("git HEAD moved to %q", head)
	}
}

func TestFindInFile(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")
	ctx := context.Background()

	if err := app.openFind(ctx); err != nil {
		t.Fatal(err)
	}

	if !app.View().FindOpen || app.Page().Form().FocusID() != "find" {
		t.Fatalf("find open=%v focus=%q", app.View().FindOpen, app.Page().Form().FocusID())
	}

	app.view.FindQuery = "func"
	if err := app.onChange(ctx, ownframe.Box{ID: "find"}); err != nil {
		t.Fatal(err)
	}

	doc := app.docs["backend/probe.go"]
	want := strings.Count(strings.ToLower(strings.Join(doc.Lines, "\n")), "func")

	if app.View().FindTotal != want || want == 0 {
		t.Fatalf("matches = %d, want %d", app.View().FindTotal, want)
	}

	app.findStep(1)

	if app.View().FindCurrent != 2 {
		t.Fatalf("current = %d", app.View().FindCurrent)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	match := false

	for _, row := range app.View().Rows {
		for _, seg := range row.Segments {
			if strings.Contains(seg.Class, "find-hit") {
				match = true
			}
		}
	}

	if !match {
		t.Fatal("visible match was not highlighted")
	}

	if err := app.closeFind(ctx); err != nil {
		t.Fatal(err)
	}

	if app.View().FindOpen || len(app.findHits) != 0 {
		t.Fatal("find did not close")
	}
}

func TestKeyboardActivatesFocusedControl(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	if err := app.Page().Form().Focus(ctx, "act-changes"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "enter"); err != nil {
		t.Fatal(err)
	}

	if app.View().Mode != "changes" {
		t.Fatalf("mode = %q", app.View().Mode)
	}

	if app.Page().Form().FocusID() != "act-changes" {
		t.Fatalf("focus = %q", app.Page().Form().FocusID())
	}

	clickID(t, app, "act-explorer")

	if err := app.Page().Form().Focus(ctx, "tree-backend/main.go"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "space"); err != nil {
		t.Fatal(err)
	}

	if app.View().Active != "backend/main.go" {
		t.Fatalf("active = %q", app.View().Active)
	}
}

func TestTreeKeyboardNavigation(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	if err := app.Page().Form().Focus(ctx, "tree-backend"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "arrowdown"); err != nil {
		t.Fatal(err)
	}

	if focus := app.Page().Form().FocusID(); focus != "tree-backend/main.go" {
		t.Fatalf("focus after down = %q", focus)
	}

	if err := app.onKeyDown(ctx, "arrowup"); err != nil {
		t.Fatal(err)
	}

	if focus := app.Page().Form().FocusID(); focus != "tree-backend" {
		t.Fatalf("focus after up = %q", focus)
	}
}

func TestHistoryLivesInSidebar(t *testing.T) {
	app := newTestApp(t)
	clickID(t, app, "act-changes")

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	sidebar, ok := boxByID(app.Page().Boxes(), "sidebar")
	if !ok {
		t.Fatal("no sidebar")
	}

	pane, ok := boxByID(app.Page().Boxes(), "history-pane")
	if !ok {
		t.Fatal("history pane is not in the sidebar")
	}

	divider, ok := boxByID(app.Page().Boxes(), "history-divider")
	if !ok {
		t.Fatal("no history divider")
	}

	code, ok := boxByID(app.Page().Boxes(), "code-body")
	if !ok {
		t.Fatal("no code body")
	}

	if pane.X < sidebar.X || pane.X+pane.W > sidebar.X+sidebar.W+1 {
		t.Fatalf("history outside sidebar: pane %.0f..%.0f sidebar %.0f..%.0f", pane.X, pane.X+pane.W, sidebar.X, sidebar.X+sidebar.W)
	}

	if divider.X < sidebar.X || divider.X+divider.W > sidebar.X+sidebar.W+1 {
		t.Fatalf("divider outside sidebar: %.0f..%.0f", divider.X, divider.X+divider.W)
	}

	if want := float64(defaultHeight - editorBodyTop - chromeBottom); code.H != want {
		t.Fatalf("code body height = %.0f, want %.0f", code.H, want)
	}

	if int(pane.H) != app.View().History.Height {
		t.Fatalf("pane height = %.0f, view = %d", pane.H, app.View().History.Height)
	}
}

func TestTreeWheelAfterPageScroll(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Fixture\n")
	for i := 0; i < 60; i++ {
		writeFile(t, dir, fmt.Sprintf("many/file%02d.go", i), "package many\n")
	}

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.Page().SetSize(1280, 720)
	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	clickID(t, app, "tree-many")
	app.scrollTo(400)
	app.Page().Scroll().SetScrollOffset(0, 400)

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	tree, ok := boxByID(app.Page().Boxes(), "tree-pane")
	if !ok {
		t.Fatal("no tree")
	}

	before := app.treeOffset
	handled, err := app.onWheel(context.Background(), tree.X+20, tree.Y+20+400, 0, -3)
	if err != nil || !handled {
		t.Fatalf("wheel = %v %v", handled, err)
	}

	if app.treeOffset <= before {
		t.Fatalf("tree offset = %d, before %d", app.treeOffset, before)
	}

	if _, scrollY := app.Page().Scroll().ScrollOffset(); scrollY != 400 {
		t.Fatalf("page scroll moved to %d", scrollY)
	}
}

func TestMouseDragHistoryDivider(t *testing.T) {
	app, _ := newHistoryApp(t)
	divider, ok := boxByID(app.Page().Boxes(), "history-divider")
	if !ok {
		t.Fatal("no divider")
	}

	before := app.View().History.Height
	ctx := context.Background()

	claimed, err := app.Page().Testing().BeginDrag(ctx, divider.X+divider.W/2, divider.Y+divider.H/2)
	if err != nil || !claimed {
		t.Fatalf("begin drag = %v %v", claimed, err)
	}

	if err := app.Page().Testing().MoveDrag(ctx, divider.X+divider.W/2, divider.Y+divider.H/2-80); err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Testing().EndDrag(ctx); err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if app.View().History.Height <= before {
		t.Fatalf("history height = %d, before %d", app.View().History.Height, before)
	}
}

func TestCollapseAllIcon(t *testing.T) {
	app := newTestApp(t)

	if len(app.folders) == 0 {
		t.Fatal("sample should start with open folders")
	}

	clickID(t, app, "side-collapse")

	if len(app.folders) != 0 {
		t.Fatalf("folders = %v", app.folders)
	}
}
