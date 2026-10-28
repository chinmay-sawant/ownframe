package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestDiffExpandIdentifyGaps(t *testing.T) {
	t.Parallel()

	// 1. Synthetic file with gap before first hunk, gap between hunks, and gap after last hunk.
	file := changedFile{
		Path: "test.go",
		Hunks: []hunk{
			{OldStart: 10, OldCount: 5, NewStart: 10, NewCount: 5},
			{OldStart: 30, OldCount: 4, NewStart: 30, NewCount: 4},
		},
		TotalLines: 40,
	}

	gaps := identifyGaps(file)
	if len(gaps) != 3 {
		t.Fatalf("len(gaps) = %d, want 3", len(gaps))
	}

	// Gap 0: before first hunk (lines 1..9, count 9)
	if gaps[0].Index != 0 || !gaps[0].Before || gaps[0].NewStart != 1 || gaps[0].NewEnd != 9 || gaps[0].Count != 9 {
		t.Errorf("gap 0 = %+v, want 1..9 (count 9, Before: true)", gaps[0])
	}

	// Gap 1: between hunks (lines 15..29, count 15)
	// Hunk 0 ends at 10 + 5 - 1 = 14 (next: 15). Hunk 1 starts at 30.
	if gaps[1].Index != 1 || gaps[1].Before || gaps[1].After || gaps[1].NewStart != 15 || gaps[1].NewEnd != 29 || gaps[1].Count != 15 {
		t.Errorf("gap 1 = %+v, want 15..29 (count 15)", gaps[1])
	}

	// Gap 2: after last hunk (lines 34..40, count 7)
	// Hunk 1 ends at 30 + 4 - 1 = 33 (next: 34). Total lines: 40.
	if gaps[2].Index != 2 || !gaps[2].After || gaps[2].NewStart != 34 || gaps[2].NewEnd != 40 || gaps[2].Count != 7 {
		t.Errorf("gap 2 = %+v, want 34..40 (count 7, After: true)", gaps[2])
	}
}

func TestDiffExpandSampleDiffGaps(t *testing.T) {
	t.Parallel()

	cs := sampleChangeSet()

	// config/services.yaml has gap before first hunk (lines 1..6) and between hunks (lines 16..21)
	services := findFile(cs.Files, "config/services.yaml")
	if services == nil {
		t.Fatal("config/services.yaml not found in sample diff")
	}

	gaps := identifyGaps(*services)
	if len(gaps) < 2 {
		t.Fatalf("services.yaml gaps = %d, want at least 2", len(gaps))
	}

	if !gaps[0].Before || gaps[0].NewStart != 1 || gaps[0].NewEnd != 6 || gaps[0].Count != 6 {
		t.Errorf("services.yaml gap 0 = %+v, want lines 1..6 (count 6, Before: true)", gaps[0])
	}

	if gaps[1].NewStart != 16 || gaps[1].NewEnd != 21 || gaps[1].Count != 6 {
		t.Errorf("services.yaml gap 1 = %+v, want lines 16..21 (count 6)", gaps[1])
	}

	// docs/deployment.md has 2 hunks: hunk 0 (1..7) and hunk 1 (10..101)
	// Gap between hunks is lines 8..9 (count 2)
	deployment := findFile(cs.Files, "docs/deployment.md")
	if deployment == nil {
		t.Fatal("docs/deployment.md not found in sample diff")
	}

	depGaps := identifyGaps(*deployment)
	if len(depGaps) != 1 {
		t.Fatalf("deployment.md gaps = %d, want 1", len(depGaps))
	}

	if depGaps[0].NewStart != 8 || depGaps[0].NewEnd != 9 || depGaps[0].Count != 2 {
		t.Errorf("deployment.md gap = %+v, want lines 8..9 (count 2)", depGaps[0])
	}

	// tools/report.py with total lines = 189 has a gap after the last hunk (lines 187..189)
	report := findFile(cs.Files, "tools/report.py")
	if report == nil {
		t.Fatal("tools/report.py not found in sample diff")
	}

	reportGaps := identifyGapsWithTotal(*report, 189)
	if len(reportGaps) != 1 {
		t.Fatalf("report.py gaps = %d, want 1", len(reportGaps))
	}

	if !reportGaps[0].After || reportGaps[0].NewStart != 187 || reportGaps[0].NewEnd != 189 || reportGaps[0].Count != 3 {
		t.Errorf("report.py gap = %+v, want lines 187..189 (count 3, After: true)", reportGaps[0])
	}
}

