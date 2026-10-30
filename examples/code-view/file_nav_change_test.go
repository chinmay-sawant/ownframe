package main

import (
	"context"
	"slices"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestFileNavChangeArrows(t *testing.T) {
	app := newTestApp(t)
	path := "backend/probe.go"
	app.changes = changeSet{Files: []changedFile{{Path: path, Status: "M", Add: 2, Hunks: []hunk{
		{NewStart: 3, Lines: []diffLine{{Kind: '+', NewNo: 3, Text: "first"}}},
		{NewStart: 12, Lines: []diffLine{{Kind: '+', NewNo: 12, Text: "second"}}},
	}}}}
	app.syncFileNav()
	app.openFile(path)
	app.wrap = false
	ctx := context.Background()
	if nav := app.View().FileNav; nav.CanPrevChange || !nav.CanNextChange {
		t.Fatalf("initial arrow state = %+v", nav)
	}
	assertCurrentCodeChange(t, app.View().Rows, 3)
	for _, step := range []struct {
		action string
		row    int
		prev   bool
		next   bool
	}{
		{"file-nav-change-next", 11, true, false},
		{"file-nav-change-prev", 2, false, true},
		{"file-nav-change-next", 11, true, false},
	} {
		if err := app.onClick(ctx, ownframe.Box{Action: step.action}); err != nil {
			t.Fatal(err)
		}
		if !app.changeNav.valid || app.changeNav.row != step.row {
			t.Fatalf("%s: navigation = %+v, want row %d", step.action, app.changeNav, step.row)
		}
		if nav := app.View().FileNav; nav.CanPrevChange != step.prev || nav.CanNextChange != step.next {
			t.Fatalf("%s: arrow state = %+v", step.action, nav)
		}
		assertCurrentCodeChange(t, app.View().Rows, step.row+1)
	}
	app.toggleSplit()
	app.openFileIn(1, path)
	if err := app.onClick(ctx, ownframe.Box{Action: "file-nav-change-next"}); err != nil {
		t.Fatal(err)
	}
	if app.changeNav.group != 1 || app.changeNav.row != 11 || app.scrollGroup[1] != app.changeNav.scroll {
		t.Fatalf("split navigation = %+v, scroll = %v", app.changeNav, app.scrollGroup)
	}
	assertCurrentCodeChange(t, app.View().Second.Rows, 12)
	app.selectDiff(path)
	rows := app.fileNavChangeRows(path)
	if len(rows) != 2 {
		t.Fatalf("diff change rows = %v", rows)
	}
	if app.fileNavChangeIndex(path, rows) < 0 {
		if err := app.onClick(ctx, ownframe.Box{Action: "file-nav-change-next"}); err != nil {
			t.Fatal(err)
		}
	}
	if err := app.onClick(ctx, ownframe.Box{Action: "file-nav-change-next"}); err != nil {
		t.Fatal(err)
	}
	if app.changeNav.row != rows[1] || app.changeNav.mode != "changes" {
		t.Fatalf("diff navigation = %+v, want row %d", app.changeNav, rows[1])
	}
	if nav := app.View().FileNav; !nav.CanPrevChange || nav.CanNextChange {
		t.Fatalf("last diff arrow state = %+v", nav)
	}
	var currentChanges int
	for _, row := range app.View().Diff.Rows {
		if diffChangedRow(row) && row.Current {
			currentChanges++
		}
	}
	if currentChanges != 1 {
		t.Fatalf("current visible changed rows = %d", currentChanges)
	}
}

func assertCurrentCodeChange(t *testing.T, rows []codeRow, line int) {
	t.Helper()
	found := false
	for _, row := range rows {
		if row.ChangeCurrent && row.Line+1 != line {
			t.Fatalf("line %d highlighted, want line %d", row.Line+1, line)
		}
		if row.Line+1 == line && row.ChangeCurrent {
			found = true
		}
	}
	if !found {
		t.Fatalf("line %d is not highlighted", line)
	}
}

func TestFileNavChangeArrowLayout(t *testing.T) {
	app := newTestApp(t)
	for _, width := range []int{1200, 800, 350} {
		app.page.SetSize(width, 700)
		if err := app.page.Redraw(context.Background()); err != nil {
			t.Fatal(err)
		}
		boxes := app.page.Boxes()
		bar, ok := findBoxID(boxes, "file-nav")
		if !ok {
			t.Fatalf("width %d: missing navigator", width)
		}
		var end float64
		for _, id := range []string{"file-nav-change-prev", "file-nav-change-next", "file-nav-next"} {
			box, found := findBoxID(boxes, id)
			if !found || box.X < end || box.X+box.W > bar.X+bar.W {
				t.Fatalf("width %d: %s = %+v, bar = %+v, previous end = %v", width, id, box, bar, end)
			}
			end = box.X + box.W
		}
	}
}

func TestCurrentDiffChangeFollowsScroll(t *testing.T) {
	rows := []diffRow{
		{Kind: rowHunk}, {Kind: rowDel}, {Kind: rowAdd}, {Kind: rowCtx},
		{Kind: rowAdd}, {Kind: rowAdd, Cont: true}, {Kind: rowCtx},
	}
	for _, tc := range []struct {
		current int
		want    []bool
	}{
		{0, []bool{false, true, true, false, false, false, false}},
		{2, []bool{false, true, true, false, false, false, false}},
		{4, []bool{false, false, false, false, true, true, false}},
		{5, []bool{false, false, false, false, true, true, false}},
	} {
		visible := append([]diffRow(nil), rows...)
		markCurrentDiffChange(rows, visible, 0, tc.current)
		for i := range visible {
			if visible[i].Current != tc.want[i] {
				t.Fatalf("scroll row %d: row %d current=%v, want %v", tc.current, i, visible[i].Current, tc.want[i])
			}
		}
	}
}

func TestSingleChangeHasNoNextOrPrevious(t *testing.T) {
	app := newTestApp(t)
	path := "backend/probe.go"
	app.changes = changeSet{Files: []changedFile{{Path: path, Status: "M", Hunks: []hunk{{
		NewStart: 1, Lines: []diffLine{{Kind: '+', NewNo: 1, Text: "first"}},
	}}}}}
	app.syncFileNav()
	app.openFile(path)
	if nav := app.View().FileNav; nav.CanPrevChange || nav.CanNextChange {
		t.Fatalf("single change arrow state = %+v", nav)
	}
	for _, split := range []bool{false, true} {
		app.selectDiff(path)
		app.setSplit(split)
		if starts := app.fileNavChangeRows(path); len(starts) != 1 {
			t.Fatalf("split=%v: change blocks = %v", split, starts)
		}
		if nav := app.View().FileNav; nav.CanPrevChange || nav.CanNextChange {
			t.Fatalf("split=%v: single diff arrow state = %+v", split, nav)
		}
	}
}

func TestOpeningChangedFileRevealsItsOnlyChange(t *testing.T) {
	app := newTestApp(t)
	path := "backend/probe.go"
	app.changes = changeSet{Files: []changedFile{{Path: path, Status: "M", Hunks: []hunk{{
		NewStart: 120, Lines: []diffLine{{Kind: '+', NewNo: 120, Text: "change"}},
	}}}}}
	app.syncFileNav()
	for _, wrap := range []bool{false, true} {
		app.wrap = wrap
		app.openFile(path)
		if !app.changeNav.valid || app.scrollY == 0 {
			t.Fatalf("wrap=%v: change not reached: scroll=%d navigation=%+v", wrap, app.scrollY, app.changeNav)
		}
		assertCurrentCodeChange(t, app.View().Rows, 120)
		app.fileNavOpen()
		assertCurrentCodeChange(t, app.View().Rows, 120)
	}
}

func TestFileArrowsRevealFirstChange(t *testing.T) {
	app := newTestApp(t)
	app.wrap = false
	app.changes = changeSet{Files: []changedFile{
		{Path: "backend/probe.go", Hunks: []hunk{{NewStart: 120, Lines: []diffLine{{Kind: '+', NewNo: 120}}}}},
		{Path: "backend/store.go", Hunks: []hunk{{NewStart: 60, Lines: []diffLine{{Kind: '+', NewNo: 60}}}}},
	}}
	app.syncFileNav()
	app.openFile("backend/probe.go")
	app.fileNavStep(1)
	if app.activeDoc().Path != "backend/store.go" {
		t.Fatalf("next file = %s", app.activeDoc().Path)
	}
	assertCurrentCodeChange(t, app.View().Rows, 60)
	app.fileNavStep(-1)
	if app.activeDoc().Path != "backend/probe.go" {
		t.Fatalf("previous file = %s", app.activeDoc().Path)
	}
	assertCurrentCodeChange(t, app.View().Rows, 120)
}

func TestDiffChangeStartsGroupsContiguousRows(t *testing.T) {
	rows := []diffRow{
		{Kind: rowHunk}, {Kind: rowCtx},
		{Kind: rowDel}, {Kind: rowDel, Cont: true}, {Kind: rowAdd},
		{Kind: rowCtx}, {Kind: "split", LeftKind: rowDel, RightKind: rowAdd},
		{Kind: "split", LeftKind: rowDel, RightKind: rowAdd, LeftCont: true},
		{Kind: rowCtx},
	}
	if got := diffChangeStarts(rows); !slices.Equal(got, []int{2, 6}) {
		t.Fatalf("change starts = %v, want [2 6]", got)
	}
}

func TestDiffNavigatorMovesBetweenChangesInOneHunk(t *testing.T) {
	app := newTestApp(t)
	path := "backend/probe.go"
	app.changes = changeSet{Files: []changedFile{{Path: path, Status: "M", Hunks: []hunk{{
		NewStart: 1, Lines: []diffLine{
			{Kind: '+', NewNo: 1, Text: "first"},
			{Kind: ' ', OldNo: 1, NewNo: 2, Text: "unchanged"},
			{Kind: '+', NewNo: 3, Text: "second"},
		},
	}}}}}
	app.syncFileNav()
	for _, split := range []bool{false, true} {
		app.selectDiff(path)
		app.setSplit(split)
		starts := app.fileNavChangeRows(path)
		if len(starts) != 2 {
			t.Fatalf("split=%v: starts = %v", split, starts)
		}
		if nav := app.View().FileNav; nav.CanPrevChange || !nav.CanNextChange {
			t.Fatalf("split=%v: first change arrows = %+v", split, nav)
		}
		app.fileNavChangeStep(1)
		if app.changeNav.row != starts[1] {
			t.Fatalf("split=%v: selected row = %d, want %d", split, app.changeNav.row, starts[1])
		}
		if nav := app.View().FileNav; !nav.CanPrevChange || nav.CanNextChange {
			t.Fatalf("split=%v: last change arrows = %+v", split, nav)
		}
	}
}

func TestSelectedFileTracksNavigator(t *testing.T) {
	app := newTestApp(t)
	app.setMode(context.Background(), "changes")
	for i := 0; i < 2; i++ {
		view := app.View()
		found := false
		for _, file := range view.Diff.Files {
			if file.Selected {
				if file.Path != view.Diff.Selected {
					t.Fatalf("highlighted %q while viewing %q", file.Path, view.Diff.Selected)
				}
				found = true
			}
		}
		if !found {
			t.Fatal("current changed file is not highlighted")
		}
		app.fileNavStep(1)
	}
}

func TestFileNavigatorRemainsAvailableWhileEditing(t *testing.T) {
	ctx := context.Background()
	app := newTestApp(t)
	app.openFile("backend/probe.go")
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	body, ok := boxByID(app.Page().Boxes(), "code-body")
	if !ok {
		t.Fatal("missing editor body")
	}
	if err := app.Page().Edit().Press(ctx, body.X+editTextLeft+codeCellW, body.Y+rowH/2); err != nil {
		t.Fatal(err)
	}
	if !app.editing || !app.View().FileNav.Open {
		t.Fatalf("editing=%v, navigator=%+v", app.editing, app.View().FileNav)
	}
	if _, ok := boxByID(app.Page().Boxes(), "file-nav-next"); !ok {
		t.Fatal("file navigator is missing while editing")
	}
	path := app.View().FileNav.Name
	if err := app.onClick(ctx, ownframe.Box{Action: "file-nav-next"}); err != nil {
		t.Fatal(err)
	}
	if app.View().FileNav.Name == path {
		t.Fatal("navigator did not move to the next file")
	}
}

func TestHunkArrowsTrackChangeBlocks(t *testing.T) {
	ctx := context.Background()
	newDiffApp := func(files []changedFile) *App {
		app := newTestApp(t)
		app.changes = changeSet{Files: files}
		app.syncFileNav()
		return app
	}
	singleAdd := []changedFile{{Path: "a.go", Status: "M", Hunks: []hunk{{
		NewStart: 1, Lines: []diffLine{{Kind: '+', NewNo: 1, Text: "one"}},
	}}}}
	addDel := []changedFile{{Path: "a.go", Status: "M", Hunks: []hunk{{
		NewStart: 1, Lines: []diffLine{
			{Kind: '-', OldNo: 1, Text: "old"},
			{Kind: '+', NewNo: 1, Text: "new"},
		},
	}}}}
	twoBlocks := []changedFile{{Path: "a.go", Status: "M", Hunks: []hunk{{
		NewStart: 1, Lines: []diffLine{
			{Kind: '+', NewNo: 1, Text: "first"},
			{Kind: ' ', OldNo: 1, NewNo: 2, Text: "keep"},
			{Kind: '-', OldNo: 2, Text: "old"},
			{Kind: '-', OldNo: 3, Text: "older"},
			{Kind: '+', NewNo: 3, Text: "new"},
			{Kind: '+', NewNo: 4, Text: "newer"},
		},
	}}}}

	for _, tc := range []struct {
		name  string
		files []changedFile
	}{
		{"single addition", singleAdd},
		{"adjacent addition and deletion", addDel},
	} {
		for _, split := range []bool{false, true} {
			app := newDiffApp(tc.files)
			app.selectDiff("a.go")
			app.setSplit(split)
			if nav := app.View().FileNav; nav.CanPrevChange || nav.CanNextChange {
				t.Fatalf("%s split=%v: file nav = %+v, want both off", tc.name, split, nav)
			}
			if diff := app.View().Diff; diff.CanPrevChange || diff.CanNextChange {
				t.Fatalf("%s split=%v: crumbs = prev %v next %v, want both off",
					tc.name, split, diff.CanPrevChange, diff.CanNextChange)
			}
			before := app.scrollY
			app.jumpHunk(1)
			app.jumpHunk(-1)
			if app.scrollY != before {
				t.Fatalf("%s split=%v: single block jump moved %d to %d", tc.name, split, before, app.scrollY)
			}
		}
	}

	for _, split := range []bool{false, true} {
		app := newDiffApp(twoBlocks)
		app.selectDiff("a.go")
		app.setSplit(split)
		if nav := app.View().FileNav; nav.CanPrevChange || !nav.CanNextChange {
			t.Fatalf("split=%v: initial file nav = %+v, want prev off next on", split, nav)
		}
		if diff := app.View().Diff; diff.CanPrevChange || !diff.CanNextChange {
			t.Fatalf("split=%v: initial crumbs = prev %v next %v, want prev off next on",
				split, diff.CanPrevChange, diff.CanNextChange)
		}
		if err := app.onClick(ctx, ownframe.Box{Action: "hunk-next"}); err != nil {
			t.Fatal(err)
		}
		if nav := app.View().FileNav; !nav.CanPrevChange || nav.CanNextChange {
			t.Fatalf("split=%v: last file nav = %+v, want prev on next off", split, nav)
		}
		if diff := app.View().Diff; !diff.CanPrevChange || diff.CanNextChange {
			t.Fatalf("split=%v: last crumbs = prev %v next %v, want prev on next off",
				split, diff.CanPrevChange, diff.CanNextChange)
		}
		var current int
		for _, row := range app.View().Diff.Rows {
			if row.Current && !diffChangedRow(row) {
				t.Fatalf("split=%v: non-change row is current: %+v", split, row)
			}
			if diffChangedRow(row) && row.Current {
				current++
			}
		}
		if current == 0 {
			t.Fatalf("split=%v: no current rows after hunk-next", split)
		}
		starts := app.fileNavChangeRows("a.go")
		if len(starts) != 2 || !app.changeNav.valid || app.changeNav.row != starts[1] {
			t.Fatalf("split=%v: navigation = %+v, starts = %v, want row %d", split, app.changeNav, starts, starts[1])
		}
		if err := app.onClick(ctx, ownframe.Box{Action: "hunk-prev"}); err != nil {
			t.Fatal(err)
		}
		if diff := app.View().Diff; diff.CanPrevChange || !diff.CanNextChange {
			t.Fatalf("split=%v: back crumbs = prev %v next %v, want prev off next on",
				split, diff.CanPrevChange, diff.CanNextChange)
		}
	}
}
