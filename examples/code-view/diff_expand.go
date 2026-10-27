package main

import (
	"fmt"
	"strings"
)

const expandStep = 20

// expandDir specifies which direction an accordion expander grows context lines.
type expandDir int

const (
	expandDown expandDir = iota
	expandUp
	expandAll
)

// diffGap represents an unexpanded context gap of unchanged lines between hunks
// or at file boundaries (before the first hunk or after the last hunk).
type diffGap struct {
	Index    int
	OldStart int
	OldEnd   int
	NewStart int
	NewEnd   int
	Count    int
	Before   bool
	After    bool
	HunkIdx  int
}

// CanDown reports whether context lines can be expanded downward from the gap top.
func (g diffGap) CanDown() bool {
	return g.Count > 0
}

// CanUp reports whether context lines can be expanded upward toward the gap bottom.
func (g diffGap) CanUp() bool {
	return g.Count > 0
}

// Label returns a summary string for the gap accordion.
func (g diffGap) Label() string {
	if g.Count == 1 {
		return "··· 1 line ···"
	}

	return fmt.Sprintf("··· %d lines ···", g.Count)
}

// diffExpandRow holds presentation metadata for an accordion expander row in the diff.
type diffExpandRow struct {
	Index    int
	OldStart int
	OldEnd   int
	NewStart int
	NewEnd   int
	Count    int
	CanUp    bool
	CanDown  bool
	CanAll   bool
	Label    string
	DownText string
	UpText   string
	AllText  string
}

// newDiffExpandRow constructs a diffRow with Kind "expand" for the given gap.
func newDiffExpandRow(gap diffGap) diffRow {
	downText := fmt.Sprintf("Expand %d lines down ↓", min(expandStep, gap.Count))
	upText := fmt.Sprintf("Expand %d lines up ↑", min(expandStep, gap.Count))
	allText := fmt.Sprintf("Expand all lines (%d lines)", gap.Count)

	return diffRow{
		Kind: rowExpand,
		Hunk: gap.Label(),
		Expand: &diffExpandRow{
			Index:    gap.Index,
			OldStart: gap.OldStart,
			OldEnd:   gap.OldEnd,
			NewStart: gap.NewStart,
			NewEnd:   gap.NewEnd,
			Count:    gap.Count,
			CanUp:    gap.CanUp(),
			CanDown:  gap.CanDown(),
			CanAll:   true,
			Label:    gap.Label(),
			DownText: downText,
			UpText:   upText,
			AllText:  allText,
		},
	}
}

// identifyGaps scans a changed file for omitted unchanged lines before the first hunk,
// between consecutive hunks, and after the last hunk (when totalLines is known).
func identifyGaps(file changedFile) []diffGap {
	return identifyGapsWithTotal(file, file.TotalLines)
}

// identifyGapsWithTotal finds gaps given an explicit total line count for the new file.
func identifyGapsWithTotal(file changedFile, totalLines int) []diffGap {
	if file.Binary || len(file.Hunks) == 0 {
		return nil
	}

	var gaps []diffGap
	idx := 0

	// 1. Gap before the first hunk:
	first := file.Hunks[0]
	if first.NewStart > 1 && first.OldStart > 1 {
		count := first.NewStart - 1
		gaps = append(gaps, diffGap{
			Index:    idx,
			OldStart: 1,
			OldEnd:   first.OldStart - 1,
			NewStart: 1,
			NewEnd:   first.NewStart - 1,
			Count:    count,
			Before:   true,
			HunkIdx:  0,
		})
		idx++
	}

	// 2. Gaps between consecutive hunks:
	for i := 0; i < len(file.Hunks)-1; i++ {
		cur := file.Hunks[i]
		next := file.Hunks[i+1]

		curEndNew := cur.NewStart + cur.NewCount
		curEndOld := cur.OldStart + cur.OldCount
		nextStartNew := next.NewStart
		nextStartOld := next.OldStart

		if nextStartNew > curEndNew {
			count := nextStartNew - curEndNew
			gaps = append(gaps, diffGap{
				Index:    idx,
				OldStart: curEndOld,
				OldEnd:   nextStartOld - 1,
				NewStart: curEndNew,
				NewEnd:   nextStartNew - 1,
				Count:    count,
				HunkIdx:  i,
			})
			idx++
		}
	}

	// 3. Gap after the last hunk:
	if totalLines > 0 {
		last := file.Hunks[len(file.Hunks)-1]
		lastEndNew := last.NewStart + last.NewCount
		lastEndOld := last.OldStart + last.OldCount

		if totalLines >= lastEndNew {
			count := totalLines - lastEndNew + 1
			gaps = append(gaps, diffGap{
				Index:    idx,
				OldStart: lastEndOld,
				OldEnd:   lastEndOld + count - 1,
				NewStart: lastEndNew,
				NewEnd:   totalLines,
				Count:    count,
				After:    true,
				HunkIdx:  len(file.Hunks) - 1,
			})
			idx++
		}
	}

	return gaps
}