func TestDiffExpandHunkGapDown(t *testing.T) {
	t.Parallel()

	// File with 2 hunks: Hunk 0 (lines 1..5), Hunk 1 (lines 31..35).
	// Gap between hunks: lines 6..30 (count 25).
	file := changedFile{
		Path: "test.go",
		Hunks: []hunk{
			{OldStart: 1, OldCount: 5, NewStart: 1, NewCount: 5, Lines: makeLines(1, 5)},
			{OldStart: 31, OldCount: 5, NewStart: 31, NewCount: 5, Lines: makeLines(31, 35)},
		},
	}

	gaps := identifyGaps(file)
	if len(gaps) != 1 {
		t.Fatalf("len(gaps) = %d, want 1", len(gaps))
	}

	gap := gaps[0]
	if gap.Count != 25 {
		t.Fatalf("gap.Count = %d, want 25", gap.Count)
	}

	// Expand down 20 lines (lines 6..25)
	contextLines := make([]string, 20)
	for i := range contextLines {
		contextLines[i] = fmt.Sprintf("line %d", 6+i)
	}

	expandHunkGap(&file, gap, expandDown, contextLines)

	// Hunk 0 should now have 5 + 20 = 25 lines (ending at line 25)
	if file.Hunks[0].NewCount != 25 {
		t.Fatalf("hunk 0 NewCount = %d, want 25", file.Hunks[0].NewCount)
	}

	if len(file.Hunks) != 2 {
		t.Fatalf("len(Hunks) = %d, want 2", len(file.Hunks))
	}

	// Remaining gap should now be lines 26..30 (count 5)
	gaps2 := identifyGaps(file)
	if len(gaps2) != 1 {
		t.Fatalf("len(gaps2) = %d, want 1", len(gaps2))
	}

	if gaps2[0].NewStart != 26 || gaps2[0].NewEnd != 30 || gaps2[0].Count != 5 {
		t.Fatalf("remaining gap = %+v, want lines 26..30 (count 5)", gaps2[0])
	}

	// Expand down remaining 5 lines (lines 26..30)
	remainingLines := make([]string, 5)
	for i := range remainingLines {
		remainingLines[i] = fmt.Sprintf("line %d", 26+i)
	}

	expandHunkGap(&file, gaps2[0], expandDown, remainingLines)

	// Since lines 26..30 touch hunk 1 (starts at 31), the two hunks should MERGE into 1 hunk!
	if len(file.Hunks) != 1 {
		t.Fatalf("after merge, len(Hunks) = %d, want 1", len(file.Hunks))
	}

	// Merged hunk should cover lines 1..35 (35 lines total)
	if file.Hunks[0].NewStart != 1 || file.Hunks[0].NewCount != 35 || len(file.Hunks[0].Lines) != 35 {
		t.Fatalf("merged hunk = %+v, want NewStart: 1, NewCount: 35, Lines: 35", file.Hunks[0])
	}

	// No remaining gaps
	if g := identifyGaps(file); len(g) != 0 {
		t.Fatalf("expected 0 gaps after full expansion, got %d", len(g))
	}
}

func TestDiffExpandHunkGapUp(t *testing.T) {
	t.Parallel()

	// File with 2 hunks: Hunk 0 (lines 1..5), Hunk 1 (lines 31..35).
	// Gap between hunks: lines 6..30 (count 25).
	file := changedFile{
		Path: "test.go",
		Hunks: []hunk{
			{OldStart: 1, OldCount: 5, NewStart: 1, NewCount: 5, Lines: makeLines(1, 5)},
			{OldStart: 31, OldCount: 5, NewStart: 31, NewCount: 5, Lines: makeLines(31, 35)},
		},
	}

	gaps := identifyGaps(file)
	if len(gaps) != 1 {
		t.Fatalf("len(gaps) = %d, want 1", len(gaps))
	}

	// Expand up 20 lines (lines 11..30)
	contextLines := make([]string, 20)
	for i := range contextLines {
		contextLines[i] = fmt.Sprintf("line %d", 11+i)
	}

	expandHunkGap(&file, gaps[0], expandUp, contextLines)

	// Hunk 1 should now start at line 11 with 25 lines (lines 11..35)
	if file.Hunks[1].NewStart != 11 || file.Hunks[1].NewCount != 25 {
		t.Fatalf("hunk 1 start/count = %d/%d, want 11/25", file.Hunks[1].NewStart, file.Hunks[1].NewCount)
	}

	// Remaining gap is now lines 6..10 (count 5)
	gaps2 := identifyGaps(file)
	if len(gaps2) != 1 || gaps2[0].NewStart != 6 || gaps2[0].NewEnd != 10 || gaps2[0].Count != 5 {
		t.Fatalf("remaining gap = %+v, want lines 6..10 (count 5)", gaps2)
	}

	// Expand up remaining 5 lines (lines 6..10)
	remLines := make([]string, 5)
	for i := range remLines {
		remLines[i] = fmt.Sprintf("line %d", 6+i)
	}

	expandHunkGap(&file, gaps2[0], expandUp, remLines)

	// Now hunks touch and merge into 1 hunk
	if len(file.Hunks) != 1 {
		t.Fatalf("after merge, len(Hunks) = %d, want 1", len(file.Hunks))
	}

	if file.Hunks[0].NewStart != 1 || file.Hunks[0].NewCount != 35 {
		t.Fatalf("merged hunk = %+v, want NewStart: 1, NewCount: 35", file.Hunks[0])
	}
}

