package main

import (
	"bytes"
	"context"
	"html/template"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestLineRevertButtonsInBothViews(t *testing.T) {
	dir, _ := historyRepo(t)
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.reloadChanges(ctx, false)
	app.mode = "changes"
	app.refresh()

	if len(app.changes.Files) == 0 {
		t.Fatal("live repo has no changes")
	}

	// Unified view offers no revert arrow on hunk headers; the discard
	// control lives on the changed lines.
	app.setSplit(false)
	view := app.View()
	if !view.Diff.CanRevert {
		t.Fatal("unified diff does not offer hunk revert")
	}

	for _, row := range view.Diff.Rows {
		if row.Kind == rowHunk && row.HasRevert {
			t.Fatal("unified hunk rows carry a revert marker")
		}
	}

	lineFound := false
	for _, row := range view.Diff.Rows {
		if (row.Kind == rowAdd || row.Kind == rowDel) && row.HasRevert && !row.Cont {
			lineFound = true
		}
	}
	if !lineFound {
		t.Fatal("unified changed rows carry no per-line revert marker")
	}

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if _, ok := boxByID(app.Page().Boxes(), "revert-hunk-0"); ok {
		t.Fatal("unified view still shows a hunk-header revert button")
	}

	perLine := false
	for _, box := range app.Page().Boxes() {
		if strings.HasPrefix(box.ID, "revert-line-") {
			perLine = true
		}
	}
	if !perLine {
		t.Fatal("unified view misses per-line revert buttons")
	}

	// Split view offers the same per-line arrows on its changed rows, and
	// no arrow on the hunk headers.
	app.setSplit(true)
	view = app.View()
	if !view.Diff.CanRevert {
		t.Fatal("split diff does not offer hunk revert")
	}

	for _, row := range view.Diff.Rows {
		if row.Kind == rowHunk && row.HasRevert {
			t.Fatal("split hunk rows carry a revert marker")
		}
	}

	splitLine := false
	for _, row := range view.Diff.Rows {
		if row.Kind == "split" && row.HasRevert {
			splitLine = true
		}
	}
	if !splitLine {
		t.Fatal("split changed rows carry no per-line revert marker")
	}

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if _, ok := boxByID(app.Page().Boxes(), "revert-hunk-0"); ok {
		t.Fatal("split view still shows a hunk-header revert button")
	}

	perLine = false
	for _, box := range app.Page().Boxes() {
		if strings.HasPrefix(box.ID, "revert-line-") {
			perLine = true
		}
	}
	if !perLine {
		t.Fatal("split view misses per-line revert buttons")
	}
}

func TestRevertHunkUnstagedLive(t *testing.T) {
	dir, _ := historyRepo(t)
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.reloadChanges(ctx, false)
	app.mode = "changes"
	app.refresh()

	if len(app.changes.Files) == 0 {
		t.Fatal("live repo has no changes")
	}

	path := app.changes.Files[0].Path
	app.selectDiff(path)

	if err := app.revertHunk(ctx, 0); err != nil {
		t.Fatal(err)
	}

	full := filepath.Join(dir, filepath.FromSlash(path))
	after, err := os.ReadFile(full) //nolint:gosec // path is a t.TempDir fixture
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(after), "working change") {
		t.Fatalf("hunk revert kept the edit: %q", after)
	}

	if findFile(app.changes.Files, path) != nil {
		t.Fatalf("reverted hunk still listed: %v", app.changes.Files)
	}
}

func TestRevertHunkStagedLive(t *testing.T) {
	dir, _ := historyRepo(t)
	runGit(t, dir, "add", "-A")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.reloadChanges(ctx, false)

	if len(app.changes.Staged) == 0 {
		t.Fatal("expected staged changes")
	}

	path := app.changes.Staged[0].Path
	app.selectDiff(path)

	if err := app.revertHunk(ctx, 0); err != nil {
		t.Fatal(err)
	}

	if findFile(app.changes.Files, path) != nil {
		t.Fatalf("staged hunk still listed: %v", app.changes.Files)
	}

	if len(app.changes.Staged) != 0 {
		t.Fatalf("staged hunks remain: %v", app.changes.Staged)
	}
}

