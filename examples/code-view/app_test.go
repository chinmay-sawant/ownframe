package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

const shotDir = "/tmp/opencode/code-view-shots"

func newTestApp(t *testing.T) *App {
	t.Helper()

	app, err := New("")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatalf("redraw: %v", err)
	}

	return app
}

func boxByID(boxes []ownframe.Box, id string) (ownframe.Box, bool) {
	for _, box := range boxes {
		if box.ID == id {
			return box, true
		}
	}

	return ownframe.Box{}, false
}

func TestShellMetrics(t *testing.T) {
	app := newTestApp(t)

	if app.Page().Probe().Display() == nil {
		t.Fatal("page fell back to a bitmap")
	}

	if len(app.Page().Probe().PNG()) == 0 {
		t.Fatal("no page image")
	}

	cases := []struct {
		id         string
		x, y, w, h float64
	}{
		{"app-shell", 0, 0, defaultWidth, defaultHeight},
		{"titlebar", 0, 0, defaultWidth, chromeTop},
		{"sidebar", activityW, chromeTop, sideWidth, defaultHeight - chromeTop - chromeBottom},
		{"editor-window", chromeLeft, chromeTop, defaultWidth - chromeLeft, defaultHeight - chromeTop - chromeBottom},
		{"code-body", chromeLeft, editorBodyTop, defaultWidth - chromeLeft, defaultHeight - editorBodyTop - chromeBottom},
		{"statusbar", 0, defaultHeight - chromeBottom, defaultWidth, chromeBottom},
	}

	for _, tc := range cases {
		box, ok := boxByID(app.Page().Boxes(), tc.id)
		if !ok {
			t.Fatalf("missing box %s", tc.id)
		}

		if math.Abs(box.X-tc.x) > 1 || math.Abs(box.Y-tc.y) > 1 || math.Abs(box.W-tc.w) > 1 || math.Abs(box.H-tc.h) > 1 {
			t.Errorf("%s at (%.0f,%.0f) %.0fx%.0f, want (%.0f,%.0f) %.0fx%.0f",
				tc.id, box.X, box.Y, box.W, box.H, tc.x, tc.y, tc.w, tc.h)
		}
	}
}

func TestOpenAndCloseTabs(t *testing.T) {
	app := newTestApp(t)

	if app.View().Active != "README.md" {
		t.Fatalf("initial file = %q", app.View().Active)
	}

	app.openFile("backend/probe.go")

	if app.View().Active != "backend/probe.go" || len(app.View().Tabs) != 2 {
		t.Fatalf("after open: active=%q tabs=%d", app.View().Active, len(app.View().Tabs))
	}

	if len(app.View().Rows) == 0 || app.View().Rows[0].No != 1 {
		t.Fatalf("rows = %d", len(app.View().Rows))
	}

	app.closeTab(app.active)

	if app.View().Active != "README.md" || len(app.View().Tabs) != 1 {
		t.Fatalf("after close: active=%q tabs=%d", app.View().Active, len(app.View().Tabs))
	}

	app.closeTab(0)

	if !app.View().Welcome || app.View().Active != "" {
		t.Fatalf("welcome = %v active=%q", app.View().Welcome, app.View().Active)
	}

	app.openFile("web/index.html")

	if app.View().Active != "web/index.html" || len(app.View().Tabs) != 1 {
		t.Fatalf("reopen = %q tabs=%d", app.View().Active, len(app.View().Tabs))
	}
}

func TestModesAndDiff(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	if err := app.onClick(ctx, ownframe.Box{Action: "mode-changes"}); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if view.Mode != "changes" || !view.Diff.HasFiles || len(view.Diff.Files) != 7 {
		t.Fatalf("changes view = %+v", view.Diff)
	}

	if view.Diff.Selected != "backend/probe.go" {
		t.Fatalf("selected = %q", view.Diff.Selected)
	}

	app.setSplit(true)

	if !app.View().Diff.Split {
		t.Fatal("split did not turn on")
	}

	app.jumpHunk(1)

	if app.scrollY <= 0 {
		t.Fatalf("hunk jump did not scroll: %d", app.scrollY)
	}

	app.selectDiff("web/theme.css")

	if app.View().Diff.Selected != "web/theme.css" || app.View().Diff.Add == 0 {
		t.Fatalf("theme.css diff = %+v", app.View().Diff)
	}

	if err := app.onClick(ctx, ownframe.Box{Action: "mode-explorer"}); err != nil {
		t.Fatal(err)
	}

	if app.View().Mode != "explorer" || app.View().Diff.HasFiles {
		t.Fatalf("explorer view leaked the diff: %+v", app.View())
	}
}