func TestDiffExpandHunkGapAll(t *testing.T) {
	t.Parallel()

	// File with 2 hunks: Hunk 0 (lines 1..5), Hunk 1 (lines 26..30).
	// Gap: lines 6..25 (count 20).
	file := changedFile{
		Path: "test.go",
		Hunks: []hunk{
			{OldStart: 1, OldCount: 5, NewStart: 1, NewCount: 5, Lines: makeLines(1, 5)},
			{OldStart: 26, OldCount: 5, NewStart: 26, NewCount: 5, Lines: makeLines(26, 30)},
		},
	}

	gaps := identifyGaps(file)
	if len(gaps) != 1 || gaps[0].Count != 20 {
		t.Fatalf("gaps = %+v", gaps)
	}

	allLines := make([]string, 20)
	for i := range allLines {
		allLines[i] = fmt.Sprintf("line %d", 6+i)
	}

	expandHunkGap(&file, gaps[0], expandAll, allLines)

	// Direct merge into 1 hunk covering all 30 lines
	if len(file.Hunks) != 1 {
		t.Fatalf("len(Hunks) = %d, want 1", len(file.Hunks))
	}

	if file.Hunks[0].NewStart != 1 || file.Hunks[0].NewCount != 30 || len(file.Hunks[0].Lines) != 30 {
		t.Fatalf("merged hunk = %+v, want NewStart: 1, NewCount: 30", file.Hunks[0])
	}
}

func TestDiffExpandBeforeFirstHunk(t *testing.T) {
	t.Parallel()

	// File where Hunk 0 starts at line 35 (gap before hunk: lines 1..34, count 34)
	file := changedFile{
		Path: "test.go",
		Hunks: []hunk{
			{OldStart: 35, OldCount: 5, NewStart: 35, NewCount: 5, Lines: makeLines(35, 39)},
		},
	}

	gaps := identifyGaps(file)
	if len(gaps) != 1 || !gaps[0].Before || gaps[0].Count != 34 {
		t.Fatalf("gaps = %+v, want 1 gap before first hunk of count 34", gaps)
	}

	// Expand up 20 lines (lines 15..34)
	upLines := make([]string, 20)
	for i := range upLines {
		upLines[i] = fmt.Sprintf("line %d", 15+i)
	}

	expandHunkGap(&file, gaps[0], expandUp, upLines)

	// Hunk 0 should now start at line 15 with 25 lines
	if file.Hunks[0].NewStart != 15 || file.Hunks[0].NewCount != 25 {
		t.Fatalf("hunk 0 = %+v, want NewStart: 15, NewCount: 25", file.Hunks[0])
	}

	// Remaining gap before hunk 0 is lines 1..14 (count 14)
	gaps2 := identifyGaps(file)
	if len(gaps2) != 1 || !gaps2[0].Before || gaps2[0].Count != 14 {
		t.Fatalf("gaps2 = %+v, want 1 gap of count 14", gaps2)
	}

	// Expand all remaining 14 lines (lines 1..14)
	allLines := make([]string, 14)
	for i := range allLines {
		allLines[i] = fmt.Sprintf("line %d", 1+i)
	}

	expandHunkGap(&file, gaps2[0], expandAll, allLines)

	// Hunk 0 should now start at line 1 with 39 lines
	if file.Hunks[0].NewStart != 1 || file.Hunks[0].NewCount != 39 {
		t.Fatalf("hunk 0 = %+v, want NewStart: 1, NewCount: 39", file.Hunks[0])
	}

	if g := identifyGaps(file); len(g) != 0 {
		t.Fatalf("expected 0 gaps, got %d", len(g))
	}
}