// firstRevertLineBox returns the first per-line revert button in layout
// order. The hunk headers carry no button anymore; the per-line control on
// the changed row reverts the whole hunk.
func firstRevertLineBox(app *App) (ownframe.Box, bool) {
	for _, box := range app.Page().Boxes() {
		if strings.HasPrefix(box.ID, "revert-line-") {
			return box, true
		}
	}

	return ownframe.Box{}, false
}

func TestRevertHunkViaClick(t *testing.T) {
	dir, _ := historyRepo(t)
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.reloadChanges(ctx, false)
	app.mode = "changes"
	app.refresh()
	app.selectDiff(app.changes.Files[0].Path)

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	box, ok := firstRevertLineBox(app)
	if !ok {
		t.Fatal("missing per-line revert button")
	}

	if err := app.onClick(ctx, ownframe.Box{ID: box.ID, Action: box.Action}); err != nil {
		t.Fatal(err)
	}

	if len(app.changes.Files) != 0 {
		t.Fatalf("click revert left %d files", len(app.changes.Files))
	}
}

func TestRevertHunkKeepsTabs(t *testing.T) {
	dir, _ := historyRepo(t)
	writeFile(t, dir, "tabs.go", "package main\n\nfunc main() {\n\tprintln(1)\n}\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "tabs")
	writeFile(t, dir, "tabs.go", "package main\n\nfunc main() {\n\tprintln(2)\n}\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.reloadChanges(ctx, false)
	app.selectDiff("tabs.go")

	if err := app.revertHunk(ctx, 0); err != nil {
		t.Fatal(err)
	}

	after, err := os.ReadFile(filepath.Join(dir, "tabs.go")) //nolint:gosec // path is a t.TempDir fixture
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(after), "\tprintln(1)") {
		t.Fatalf("tabs lost after revert: %q", after)
	}
}

func TestSelectRawHunksOverlap(t *testing.T) {
	raw := `diff --git a/x.go b/x.go
index 111..222 100644
--- a/x.go
+++ b/x.go
@@ -1,3 +1,4 @@
 keep
-old
+new
+more
 keep
@@ -20,3 +21,3 @@
 a
-b
+c
 d
`
	first := selectRawHunks(raw, 1, 3)
	if !strings.Contains(first, "-old") || strings.Contains(first, "-b") {
		t.Fatalf("first hunk select = %q", first)
	}

	// An expanded UI hunk covering both blocks reverts both.
	both := selectRawHunks(raw, 1, 25)
	if !strings.Contains(both, "-old") || !strings.Contains(both, "-b") {
		t.Fatalf("expanded select = %q", both)
	}

	if got := selectRawHunks(raw, 100, 3); got != "" {
		t.Fatalf("non-overlapping select = %q, want empty", got)
	}
}

func TestRevertHunkUntrackedLive(t *testing.T) {
	dir, _ := historyRepo(t)
	writeFile(t, dir, "scratch.txt", "scratch\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.reloadChanges(ctx, false)
	app.selectDiff("scratch.txt")

	if err := app.revertHunk(ctx, 0); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "scratch.txt")); !os.IsNotExist(err) {
		t.Fatalf("untracked hunk did not delete the file: %v", err)
	}
}

func TestRevertHunkAddedLive(t *testing.T) {
	dir, _ := historyRepo(t)
	writeFile(t, dir, "added.txt", "one\ntwo\n")
	runGit(t, dir, "add", "-A")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.reloadChanges(ctx, false)
	app.selectDiff("added.txt")

	if err := app.revertHunk(ctx, 0); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "added.txt")); !os.IsNotExist(err) {
		t.Fatalf("added hunk did not delete the file: %v", err)
	}

	if findFile(app.changes.Files, "added.txt") != nil {
		t.Fatal("added file still listed after hunk revert")
	}
}

func TestRevertHunkDeletedLive(t *testing.T) {
	dir, _ := historyRepo(t)

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := os.Remove(filepath.Join(dir, "app.go")); err != nil {
		t.Fatal(err)
	}

	app.reloadChanges(ctx, false)
	app.selectDiff("app.go")

	if err := app.revertHunk(ctx, 0); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "app.go")); err != nil {
		t.Fatalf("deleted hunk did not restore the file: %v", err)
	}
}

