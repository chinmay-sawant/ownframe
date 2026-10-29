package main

import (
	"bytes"
	"context"
	"html/template"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// splitDividerRevertID mirrors the split divider revert button in
// components/editor.html: one button per changed (del/add) non-continuation
// split row, sitting on the center divider, with id
// revert-line-<hunk>-<left>-<right>. Context, hunk, meta, expander,
// fully padded, and wrapped continuation rows render no button.
func splitDividerRevertID(r diffRow) (string, bool) {
	if r.Kind != "split" || !r.HasRevert {
		return "", false
	}

	if r.LeftKind != rowDel && r.RightKind != rowAdd {
		return "", false
	}

	if r.LeftPad && r.RightPad {
		return "", false
	}

	if r.LeftCont || r.RightCont {
		return "", false
	}

	return "revert-line-" + strconv.Itoa(r.RevertHunk) + "-" + r.LeftNo + "-" + r.RightNo, true
}

// splitDividerFixture covers every changed split shape in one block: a
// paired row, a left-only tail, another paired row, a right-only tail, and
// context rows on both sides that must stay button-free. One block means one
// divider button, on the first row with added (right) content.
func splitDividerFixture() changedFile {
	return changedFile{
		Path:   "a.go",
		Status: "M",
		Hunks: []hunk{{
			OldStart: 1,
			OldCount: 5,
			NewStart: 1,
			NewCount: 5,
			Lines: []diffLine{
				{Kind: ' ', OldNo: 1, NewNo: 1, Text: "keep top"},
				{Kind: '-', OldNo: 2, Text: "old one"},
				{Kind: '-', OldNo: 3, Text: "old two"},
				{Kind: '+', NewNo: 2, Text: "new one"},
				{Kind: '-', OldNo: 4, Text: "old three"},
				{Kind: '+', NewNo: 3, Text: "new two"},
				{Kind: '+', NewNo: 4, Text: "new three"},
				{Kind: ' ', OldNo: 5, NewNo: 5, Text: "keep bottom"},
			},
		}},
	}
}

// splitDividerWantIDs is the divider button for splitDividerFixture: one
// block carries one button, on its first row with added content (the paired
// 2/2 row). Tails never collide because only one row per block keeps the
// marker.
func splitDividerWantIDs() map[string]bool {
	return map[string]bool{
		"revert-line-0-2-2": true,
	}
}

// TestSplitDividerRevertButtons pins the VS Code-style revert affordance on
// the split center divider: exactly one always-visible button per change
// block on the divider of its first row with added content, none on context
// or wrapped continuation rows, unique ids, tooltip plus accessible name,
// and geometry that lands on the visible divider at wide and narrow widths
// without shifting the 20px row grid.
func TestSplitDividerRevertButtons(t *testing.T) {
	rows := diffRowsSplit(splitDividerFixture(), "go")
	if len(rows) != 7 {
		t.Fatalf("split rows = %d, want 7 (hunk + 2 ctx + 4 changed)", len(rows))
	}

	got := map[string]bool{}

	for _, r := range rows {
		id, ok := splitDividerRevertID(r)
		if !ok {
			continue
		}

		if got[id] {
			t.Errorf("duplicate split divider revert id %q", id)
		}

		got[id] = true
	}

	want := splitDividerWantIDs()
	if len(got) != len(want) {
		t.Fatalf("split divider revert ids = %v, want %v", got, want)
	}

	for id := range want {
		if !got[id] {
			t.Errorf("missing split divider revert id %q (got %v)", id, got)
		}
	}

	for _, r := range rows {
		if r.Kind == "split" && r.LeftKind == rowCtx && r.HasRevert {
			t.Errorf("context split row left=%s right=%s carries a revert marker", r.LeftNo, r.RightNo)
		}
	}

	// Narrow columns force wrapped continuations. They must not duplicate
	// button ids: only the first visual chunk keeps the discard arrow.
	longFile := changedFile{
		Path:   "w.go",
		Status: "M",
		Hunks: []hunk{{
			OldStart: 1,
			OldCount: 1,
			NewStart: 1,
			NewCount: 1,
			Lines: []diffLine{
				{Kind: '-', OldNo: 1, Text: "old value with a long tail so soft wrap splits it"},
				{Kind: '+', NewNo: 1, Text: "new value with a long tail so soft wrap splits it"},
			},
		}},
	}

	wrapped := wrapDiffRows(diffRowsSplit(longFile, "go"), 8, 8, true)
	wrappedGot := map[string]bool{}

	for _, r := range wrapped {
		if (r.LeftCont || r.RightCont) && r.HasRevert {
			t.Errorf("wrapped continuation row keeps a revert marker: left=%s right=%s", r.LeftNo, r.RightNo)
		}

		id, ok := splitDividerRevertID(r)
		if !ok {
			continue
		}

		if wrappedGot[id] {
			t.Errorf("wrapped rows duplicate split divider revert id %q", id)
		}

		wrappedGot[id] = true
	}

	if len(wrappedGot) != 1 || !wrappedGot["revert-line-0-1-1"] {
		t.Fatalf("wrapped split divider revert ids = %v, want [revert-line-0-1-1]", wrappedGot)
	}

	app := newTestApp(t)
	app.wrap = false
	app.changes = changeSet{Files: []changedFile{splitDividerFixture()}}
	app.syncFileNav()
	app.selectDiff("a.go")
	app.setSplit(true)
	app.refresh()

	view := app.View()
	if !view.Diff.Split {
		t.Fatal("expected the split view")
	}

	if !view.Diff.CanRevert {
		t.Fatal("split diff does not offer hunk revert")
	}

	expected := map[string]int{}

	for _, r := range view.Diff.Rows {
		id, ok := splitDividerRevertID(r)
		if !ok {
			continue
		}

		if _, dup := expected[id]; dup {
			t.Errorf("duplicate split divider revert id %q", id)
		}

		expected[id] = r.RevertHunk
	}

	// One button per change block: the count must match the navigator's
	// own block list for the same file.
	if starts := app.fileNavChangeRows(view.Diff.Selected); len(view.Diff.Rows) == len(app.diffDisplayRows()) {
		if len(expected) != len(starts) {
			t.Errorf("split divider buttons = %d, want one per block (%d blocks)", len(expected), len(starts))
		}
	}

	if len(expected) != len(want) {
		t.Fatalf("split view divider revert ids = %v, want %v", expected, want)
	}

	for id := range want {
		if _, ok := expected[id]; !ok {
			t.Errorf("split view is missing divider button %q", id)
		}
	}

	// The executed template must emit exactly the expected divider buttons,
	// each with the Discard changes tooltip and accessible name.
	shell := buildHTML()
	tmpl, err := template.New("shell").Parse(shell)
	if err != nil {
		t.Fatalf("parse shell template: %v", err)
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, &view); err != nil {
		t.Fatalf("execute shell template: %v", err)
	}

	button := regexp.MustCompile(`<span class="dmid"><button class="line-revert divider" id="([^"]+)" data-action="([^"]+)" type="button" title="Discard changes" aria-label="Discard changes">`)
	matches := button.FindAllStringSubmatch(rendered.String(), -1)

	if len(matches) != len(expected) {
		t.Fatalf("rendered split divider buttons = %d, want %d", len(matches), len(expected))
	}

	for _, m := range matches {
		idx, ok := expected[m[1]]
		if !ok {
			t.Errorf("rendered unexpected split divider button %q", m[1])

			continue
		}

		if m[2] != "revert-hunk:"+strconv.Itoa(idx) {
			t.Errorf("button %q action = %q, want %q", m[1], m[2], "revert-hunk:"+strconv.Itoa(idx))
		}
	}

	// The split template carries exactly one divider emission site, a
	// zero-width slot between the cells, and the old gutter-side split
	// buttons are gone.
	htmlBytes, err := os.ReadFile(filepath.Join("components", "editor.html"))
	if err != nil {
		t.Fatalf("read editor.html: %v", err)
	}

	htmlSrc := string(htmlBytes)
	if got := strings.Count(htmlSrc, `<span class="dmid">`); got != 1 {
		t.Errorf("editor.html divider slots = %d, want exactly one emission site", got)
	}

	if got := strings.Count(htmlSrc, `class="line-revert divider"`); got != 1 {
		t.Errorf("editor.html divider buttons = %d, want exactly one emission site", got)
	}

	if strings.Contains(htmlSrc, "line-revert left") || strings.Contains(htmlSrc, "line-revert right") {
		t.Error("split view still emits gutter-side revert buttons; the divider button is the only split affordance")
	}

	// The zero-width slot sits exactly between the cells without taking
	// space, and the arrow centers on it with a fixed half-width offset.
	cssBytes, err := os.ReadFile(filepath.Join("components", "diff.css"))
	if err != nil {
		t.Fatalf("read diff.css: %v", err)
	}

	// Strip comments so a rule's selector text is bare, like the
	// shell-parser tests do.
	rule := regexp.MustCompile(`(?s)([^{}]*)\{([^{}]*)\}`)
	noComments := regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAllString(string(cssBytes), "")

	var slotZero, arrowCenter bool

	for _, m := range rule.FindAllStringSubmatch(noComments, -1) {
		for _, s := range strings.Split(m[1], ",") {
			sel := strings.Join(strings.Fields(s), " ")

			if sel == ".drow.split .dmid" &&
				strings.Contains(m[2], "flex: 0 0 0") &&
				strings.Contains(m[2], "position: relative") {
				slotZero = true
			}

			if sel == ".drow.split .line-revert.divider" &&
				strings.Contains(m[2], "left: -8px") {
				arrowCenter = true
			}
		}
	}

	if !slotZero {
		t.Error("diff.css: .drow.split .dmid is not a zero-width positioned slot")
	}

	if !arrowCenter {
		t.Error("diff.css: split divider button is not centered on the slot with left: -8px")
	}

	// Boxes must carry exactly the expected divider buttons at wide and
	// narrow widths. Each button fires the hunk's revert action, centers on
	// its row's divider, and emits paint operations for the button. The
	// divider comes from the row's own full-width edit-hit box, so the
	// check holds at every width without assuming absolute coordinates.
	// Layout boxes stop at the paint boundary: whether the glyph rasterizes
	// is owned by the shared paint path, not this affordance.
	ctx := context.Background()

	for _, width := range []int{1200, 800} {
		wapp := newTestApp(t)
		wapp.wrap = false
		wapp.changes = changeSet{Files: []changedFile{splitDividerFixture()}}
		wapp.syncFileNav()
		wapp.selectDiff("a.go")
		wapp.setSplit(true)
		wapp.refresh()
		wapp.Page().SetSize(width, 700)

		if err := wapp.Page().Redraw(ctx); err != nil {
			t.Fatalf("width %d: %v", width, err)
		}

		body, ok := boxByID(wapp.Page().Boxes(), "code-body")
		if !ok {
			t.Fatalf("width %d: missing editor body", width)
		}

		var boxes, hits []ownframe.Box

		for _, b := range wapp.Page().Boxes() {
			switch {
			case strings.HasPrefix(b.ID, "revert-line-"):
				boxes = append(boxes, b)
			case strings.HasPrefix(b.ID, "diff-edit-line-"):
				hits = append(hits, b)
			}
		}

		if len(boxes) != len(expected) {
			t.Fatalf("width %d: revert-line boxes = %d, want %d", width, len(boxes), len(expected))
		}

		minX, maxX := 0.0, 0.0
		ys := make([]float64, 0, len(boxes))

		for i, b := range boxes {
			idx, ok := expected[b.ID]
			if !ok {
				t.Errorf("width %d: unexpected revert-line box %q", width, b.ID)

				continue
			}

			if b.Action != "revert-hunk:"+strconv.Itoa(idx) {
				t.Errorf("width %d: box %q action = %q, want %q", width, b.ID, b.Action, "revert-hunk:"+strconv.Itoa(idx))
			}

			if len(b.OperationIDs) == 0 {
				t.Errorf("width %d: box %q emitted no paint operations", width, b.ID)
			}

			if math.Abs(b.W-16) > 1 || math.Abs(b.H-16) > 1 {
				t.Errorf("width %d: box %q is %.0fx%.0f, want the 16px divider arrow", width, b.ID, b.W, b.H)
			}

			// The row's edit-hit box spans the row, so its midpoint is the
			// visible divider the slot sits on.
			var divider float64

			found := false

			for _, h := range hits {
				if h.Y <= b.Y && b.Y+b.H <= h.Y+h.H {
					divider = h.X + h.W/2
					found = true

					break
				}
			}

			if !found {
				t.Errorf("width %d: box %q matches no row", width, b.ID)

				continue
			}

			if math.Abs(b.X+b.W/2-divider) > 1.5 {
				t.Errorf("width %d: box %q center is %.1f, want the row divider %.1f", width, b.ID, b.X+b.W/2, divider)
			}

			if b.X < divider-12 || b.X+b.W > divider+44+1 {
				t.Errorf("width %d: box %q at x=%.1f w=%.0f reaches past the gutter into a cell", width, b.ID, b.X, b.W)
			}

			if b.X < body.X || b.X+b.W > body.X+body.W+1 || b.Y < body.Y || b.Y+b.H > body.Y+body.H+1 {
				t.Errorf("width %d: box %q at (%.1f,%.1f) escapes the editor body", width, b.ID, b.X, b.Y)
			}

			if i == 0 {
				minX, maxX = b.X, b.X
			} else {
				minX, maxX = math.Min(minX, b.X), math.Max(maxX, b.X)
			}

			ys = append(ys, b.Y)
		}

		if maxX-minX > 1 {
			t.Errorf("width %d: divider buttons span %.1f px horizontally, want one divider column", width, maxX-minX)
		}

		sort.Float64s(ys)

		for _, y := range ys[1:] {
			if d := y - ys[0]; math.Abs(d-math.Round(d/20)*20) > 1 {
				t.Errorf("width %d: divider button at y=%.1f breaks the 20px row grid", width, y)
			}
		}
	}

	// The accessible name is the aria-label from the template.
	tree := accessibilityTree(t, app)

	for id := range expected {
		assertAccessibilityName(t, tree, id, "Discard changes")
	}
}

// TestSplitDividerRevertClick checks that a split divider button reverts
// through the existing hunk action without changing revert semantics.
func TestSplitDividerRevertClick(t *testing.T) {
	dir, _ := historyRepo(t)

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.reloadChanges(ctx, false)
	app.mode = "changes"
	app.refresh()
	app.selectDiff("README.md")
	app.setSplit(true)
	app.wrap = false
	app.refresh()

	if !app.View().Diff.CanRevert {
		t.Fatal("split diff does not offer hunk revert")
	}

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	var target *ownframe.Box

	for _, b := range app.Page().Boxes() {
		if !strings.HasPrefix(b.ID, "revert-line-") {
			continue
		}

		if target == nil || b.ID < target.ID {
			box := b
			target = &box
		}
	}

	if target == nil {
		t.Fatal("split view has no divider revert button")
	}

	if !strings.HasPrefix(target.Action, "revert-hunk:") {
		t.Fatalf("box %q action = %q, want the hunk revert action", target.ID, target.Action)
	}

	if err := app.onClick(ctx, ownframe.Box{ID: target.ID, Action: target.Action}); err != nil {
		t.Fatal(err)
	}

	after, err := os.ReadFile(filepath.Join(dir, "README.md")) //nolint:gosec // path is a t.TempDir fixture
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(after), "working change") {
		t.Fatalf("divider revert kept the edit: %q", after)
	}

	if findFile(app.changes.Files, "README.md") != nil {
		t.Fatalf("divider revert left the file listed: %v", app.changes.Files)
	}
}