func TestDiffExpandAfterLastHunk(t *testing.T) {
	t.Parallel()

	// File where last hunk ends at line 70, total lines = 95 (gap: lines 71..95, count 25)
	file := changedFile{
		Path: "test.go",
		Hunks: []hunk{
			{OldStart: 66, OldCount: 5, NewStart: 66, NewCount: 5, Lines: makeLines(66, 70)},
		},
		TotalLines: 95,
	}

	gaps := identifyGaps(file)
	if len(gaps) != 2 {
		// Gap 0 is before first hunk (1..65), Gap 1 is after last hunk (71..95)
		t.Fatalf("len(gaps) = %d, want 2", len(gaps))
	}

	afterGap := gaps[1]
	if !afterGap.After || afterGap.NewStart != 71 || afterGap.NewEnd != 95 || afterGap.Count != 25 {
		t.Fatalf("afterGap = %+v, want lines 71..95 (count 25, After: true)", afterGap)
	}

	// Expand down 20 lines (lines 71..90)
	downLines := make([]string, 20)
	for i := range downLines {
		downLines[i] = fmt.Sprintf("line %d", 71+i)
	}

	expandHunkGap(&file, afterGap, expandDown, downLines)

	// Last hunk should now cover lines 66..90 (count 25)
	if file.Hunks[0].NewCount != 25 {
		t.Fatalf("last hunk NewCount = %d, want 25", file.Hunks[0].NewCount)
	}

	// Remaining gap after last hunk should be lines 91..95 (count 5)
	gaps2 := identifyGaps(file)
	if len(gaps2) != 2 || !gaps2[1].After || gaps2[1].Count != 5 {
		t.Fatalf("gaps2 = %+v, want after gap of count 5", gaps2)
	}

	// Expand all remaining 5 lines (lines 91..95)
	remLines := make([]string, 5)
	for i := range remLines {
		remLines[i] = fmt.Sprintf("line %d", 91+i)
	}

	expandHunkGap(&file, gaps2[1], expandAll, remLines)

	// Last hunk now covers lines 66..95 (count 30)
	if file.Hunks[0].NewCount != 30 {
		t.Fatalf("last hunk NewCount = %d, want 30", file.Hunks[0].NewCount)
	}

	// Verify after gap is gone
	gaps3 := identifyGaps(file)
	for _, g := range gaps3 {
		if g.After {
			t.Fatalf("unexpected after gap after full expansion: %+v", g)
		}
	}
}

func TestDiffExpandRowsUnifiedAndSplit(t *testing.T) {
	t.Parallel()

	file := changedFile{
		Path: "test.go",
		Hunks: []hunk{
			{OldStart: 1, OldCount: 3, NewStart: 1, NewCount: 3, Lines: makeLines(1, 3)},
			{OldStart: 15, OldCount: 3, NewStart: 15, NewCount: 3, Lines: makeLines(15, 17)},
		},
	}

	// 1. Unified rows
	uRows := diffRowsUnified(file, langGo)
	hasExpand := false
	for _, r := range uRows {
		if r.Kind == rowExpand {
			hasExpand = true
			if r.Expand == nil {
				t.Fatal("Expand metadata is nil in unified row")
			}
			if r.Expand.Count != 11 { // lines 4..14 (count 11)
				t.Fatalf("Expand count = %d, want 11", r.Expand.Count)
			}
			if !r.Expand.CanDown || !r.Expand.CanUp || !r.Expand.CanAll {
				t.Fatal("expected CanDown, CanUp, and CanAll to be true")
			}
		}
	}
	if !hasExpand {
		t.Fatal("diffRowsUnified did not include any expand rows")
	}

	// 2. Split rows
	sRows := diffRowsSplit(file, langGo)
	hasExpandSplit := false
	for _, r := range sRows {
		if r.Kind == rowExpand {
			hasExpandSplit = true
			if r.Expand == nil {
				t.Fatal("Expand metadata is nil in split row")
			}
			if r.Expand.Count != 11 {
				t.Fatalf("Split Expand count = %d, want 11", r.Expand.Count)
			}
		}
	}
	if !hasExpandSplit {
		t.Fatal("diffRowsSplit did not include any expand rows")
	}

	// 3. Wrap should preserve expand rows without alteration
	wrapped := wrapDiffRows(uRows, 80, 40, true)
	wrapExpandCount := 0
	for _, r := range wrapped {
		if r.Kind == rowExpand {
			wrapExpandCount++
		}
	}
	if wrapExpandCount != 1 {
		t.Fatalf("wrapped expand row count = %d, want 1", wrapExpandCount)
	}
}