func TestRevertHunkKeepsOtherHunk(t *testing.T) {
	dir, _ := historyRepo(t)

	var b strings.Builder
	b.WriteString("line 1\nline 2\nline 3\nline 4\nline 5\nline 6\nline 7\nline 8\n")
	for i := 9; i <= 40; i++ {
		b.WriteString(strings.Repeat("x", 1) + "\n")
	}
	_ = b.String()
	writeFile(t, dir, "multi.txt", "a1\na2\na3\na4\na5\na6\na7\na8\na9\na10\na11\na12\na13\na14\na15\na16\na17\na18\na19\na20\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "multi")
	writeFile(t, dir, "multi.txt", "b1\na2\na3\na4\na5\na6\na7\na8\na9\na10\na11\na12\na13\na14\na15\na16\na17\na18\na19\nb20\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.reloadChanges(ctx, false)
	app.selectDiff("multi.txt")

	file := findFile(app.changes.Files, "multi.txt")
	if file == nil || len(file.Hunks) < 2 {
		t.Fatalf("expected two hunks, got %+v", file)
	}

	if err := app.revertHunk(ctx, 0); err != nil {
		t.Fatal(err)
	}

	after, err := os.ReadFile(filepath.Join(dir, "multi.txt")) //nolint:gosec // path is a t.TempDir fixture
	if err != nil {
		t.Fatal(err)
	}

	if strings.HasPrefix(string(after), "b1\n") {
		t.Fatalf("first hunk was not reverted: %q", string(after)[:20])
	}

	if !strings.HasSuffix(strings.TrimSpace(string(after)), "b20") {
		t.Fatalf("second hunk was also reverted: %q", after)
	}
}

func TestRevertHunkUndoRedoLive(t *testing.T) {
	dir, _ := historyRepo(t)
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.reloadChanges(ctx, false)
	app.mode = "changes"
	app.refresh()
	path := app.changes.Files[0].Path
	app.selectDiff(path)

	if err := app.revertHunk(ctx, 0); err != nil {
		t.Fatal(err)
	}

	if len(app.changes.Files) != 0 {
		t.Fatalf("revert left %d files", len(app.changes.Files))
	}

	// Ctrl+Z path: onUndo when the editor is closed restores the hunk.
	app.editing = false
	if err := app.onUndo(ctx); err != nil {
		t.Fatal(err)
	}

	after, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(path))) //nolint:gosec // path is a t.TempDir fixture
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(after), "working change") {
		t.Fatalf("undo did not restore the hunk: %q", after)
	}

	if err := app.onRedo(ctx); err != nil {
		t.Fatal(err)
	}

	after, err = os.ReadFile(filepath.Join(dir, filepath.FromSlash(path))) //nolint:gosec // path is a t.TempDir fixture
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(after), "working change") {
		t.Fatalf("redo did not re-revert the hunk: %q", after)
	}
}

// unifiedLineRevertID mirrors the unified line-revert button in
// components/editor.html: one button per changed (add/del) non-continuation
// row, with id revert-line-<hunk>-<old>-<new>. Context, hunk, meta, expander,
// and wrapped continuation rows render no button.
func unifiedLineRevertID(r diffRow) (string, bool) {
	if r.HasRevert && (r.Kind == rowAdd || r.Kind == rowDel) && !r.Cont {
		return "revert-line-" + strconv.Itoa(r.RevertHunk) + "-" + r.OldNo + "-" + r.NewNo, true
	}

	return "", false
}