// expandHunkGap expands unchanged context lines into the hunks of a changedFile.
func expandHunkGap(file *changedFile, gap diffGap, dir expandDir, lines []string) {
	if len(lines) == 0 {
		return
	}

	take := len(lines)
	var startLine, endLine, oldStartLine, oldEndLine int

	switch dir {
	case expandDown:
		startLine = gap.NewStart
		endLine = gap.NewStart + take - 1
		oldStartLine = gap.OldStart
		oldEndLine = gap.OldStart + take - 1
	case expandUp:
		startLine = gap.NewEnd - take + 1
		endLine = gap.NewEnd
		oldStartLine = gap.OldEnd - take + 1
		oldEndLine = gap.OldEnd
	case expandAll:
		startLine = gap.NewStart
		endLine = gap.NewEnd
		oldStartLine = gap.OldStart
		oldEndLine = gap.OldEnd
	}

	diffLines := make([]diffLine, len(lines))
	for i, l := range lines {
		diffLines[i] = diffLine{
			Kind:  ' ',
			OldNo: oldStartLine + i,
			NewNo: startLine + i,
			Text:  expandTabs(l),
		}
	}

	if gap.Before {
		if len(file.Hunks) == 0 {
			file.Hunks = []hunk{
				{
					OldStart: oldStartLine,
					OldCount: len(diffLines),
					NewStart: startLine,
					NewCount: len(diffLines),
					Lines:    diffLines,
					oldNext:  oldEndLine + 1,
					newNext:  endLine + 1,
				},
			}

			return
		}

		if endLine+1 == file.Hunks[0].NewStart {
			file.Hunks[0].Lines = append(diffLines, file.Hunks[0].Lines...)
			file.Hunks[0].NewStart = startLine
			file.Hunks[0].OldStart = oldStartLine
			file.Hunks[0].NewCount += len(diffLines)
			file.Hunks[0].OldCount += len(diffLines)
		} else {
			newH := hunk{
				OldStart: oldStartLine,
				OldCount: len(diffLines),
				NewStart: startLine,
				NewCount: len(diffLines),
				Lines:    diffLines,
				oldNext:  oldEndLine + 1,
				newNext:  endLine + 1,
			}
			file.Hunks = append([]hunk{newH}, file.Hunks...)
		}

		return
	}

	if gap.After {
		lastIdx := len(file.Hunks) - 1
		if startLine == file.Hunks[lastIdx].NewStart+file.Hunks[lastIdx].NewCount {
			file.Hunks[lastIdx].Lines = append(file.Hunks[lastIdx].Lines, diffLines...)
			file.Hunks[lastIdx].NewCount += len(diffLines)
			file.Hunks[lastIdx].OldCount += len(diffLines)
			file.Hunks[lastIdx].newNext = endLine + 1
			file.Hunks[lastIdx].oldNext = oldEndLine + 1
		} else {
			newH := hunk{
				OldStart: oldStartLine,
				OldCount: len(diffLines),
				NewStart: startLine,
				NewCount: len(diffLines),
				Lines:    diffLines,
				oldNext:  oldEndLine + 1,
				newNext:  endLine + 1,
			}
			file.Hunks = append(file.Hunks, newH)
		}

		return
	}

	hIdx := gap.HunkIdx
	if hIdx < 0 || hIdx+1 >= len(file.Hunks) {
		return
	}

	switch dir {
	case expandDown:
		file.Hunks[hIdx].Lines = append(file.Hunks[hIdx].Lines, diffLines...)
		file.Hunks[hIdx].NewCount += len(diffLines)
		file.Hunks[hIdx].OldCount += len(diffLines)
		file.Hunks[hIdx].newNext = endLine + 1
		file.Hunks[hIdx].oldNext = oldEndLine + 1

		next := file.Hunks[hIdx+1]
		if file.Hunks[hIdx].NewStart+file.Hunks[hIdx].NewCount == next.NewStart {
			file.Hunks[hIdx].Lines = append(file.Hunks[hIdx].Lines, next.Lines...)
			file.Hunks[hIdx].NewCount += next.NewCount
			file.Hunks[hIdx].OldCount += next.OldCount
			file.Hunks[hIdx].newNext = next.newNext
			file.Hunks[hIdx].oldNext = next.oldNext
			file.Hunks = append(file.Hunks[:hIdx+1], file.Hunks[hIdx+2:]...)
		}

	case expandUp:
		file.Hunks[hIdx+1].Lines = append(diffLines, file.Hunks[hIdx+1].Lines...)
		file.Hunks[hIdx+1].NewStart = startLine
		file.Hunks[hIdx+1].OldStart = oldStartLine
		file.Hunks[hIdx+1].NewCount += len(diffLines)
		file.Hunks[hIdx+1].OldCount += len(diffLines)

		if file.Hunks[hIdx].NewStart+file.Hunks[hIdx].NewCount == file.Hunks[hIdx+1].NewStart {
			next := file.Hunks[hIdx+1]
			file.Hunks[hIdx].Lines = append(file.Hunks[hIdx].Lines, next.Lines...)
			file.Hunks[hIdx].NewCount += next.NewCount
			file.Hunks[hIdx].OldCount += next.OldCount
			file.Hunks[hIdx].newNext = next.newNext
			file.Hunks[hIdx].oldNext = next.oldNext
			file.Hunks = append(file.Hunks[:hIdx+1], file.Hunks[hIdx+2:]...)
		}

	case expandAll:
		next := file.Hunks[hIdx+1]
		file.Hunks[hIdx].Lines = append(file.Hunks[hIdx].Lines, diffLines...)
		file.Hunks[hIdx].Lines = append(file.Hunks[hIdx].Lines, next.Lines...)
		file.Hunks[hIdx].NewCount += len(diffLines) + next.NewCount
		file.Hunks[hIdx].OldCount += len(diffLines) + next.OldCount
		file.Hunks[hIdx].newNext = next.newNext
		file.Hunks[hIdx].oldNext = next.oldNext
		file.Hunks = append(file.Hunks[:hIdx+1], file.Hunks[hIdx+2:]...)
	}
}