func TestDiffExpandIntegration(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	// Switch to changes mode
	app.setMode(context.Background(), "changes")

	// Select docs/deployment.md which has a 2-line gap (lines 8..9) between hunks
	app.selectDiff("docs/deployment.md")

	if app.diffPath != "docs/deployment.md" {
		t.Fatalf("diffPath = %s, want docs/deployment.md", app.diffPath)
	}

	// Before expansion: verify gap exists
	file := app.currentDiffFile()
	if file == nil {
		t.Fatal("currentDiffFile is nil")
	}

	initialGaps := identifyGaps(*file)
	if len(initialGaps) != 1 {
		t.Fatalf("initialGaps = %d, want 1", len(initialGaps))
	}
	if initialGaps[0].NewStart != 8 || initialGaps[0].NewEnd != 9 {
		t.Fatalf("initial gap = %+v, want lines 8..9", initialGaps[0])
	}

	// Expand the gap via action click `diff-expand-all:0`
	err := app.onClick(context.Background(), ownframe.Box{Action: "diff-expand-all:0"})
	if err != nil {
		t.Fatalf("onClick failed: %v", err)
	}

	// After expansion: the gap should be gone, hunks merged into 1
	postGaps := identifyGaps(*file)
	if len(postGaps) != 0 {
		t.Fatalf("postGaps = %d, want 0 after expand all", len(postGaps))
	}

	// Verify rows contain the expanded context lines (lines 8 and 9)
	foundLine8 := false
	foundLine9 := false
	for _, r := range app.diffRows {
		if r.NewNo == "8" {
			foundLine8 = true
		}
		if r.NewNo == "9" {
			foundLine9 = true
		}
	}

	if !foundLine8 || !foundLine9 {
		t.Fatalf("expanded lines 8 and 9 not found in diffRows (line8: %v, line9: %v)", foundLine8, foundLine9)
	}
}

func TestDiffExpandDownAndUpIntegration(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	app.setMode(context.Background(), "changes")

	// config/services.yaml has gap 0: lines 1..6 (count 6) and gap 1: lines 16..21 (count 6)
	app.selectDiff("config/services.yaml")

	file := app.currentDiffFile()
	if file == nil {
		t.Fatal("currentDiffFile is nil")
	}

	// Click diff-expand-down:0 (expands lines 1..6)
	err := app.onClick(context.Background(), ownframe.Box{Action: "diff-expand-down:0"})
	if err != nil {
		t.Fatalf("onClick diff-expand-down:0 failed: %v", err)
	}

	// Gap 0 should now be expanded into lines 1..6
	foundLine1 := false
	for _, r := range app.diffRows {
		if r.NewNo == "1" {
			foundLine1 = true
			break
		}
	}
	if !foundLine1 {
		t.Fatal("line 1 not found after expanding down gap 0")
	}

	// Next, click diff-expand-up on remaining gap (now index 0)
	gapsAfter := identifyGaps(*file)
	if len(gapsAfter) != 1 {
		t.Fatalf("len(gapsAfter) = %d, want 1", len(gapsAfter))
	}

	err = app.onClick(context.Background(), ownframe.Box{Action: "diff-expand-up:0"})
	if err != nil {
		t.Fatalf("onClick diff-expand-up:0 failed: %v", err)
	}

	// Now all gaps should be expanded
	finalGaps := identifyGaps(*file)
	if len(finalGaps) != 0 {
		t.Fatalf("finalGaps = %d, want 0", len(finalGaps))
	}
}

func makeLines(start, end int) []diffLine {
	lines := make([]diffLine, end-start+1)
	for i := range lines {
		n := start + i
		lines[i] = diffLine{
			Kind:  ' ',
			OldNo: n,
			NewNo: n,
			Text:  fmt.Sprintf("line %d", n),
		}
	}
	return lines
}