// TestUnifiedLineRevertButtons pins the VS Code-style per-line discard
// affordance in the unified diff: exactly one always-visible button per
// change block, on the first added row (or the first row when the block
// deletes only), none on context, hunk, meta, or wrapped continuation rows,
// unique ids, tooltip plus accessible name, and a click that fires the
// existing revert-hunk action.
func TestUnifiedLineRevertButtons(t *testing.T) {
	file := changedFile{
		Path:   "x.go",
		Status: "M",
		Hunks: []hunk{{
			OldStart: 1,
			OldCount: 4,
			NewStart: 1,
			NewCount: 4,
			Lines: []diffLine{
				{Kind: ' ', OldNo: 1, NewNo: 1, Text: "ctx"},
				{Kind: '-', OldNo: 2, Text: "old"},
				{Kind: '+', NewNo: 2, Text: "new value with a long tail so soft wrap splits it"},
				{Kind: ' ', OldNo: 3, NewNo: 3, Text: "ctx2"},
			},
		}},
	}

	rows := diffRowsUnified(file, "go")
	if len(rows) != 5 {
		t.Fatalf("unified rows = %d, want 5 (hunk + 4 lines)", len(rows))
	}

	if rows[0].HasRevert {
		t.Error("hunk header carries a revert marker")
	}

	got := map[string]bool{}
	for _, r := range rows {
		id, ok := unifiedLineRevertID(r)
		if !ok {
			continue
		}

		if got[id] {
			t.Errorf("duplicate unified revert id %q", id)
		}

		got[id] = true
	}

	want := map[string]bool{"revert-line-0--2": true}
	if len(got) != len(want) {
		t.Fatalf("unified revert ids = %v, want %v", got, want)
	}

	for id := range want {
		if !got[id] {
			t.Errorf("missing unified revert id %q (got %v)", id, got)
		}
	}

	// Narrow columns force wrapped continuations. They must not duplicate
	// button ids: only the first visual chunk keeps the discard arrow.
	wrapped := wrapDiffRows(rows, 8, 8, true)
	wrappedGot := map[string]bool{}

	for _, r := range wrapped {
		if r.Cont && r.HasRevert {
			t.Errorf("wrapped continuation row keeps a revert marker: kind=%s old=%s new=%s", r.Kind, r.OldNo, r.NewNo)
		}

		id, ok := unifiedLineRevertID(r)
		if !ok {
			continue
		}

		if wrappedGot[id] {
			t.Errorf("wrapped rows duplicate unified revert id %q", id)
		}

		wrappedGot[id] = true
	}

	if len(wrappedGot) != len(want) {
		t.Fatalf("wrapped unified revert ids = %v, want %v", wrappedGot, want)
	}

	dir, _ := historyRepo(t)

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.reloadChanges(ctx, false)
	app.mode = "changes"
	app.setSplit(false)
	app.wrap = false
	app.refresh()

	view := app.View()
	if view.Diff.Split {
		t.Fatal("expected the unified view")
	}

	if !view.Diff.CanRevert {
		t.Fatal("unified diff does not offer hunk revert")
	}

	expected := map[string]int{}

	for _, r := range view.Diff.Rows {
		id, ok := unifiedLineRevertID(r)
		if !ok {
			continue
		}

		if _, dup := expected[id]; dup {
			t.Errorf("duplicate unified revert id %q", id)
		}

		expected[id] = r.RevertHunk
	}

	// One button per change block: the count must match the navigator's
	// own block list for the same file.
	if starts := app.fileNavChangeRows(view.Diff.Selected); len(view.Diff.Rows) == len(app.diffDisplayRows()) {
		if len(expected) != len(starts) {
			t.Errorf("unified revert buttons = %d, want one per block (%d blocks)", len(expected), len(starts))
		}
	}

	if len(expected) == 0 {
		t.Fatal("unified view has no per-line revert buttons")
	}

	// The executed template must emit exactly the expected buttons, each
	// with the Discard changes tooltip and accessible name.
	shell := buildHTML()
	tmpl, err := template.New("shell").Parse(shell)
	if err != nil {
		t.Fatalf("parse shell template: %v", err)
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, &view); err != nil {
		t.Fatalf("execute shell template: %v", err)
	}

	button := regexp.MustCompile(`<button class="line-revert" id="([^"]+)" data-action="([^"]+)" type="button" title="Discard changes" aria-label="Discard changes">`)
	matches := button.FindAllStringSubmatch(rendered.String(), -1)

	if len(matches) != len(expected) {
		t.Fatalf("rendered unified revert buttons = %d, want %d", len(matches), len(expected))
	}

	for _, m := range matches {
		idx, ok := expected[m[1]]
		if !ok {
			t.Errorf("rendered unexpected unified revert button %q", m[1])

			continue
		}

		if m[2] != "revert-hunk:"+strconv.Itoa(idx) {
			t.Errorf("button %q action = %q, want %q", m[1], m[2], "revert-hunk:"+strconv.Itoa(idx))
		}
	}

	// The reveal stays hover (plus keyboard focus) driven: hidden by
	// default, shown on row hover or button focus, pinned to the left
	// gutter without covering the +/- sign or the line text.
	cssBytes, err := os.ReadFile(filepath.Join("components", "diff.css"))
	if err != nil {
		t.Fatalf("read diff.css: %v", err)
	}

	rule := regexp.MustCompile(`(?s)([^{}]*)\{([^{}]*)\}`)
	comment := regexp.MustCompile(`(?s)/\*.*?\*/`)
	cssText := comment.ReplaceAllString(string(cssBytes), "")

	var baseVisible, hoverDanger, focusReveal, gutterLeft bool

	for _, m := range rule.FindAllStringSubmatch(cssText, -1) {
		sels := strings.Split(m[1], ",")
		decl := m[2]

		for _, s := range sels {
			switch strings.Join(strings.Fields(s), " ") {
			case ".line-revert":
				if strings.Contains(decl, "opacity: 1") && strings.Contains(decl, "position: absolute") {
					baseVisible = true
				}
			case ".line-revert:hover":
				if strings.Contains(decl, "#f07178") {
					hoverDanger = true
				}
			case ".line-revert:focus-visible":
				if strings.Contains(decl, "outline:") {
					focusReveal = true
				}
			case ".drow.add .line-revert", ".drow.del .line-revert":
				if strings.Contains(decl, "left: 2px") {
					gutterLeft = true
				}
			}
		}
	}

	if !baseVisible {
		t.Error("diff.css: .line-revert is not statically visible and absolutely positioned by default")
	}

	if !hoverDanger {
		t.Error("diff.css: .line-revert:hover does not recolor the button")
	}

	if !focusReveal {
		t.Error("diff.css: .line-revert:focus-visible has no outline for keyboard users")
	}

	if !gutterLeft {
		t.Error("diff.css: changed unified rows do not pin .line-revert to the left gutter")
	}

	// Boxes must carry exactly the expected buttons, each firing the
	// hunk's revert action from a shared gutter column sized for the
	// gutter, with paint operations emitted for the button.
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	var boxes []ownframe.Box

	for _, b := range app.Page().Boxes() {
		if strings.HasPrefix(b.ID, "revert-line-") {
			boxes = append(boxes, b)
		}
	}

	if len(boxes) != len(expected) {
		t.Fatalf("revert-line boxes = %d, want %d", len(boxes), len(expected))
	}

	minX, maxX := 0.0, 0.0

	for i, b := range boxes {
		idx, ok := expected[b.ID]
		if !ok {
			t.Errorf("unexpected revert-line box %q", b.ID)

			continue
		}

		if b.Action != "revert-hunk:"+strconv.Itoa(idx) {
			t.Errorf("box %q action = %q, want %q", b.ID, b.Action, "revert-hunk:"+strconv.Itoa(idx))
		}

		if len(b.OperationIDs) == 0 {
			t.Errorf("box %q emitted no paint operations", b.ID)
		}

		if b.W > 20 || b.H > 20 {
			t.Errorf("box %q is %.0fx%.0f, too large for the left gutter", b.ID, b.W, b.H)
		}

		if i == 0 {
			minX, maxX = b.X, b.X
		} else {
			minX, maxX = math.Min(minX, b.X), math.Max(maxX, b.X)
		}
	}

	if maxX-minX > 1 {
		t.Errorf("revert-line boxes span %.1f px horizontally, want one gutter column", maxX-minX)
	}

	// The accessible name is the aria-label from the template.
	tree := accessibilityTree(t, app)

	for id := range expected {
		assertAccessibilityName(t, tree, id, "Discard changes")
	}

	// Clicking a per-line button reverts through the existing hunk action.
	var firstID string

	for id := range expected {
		if firstID == "" || id < firstID {
			firstID = id
		}
	}

	box, ok := boxByID(app.Page().Boxes(), firstID)
	if !ok {
		t.Fatalf("missing box %q", firstID)
	}

	if !strings.HasPrefix(box.Action, "revert-hunk:") {
		t.Fatalf("box %q action = %q, want the hunk revert action", firstID, box.Action)
	}

	if err := app.onClick(ctx, ownframe.Box{ID: box.ID, Action: box.Action}); err != nil {
		t.Fatal(err)
	}

	after, err := os.ReadFile(filepath.Join(dir, "README.md")) //nolint:gosec // path is a t.TempDir fixture
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(after), "working change") {
		t.Fatalf("per-line revert kept the edit: %q", after)
	}

	if findFile(app.changes.Files, "README.md") != nil {
		t.Fatalf("per-line revert left the file listed: %v", app.changes.Files)
	}
}