// currentDiffFile returns a pointer to the active file whose diff is displayed.
func (a *App) currentDiffFile() *changedFile {
	if a.diffSource == "commit" {
		return a.commitDetails[commitDetailKey(a.commitOpen, a.commitFile)]
	}

	return findFile(a.changes.Files, a.diffPath)
}

// fileTotalLines reads or calculates the line count of a workspace file.
func (a *App) fileTotalLines(path string) int {
	if a.ws == nil || path == "" {
		return 0
	}

	if doc := a.docs[path]; doc != nil && len(doc.Lines) > 0 {
		return len(doc.Lines)
	}

	raw, err := a.ws.ReadFile(path)
	if err != nil {
		return 0
	}

	text := string(raw)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")

	if text == "" {
		return 0
	}

	return strings.Count(text, "\n") + 1
}

// loadContextLines extracts 1-based lines from startLine to endLine from the file.
func (a *App) loadContextLines(path string, startLine, endLine int) ([]string, error) {
	if a.ws == nil {
		return nil, fmt.Errorf("no workspace available to read %s", path)
	}

	raw, err := a.ws.ReadFile(path)
	if err != nil {
		return nil, err
	}

	text := string(raw)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")

	var allLines []string
	if text != "" {
		allLines = strings.Split(text, "\n")
	}

	var lines []string
	for lineNo := startLine; lineNo <= endLine; lineNo++ {
		idx := lineNo - 1
		if idx >= 0 && idx < len(allLines) {
			lines = append(lines, allLines[idx])
		} else {
			lines = append(lines, "")
		}
	}

	return lines, nil
}

// expandDiffGap coordinates expanding unchanged lines for the given gap index.
func (a *App) expandDiffGap(idx int, dir expandDir) {
	file := a.currentDiffFile()
	if file == nil {
		return
	}

	if file.TotalLines == 0 {
		file.TotalLines = a.fileTotalLines(file.Path)
	}

	gaps := identifyGaps(*file)
	if idx < 0 || idx >= len(gaps) {
		return
	}

	gap := gaps[idx]
	take := min(expandStep, gap.Count)
	if dir == expandAll {
		take = gap.Count
	}

	var startLine, endLine int
	switch dir {
	case expandDown, expandAll:
		startLine = gap.NewStart
		endLine = gap.NewStart + take - 1
	case expandUp:
		startLine = gap.NewEnd - take + 1
		endLine = gap.NewEnd
	}

	lines, err := a.loadContextLines(file.Path, startLine, endLine)
	if err != nil || len(lines) == 0 {
		return
	}

	expandHunkGap(file, gap, dir, lines)

	a.rebuildDiffRows()
	a.refresh()
}

// diffExpandUp handles the `diff-expand-up:<idx>` action.
func (a *App) diffExpandUp(idx int) {
	a.expandDiffGap(idx, expandUp)
}

// diffExpandDown handles the `diff-expand-down:<idx>` action.
func (a *App) diffExpandDown(idx int) {
	a.expandDiffGap(idx, expandDown)
}

// diffExpandAll handles the `diff-expand-all:<idx>` action.
func (a *App) diffExpandAll(idx int) {
	a.expandDiffGap(idx, expandAll)
}