func TestFilter(t *testing.T) {
	app := newTestApp(t)

	app.filter = "probe"
	app.refresh()

	view := app.View()
	if !view.Filtered || len(view.Tree) == 0 || view.Tree[0].Path != "backend/probe.go" {
		t.Fatalf("filtered tree = %+v", view.Tree)
	}

	if err := app.onClick(context.Background(), ownframe.Box{Action: "filter-clear"}); err != nil {
		t.Fatal(err)
	}

	if app.View().Filtered || app.filter != "" {
		t.Fatalf("filter not cleared: %q", app.filter)
	}
}

func TestFilterBinding(t *testing.T) {
	app := newTestApp(t)

	app.view.Filter = "theme"

	if err := app.onChange(context.Background(), ownframe.Box{ID: "filter"}); err != nil {
		t.Fatal(err)
	}

	if app.filter != "theme" || !app.View().Filtered {
		t.Fatalf("filter = %q filtered = %v", app.filter, app.View().Filtered)
	}

	if err := app.onKeyDown(context.Background(), "1"); err != nil {
		t.Fatal(err)
	}
}

func TestFilterShortcut(t *testing.T) {
	app := newTestApp(t)

	if err := app.onKeyDown(context.Background(), "2"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(context.Background(), "f"); err != nil {
		t.Fatal(err)
	}

	if app.View().Mode != "explorer" || app.Page().Form().FocusID() != "filter" {
		t.Fatalf("mode = %q focus = %q", app.View().Mode, app.Page().Form().FocusID())
	}
}

func TestScrollWindow(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")

	space := app.View().SpaceHeight
	if space <= defaultHeight {
		t.Fatalf("space = %d, want a scrollable page", space)
	}

	app.scrollTo(2000)
	app.Page().Scroll().SetScrollOffset(0, 2000)

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if view.TrackTop != 0 || view.Rows[0].No <= 1 || view.Rows[0].No > 200 {
		t.Fatalf("track = %d first = %d", view.TrackTop, view.Rows[0].No)
	}

	before := view.Rows[0].No
	if !app.onScroll(2200, defaultHeight) {
		t.Fatal("scroll callback reported no change")
	}

	if app.View().Rows[0].No <= before {
		t.Fatalf("first row did not advance: %d -> %d", before, app.View().Rows[0].No)
	}

	app.scrollTo(0)

	if app.View().TrackTop != 0 || app.View().Rows[0].No != 1 {
		t.Fatalf("reset = track %d first %d", app.View().TrackTop, app.View().Rows[0].No)
	}
}

func TestDetails(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")

	if len(app.View().Outline) == 0 {
		t.Fatal("no outline")
	}

	app.toggleDetails()

	if !app.View().Details {
		t.Fatal("details did not open")
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	editor, ok := boxByID(app.Page().Boxes(), "editor-window")
	if !ok || math.Abs(editor.W-(defaultWidth-chromeLeft-280)) > 1 {
		t.Fatalf("editor width = %.0f", editor.W)
	}

	app.toggleDetails()

	if app.View().Details {
		t.Fatal("details did not close")
	}
}

func TestDevTools(t *testing.T) {
	app := newTestApp(t)

	if !app.devOn || !app.View().DevOn {
		t.Fatal("devtools did not start on")
	}

	app.SetDevTools(false)

	if app.devOn || app.View().DevOn || app.Page().Window().DevTools() {
		t.Fatal("devtools did not turn off")
	}

	app.SetDevTools(true)

	if !app.Page().Window().DevTools() {
		t.Fatal("devtools did not turn back on")
	}
}

func TestWheelZoomStaysOff(t *testing.T) {
	app := newTestApp(t)

	if app.Page().Window().WheelZoomAllowed() {
		t.Fatal("control+wheel still zooms the window")
	}

	if app.Page().Window().KeyZoomAllowed() {
		t.Fatal("control plus and control minus still zoom the window")
	}
}

func TestCopyPathAndFile(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)

	app := newTestApp(t)
	app.openFile("backend/probe.go")
	app.copyActive(false)

	if got, err := ownframe.Read(); err != nil || got != "backend/probe.go" {
		t.Fatalf("clipboard path = %q (%v)", got, err)
	}

	app.copyActive(true)

	got, err := ownframe.Read()
	if err != nil || len(got) < 100 {
		t.Fatalf("clipboard file = %d bytes (%v)", len(got), err)
	}
}

func TestKeyboard(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	if err := app.onKeyDown(ctx, "2"); err != nil {
		t.Fatal(err)
	}

	if app.View().Mode != "changes" {
		t.Fatalf("mode = %q", app.View().Mode)
	}

	if err := app.onKeyDown(ctx, "1"); err != nil {
		t.Fatal(err)
	}

	if app.View().Mode != "explorer" {
		t.Fatalf("mode = %q", app.View().Mode)
	}

	app.openFile("backend/probe.go")

	if err := app.onKeyDown(ctx, "w"); err != nil {
		t.Fatal(err)
	}

	if len(app.View().Tabs) != 1 {
		t.Fatalf("tabs = %d", len(app.View().Tabs))
	}
}

func TestCleanChangeSet(t *testing.T) {
	app := newTestApp(t)
	app.changes = changeSet{Note: "clean tree"}
	app.statuses = statusMap(app.changes)
	app.mode = "changes"
	app.selectFirstDiff()
	app.refresh()

	view := app.View()
	if view.Diff.HasFiles || view.Diff.Note != "clean tree" || view.SpaceHeight < defaultHeight {
		t.Fatalf("clean view = %+v", view.Diff)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestToast(t *testing.T) {
	app := newTestApp(t)
	app.showToast("success", "Saved", "The change is local.")
	app.refresh()

	if !app.View().Toast.Show || app.View().Toast.Kind != "success" {
		t.Fatalf("toast = %+v", app.View().Toast)
	}

	app.clearToast()

	if app.View().Toast.Show {
		t.Fatal("toast did not clear")
	}
}

func TestLiveRoot(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "tool.go"), []byte("package main\n\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if app.ws.Sample || app.View().Workspace != filepath.Base(dir) || app.View().Active != "tool.go" {
		t.Fatalf("live root = sample %v workspace %q active %q", app.ws.Sample, app.View().Workspace, app.View().Active)
	}

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	// A plain temporary directory is not a repository. A quiet reload must
	// replace the change set with a no-repository state, not a startup toast.
	app.reloadChanges(context.Background(), false)

	if app.View().Toast.Show {
		t.Fatalf("quiet reload shows a toast: %+v", app.View().Toast)
	}

	if app.changes.EmptyTitle != "Git unavailable" || !strings.Contains(app.changes.Note, "Git") {
		t.Fatalf("no-repository state = %+v", app.changes)
	}
}

func TestResponsiveLayout(t *testing.T) {
	sizes := [][2]int{{1500, 900}, {1280, 720}, {1920, 1080}}

	for _, size := range sizes {
		app := newTestApp(t)
		app.Page().SetSize(size[0], size[1])

		if err := app.Page().Redraw(context.Background()); err != nil {
			t.Fatalf("%dx%d: %v", size[0], size[1], err)
		}

		status, ok := boxByID(app.Page().Boxes(), "statusbar")
		if !ok || math.Abs(status.Y+status.H-float64(size[1])) > 1 {
			t.Fatalf("%dx%d: status bar at %.0f", size[0], size[1], status.Y)
		}

		editor, ok := boxByID(app.Page().Boxes(), "editor-window")
		if !ok || editor.W <= 0 || editor.H <= 0 {
			t.Fatalf("%dx%d: editor %.0fx%.0f", size[0], size[1], editor.W, editor.H)
		}

		if editor.X+editor.W > float64(size[0])+1 {
			t.Fatalf("%dx%d: editor runs past the page", size[0], size[1])
		}
	}
}

func writeShot(t *testing.T, app *App, name string) {
	t.Helper()

	if err := os.MkdirAll(shotDir, 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(shotDir, name+".png"), app.Page().Probe().PNG(), 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestScreenshotMatrix captures the explorer and change view at the three
// review sizes. It is a development aid, not part of the default run.
func TestScreenshotMatrix(t *testing.T) {
	if os.Getenv("OWNFRAME_CODE_VIEW_SCREENSHOT") != "1" {
		t.Skip("set OWNFRAME_CODE_VIEW_SCREENSHOT=1 to capture the matrix")
	}

	for _, size := range [][2]int{{1280, 720}, {1500, 900}, {1920, 1080}} {
		app := newTestApp(t)
		app.Page().SetSize(size[0], size[1])

		if err := app.Page().Redraw(context.Background()); err != nil {
			t.Fatalf("%dx%d explorer: %v", size[0], size[1], err)
		}

		writeShot(t, app, fmt.Sprintf("matrix-%dx%d-explorer", size[0], size[1]))

		app.openFile("backend/probe.go")
		app.toggleDetails()

		if err := app.Page().Redraw(context.Background()); err != nil {
			t.Fatalf("%dx%d details: %v", size[0], size[1], err)
		}

		writeShot(t, app, fmt.Sprintf("matrix-%dx%d-details", size[0], size[1]))

		app.toggleDetails()
		app.scrollTo(800)
		app.Page().Scroll().SetScrollOffset(0, 800)

		if err := app.Page().Redraw(context.Background()); err != nil {
			t.Fatalf("%dx%d scrolled: %v", size[0], size[1], err)
		}

		writeShot(t, app, fmt.Sprintf("matrix-%dx%d-scrolled", size[0], size[1]))

		if err := app.onClick(context.Background(), ownframe.Box{Action: "mode-changes"}); err != nil {
			t.Fatalf("%dx%d changes: %v", size[0], size[1], err)
		}

		app.Page().Scroll().SetScrollOffset(0, 0)

		if err := app.Page().Redraw(context.Background()); err != nil {
			t.Fatalf("%dx%d changes redraw: %v", size[0], size[1], err)
		}

		writeShot(t, app, fmt.Sprintf("matrix-%dx%d-changes", size[0], size[1]))

		app.selectDiff("docs/deployment.md")
		app.setSplit(true)
		app.Page().Scroll().SetScrollOffset(0, 0)

		if err := app.Page().Redraw(context.Background()); err != nil {
			t.Fatalf("%dx%d split: %v", size[0], size[1], err)
		}

		writeShot(t, app, fmt.Sprintf("matrix-%dx%d-split", size[0], size[1]))
	}
}

// TestScreenshotScrollRails renders a long change list and history rail.
func TestScreenshotScrollRails(t *testing.T) {
	if os.Getenv("OWNFRAME_CODE_VIEW_SCREENSHOT") != "1" {
		t.Skip("set OWNFRAME_CODE_VIEW_SCREENSHOT=1 to capture the page")
	}

	app := newTestApp(t)
	app.mode = "changes"
	app.changes = changeSet{Branch: "feature/rails", Note: "Synthetic rail fixture."}

	for i := 0; i < 30; i++ {
		app.changes.Files = append(app.changes.Files, changedFile{
			Path: fmt.Sprintf("pkg/file-%02d.go", i), Status: "M", Add: i, Del: i / 2,
		})
	}

	app.commits.Commits = nil
	app.commits.Branch = "feature/rails"

	for i := 0; i < 30; i++ {
		hash := fmt.Sprintf("%040x", i+1)
		app.commits.Commits = append(app.commits.Commits, gitCommit{
			Hash: hash, Short: hash[:7], Subject: fmt.Sprintf("Commit %02d", i),
			Author: "Mira Chen", Date: "2026-10-24",
			Parents: func() []string {
				if i+1 < 30 {
					return []string{fmt.Sprintf("%040x", i+2)}
				}

				return nil
			}(),
		})
	}

	app.historyLabel = "feature/rails"
	app.headBranch = "feature/rails"
	app.statuses = statusMap(app.changes)
	app.selectFirstDiff()
	app.historyScroll.SetOffset(0, len(app.commits.Commits), app.historyVisibleRows())
	app.changesScroll.SetOffset(0, len(app.changes.Files), app.changeVisibleRows())
	app.refresh()

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	writeShot(t, app, "scrollrails")
}

// TestScreenshotIgnored renders a workspace with git-ignored files and
// folders.
func TestScreenshotIgnored(t *testing.T) {
	if os.Getenv("OWNFRAME_CODE_VIEW_SCREENSHOT") != "1" {
		t.Skip("set OWNFRAME_CODE_VIEW_SCREENSHOT=1 to capture the page")
	}

	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	writeFile(t, dir, ".gitignore", "build/\nnode_modules/\n*.test\nsecret.txt\n")
	writeFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")
	writeFile(t, dir, "secret.txt", "secret\n")
	writeFile(t, dir, "motion.test", "ignored\n")
	writeFile(t, dir, "build/out.txt", "build\n")
	writeFile(t, dir, "node_modules/pkg/index.js", "module\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.folders["build"] = true
	app.refresh()

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	writeShot(t, app, "ignored")
}

func TestScreenshot(t *testing.T) {
	if os.Getenv("OWNFRAME_CODE_VIEW_SCREENSHOT") != "1" {
		t.Skip("set OWNFRAME_CODE_VIEW_SCREENSHOT=1 to capture the page")
	}

	app := newTestApp(t)
	writeShot(t, app, "explorer")

	app.togglePreview()

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	writeShot(t, app, "preview")
	app.togglePreview()

	if err := app.onClick(context.Background(), ownframe.Box{Action: "palette-open"}); err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	writeShot(t, app, "palette")

	if err := app.onClick(context.Background(), ownframe.Box{Action: "palette-close"}); err != nil {
		t.Fatal(err)
	}

	app.openFile("backend/probe.go")
	app.toggleDetails()

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	writeShot(t, app, "details")

	app.toggleDetails()

	app.openFile("README.md")
	app.toggleEdit()

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	writeShot(t, app, "edit")
	app.finishEdit()
	app.openFile("backend/probe.go")

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	for _, row := range app.View().Rows {
		if row.Mark != "" {
			app.showInlineCard(row.ChangeLine)
			break
		}
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	writeShot(t, app, "inline-card")
	app.closeInlineCard()

	app.toggleSplit()
	app.openFileIn(1, "web/theme.css")
	app.resetScroll()

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	writeShot(t, app, "split-editor")
	app.toggleSplit()

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

		if !app.View().Search.Busy {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	writeShot(t, app, "search")
	_ = app.closeWorkspaceSearch(context.Background())

	app.scrollTo(1600)
	app.Page().Scroll().SetScrollOffset(0, 1600)

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	writeShot(t, app, "scrolled")

	if err := app.onClick(context.Background(), ownframe.Box{Action: "mode-changes"}); err != nil {
		t.Fatal(err)
	}

	app.Page().Scroll().SetScrollOffset(0, 0)

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	writeShot(t, app, "changes")

	if len(app.View().History.Commits) > 0 {
		short := app.View().History.Commits[0].Short
		if box, ok := boxByID(app.Page().Boxes(), "commit-"+short); ok {
			if err := app.Page().Edit().Hover(context.Background(), box.X+box.W/2, box.Y+box.H/2); err != nil {
				t.Fatal(err)
			}

			if err := app.Page().Redraw(context.Background()); err != nil {
				t.Fatal(err)
			}

			writeShot(t, app, "commit-tip")
			_ = app.Page().Edit().Hover(context.Background(), 1, 1)
		}
	}

	app.selectDiff("docs/deployment.md")
	app.setSplit(true)
	app.Page().Scroll().SetScrollOffset(0, 0)

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	writeShot(t, app, "split")

	app.mode = "explorer"

	for len(app.tabs) > 0 {
		app.closeTab(0)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	writeShot(t, app, "welcome")
}

func TestAppName(t *testing.T) {
	html := buildHTML()

	if !strings.Contains(html, "Code Review") {
		t.Fatal("the shell does not use the Code Review name")
	}

	if strings.Contains(html, "Beacon CODE VIEW") {
		t.Fatal("the old Beacon brand is still in the shell")
	}
}

func TestRealWorkspaceAtStartup(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	app, err := New(".")
	if err != nil {
		t.Fatal(err)
	}

	if app.ws.Sample {
		t.Fatal("the current directory was scanned as the sample")
	}

	if app.View().Workspace != "code-view" || app.View().Active != "README.md" {
		t.Fatalf("workspace = %q active = %q", app.View().Workspace, app.View().Active)
	}

	app.reloadChanges(context.Background(), false)

	if app.changes.Sample || app.changes.Branch == "" || app.changes.Branch == "no git" {
		t.Fatalf("live change set = %+v", app.changes)
	}
}