// TestLineRevertStaticDisplay pins the always-visible affordance. The button
// shows on every changed row without hovering (VS Code-style static gutter
// control); hover only recolors it. A .drow:hover reveal rule must stay out:
// Blinkless matches :hover only against the caller-supplied hover id (see
// the :hover row in documentation/compatibility-matrix.md), so it could
// never fire for a hover on a row child.
func TestLineRevertStaticDisplay(t *testing.T) {
	css, err := os.ReadFile("components/diff.css")
	if err != nil {
		t.Fatal(err)
	}

	style := string(css)
	for _, want := range []string{
		"opacity: 1",
		".line-revert:hover",
		".line-revert:focus-visible",
	} {
		if !strings.Contains(style, want) {
			t.Errorf("diff.css misses static display rule %q", want)
		}
	}

	if strings.Contains(style, ".drow:hover .line-revert") {
		t.Error("diff.css still reveals on .drow:hover, which never matches")
	}

	if strings.Contains(style, "opacity: 0") {
		t.Error("diff.css still hides the revert button until hover")
	}
}

// TestOneRevertPerBlock pins the one-button-per-block rule: the marker sits
// on the first added row of a block, or on the first row when the block
// deletes only, so a deletion-only hunk stays revertable.
func TestOneRevertPerBlock(t *testing.T) {
	file := changedFile{
		Path:   "a.go",
		Status: "M",
		Hunks: []hunk{{
			OldStart: 1,
			OldCount: 6,
			NewStart: 1,
			NewCount: 6,
			Lines: []diffLine{
				{Kind: ' ', OldNo: 1, NewNo: 1, Text: "ctx"},
				{Kind: '-', OldNo: 2, Text: "del one"},
				{Kind: '+', NewNo: 2, Text: "add one"},
				{Kind: ' ', OldNo: 3, NewNo: 3, Text: "ctx"},
				{Kind: '-', OldNo: 4, Text: "del only"},
				{Kind: ' ', OldNo: 5, NewNo: 4, Text: "ctx"},
			},
		}},
	}

	for _, split := range []bool{false, true} {
		var rows []diffRow
		if split {
			rows = diffRowsSplit(file, "go")
		} else {
			rows = diffRowsUnified(file, "go")
		}

		var flagged []diffRow
		for _, r := range rows {
			if r.HasRevert {
				flagged = append(flagged, r)
			}
		}

		if len(flagged) != 2 {
			t.Fatalf("split=%v: %d revert markers, want 2 (one per block)", split, len(flagged))
		}

		first, second := flagged[0], flagged[1]
		if split {
			if first.RightNo != "2" || first.RightKind != rowAdd {
				t.Fatalf("split=%v: first marker on right=%q kind=%q, want the added row", split, first.RightNo, first.RightKind)
			}
			if second.LeftNo != "4" || second.RightNo != "" {
				t.Fatalf("split=%v: second marker on left=%q right=%q, want the deletion-only row", split, second.LeftNo, second.RightNo)
			}
		} else {
			if first.Kind != rowAdd || first.NewNo != "2" {
				t.Fatalf("split=%v: first marker on kind=%q new=%q, want the added row", split, first.Kind, first.NewNo)
			}
			if second.Kind != rowDel || second.OldNo != "4" {
				t.Fatalf("split=%v: second marker on kind=%q old=%q, want the deletion-only row", split, second.Kind, second.OldNo)
			}
		}
	}
}
