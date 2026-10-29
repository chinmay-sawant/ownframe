package main

import (
	"github.com/chinmay-sawant/ownframe"
)

const (
	// diffUnifiedText is the x offset of unified diff text inside the
	// code body: two 46 px number gutters plus the 20 px sign gutter.
	diffUnifiedText = 46 + 46 + 20
	// diffSplitLeftText is the x offset of the left cell text in split
	// diffs: one 44 px number gutter.
	diffSplitLeftText = 44
	// diffSplitRightText is the offset of the right cell text from the
	// middle of the row: one 44 px number gutter after the left half.
	diffSplitRightText = 44
)

// workingDiffFile returns the selected working-tree file when Source Control
// shows the working tree. Commit and branch diffs are read-only snapshots,
// so they have no editable target.
func (a *App) workingDiffFile() *changedFile {
	if a.mode != "changes" || a.diffSource != "working" {
		return nil
	}

	if a.diffPath == "" {
		return nil
	}

	return findFile(a.changes.Files, a.diffPath)
}

// canEditDiffFile reports whether one changed file can be opened for editing.
// Binary files and deleted paths have no editable text.
func canEditDiffFile(file *changedFile) bool {
	if file == nil || file.Binary || file.Status == "D" {
		return false
	}

	return true
}

// editDiffFile opens the selected working-tree file in the explorer and
// enters edit mode. A non-negative line is a zero-based source line; a
// negative line jumps to the first changed line, or the top when the file
// has no tracked changes.
func (a *App) editDiffFile(line int) {
	a.editDiffFileAt(line, 0)
}

// editDiffFileAt opens the selected working-tree file and places the caret
// at a display column, so a click in the diff lands on the same glyph in
// the source. A negative column keeps the caret at the line start.
func (a *App) editDiffFileAt(line, displayCol int) {
	if a.diffSource == "commit" {
		a.showToast("info", "Commit diff is read-only", "Open the working tree to edit this file.")
		a.refresh()

		return
	}

	if a.diffSource == "branch" {
		a.showToast("info", "Branch changes are read-only", "Open the working tree to edit this file.")
		a.refresh()

		return
	}

	file := a.workingDiffFile()
	if file == nil {
		a.showToast("info", "No file", "Select a working-tree file before editing.")
		a.refresh()

		return
	}

	if !canEditDiffFile(file) {
		if file.Binary {
			a.showToast("info", "Binary file", "The viewer reads text files only.")
		} else {
			a.showToast("info", "Deleted file", "This file was deleted in the working tree.")
		}
		a.refresh()

		return
	}

	path := file.Path
	a.openFile(path)

	if line < 0 {
		if lines := inlineChangeLines(a.changedLinesFor(path)); len(lines) > 0 {
			line = lines[0] - 1
		} else {
			line = 0
		}
	}

	if displayCol < 0 {
		displayCol = 0
	}

	a.beginEditAt(line, displayCol)
	a.centerCaretInView()
}

// centerCaretInView scrolls so the edit caret sits in the middle of the
// visible editor body. Redirects from the diff use this so the target line
// lands centered instead of at an edge. Typing keeps the minimal
// ensureEditCaretVisible behavior and does not call this.
func (a *App) centerCaretInView() {
	if a.edit == nil || !a.editing {
		return
	}

	line, col := a.editDisplayLineCol()
	row := line
	if a.wrap {
		if doc := a.groupDoc(0); doc != nil {
			row = caretVisualRow(doc.wrapped(a.codeColsForGroup(0)).Rows, line, col)
		}
	}

	capacity := a.editorClearHeight() / a.zx(rowH)
	if capacity < 1 {
		capacity = 1
	}

	a.scrollTo(row*a.zx(rowH) - (capacity-1)*a.zx(rowH)/halfDivisor)
}

// diffDisplayColForBox maps a press x position over one diff edit button to
// a display column in the new file. Wrapped continuation chunks record the
// display offset their text starts at, so a click on a wrapped tail lands on
// the right glyph instead of the line start. Split rows use the pressed
// side's origin; a padded empty side falls back to the side with text.
func (a *App) diffDisplayColForBox(box ownframe.Box, x float64) int {
	var row *diffRow
	for i := range a.view.Diff.Rows {
		if a.view.Diff.Rows[i].EditID == box.ID {
			row = &a.view.Diff.Rows[i]
			break
		}
	}

	if a.page == nil {
		return 0
	}

	codeBody, ok := findBoxID(a.page.Boxes(), "code-body")
	if !ok {
		return 0
	}

	pageX, _ := a.page.Scroll().ScrollOffset()
	x -= float64(pageX)

	scrollX := 0
	if !a.wrap {
		scrollX = a.scrollX
	}

	if !a.diffSplit {
		start := 0
		if row != nil {
			start = row.EditStart
		}
		textLeft := codeBody.X + float64(a.zx(diffUnifiedText))
		return max(0, int((x-textLeft+float64(scrollX))/a.zf(diffCellW)+roundHalf)+start)
	}

	w := codeBody.W
	if w < 1 {
		return 0
	}

	mid := codeBody.X + w/halfDivisor
	useLeft := x < mid
	if row != nil {
		if row.RightPad {
			useLeft = true
		} else if row.LeftPad {
			useLeft = false
		}
	}

	if useLeft {
		start := 0
		if row != nil {
			start = row.EditStartL
		}
		textLeft := codeBody.X + float64(a.zx(diffSplitLeftText))
		return max(0, int((x-textLeft+float64(scrollX))/a.zf(diffCellW)+roundHalf)+start)
	}

	start := 0
	if row != nil {
		start = row.EditStart
	}
	textLeft := codeBody.X + w/halfDivisor + float64(a.zx(diffSplitRightText))
	return max(0, int((x-textLeft+float64(scrollX))/a.zf(diffCellW)+roundHalf)+start)
}

// saveDirtyActive writes the active dirty file when the editor has no live
// buffer. Editing keeps a live buffer, but Escape folds it back into the
// document and clears the buffer while keeping the dirty flag. Source
// Control also shows a dirty working-tree file with no live buffer. In both
// cases Ctrl+S should save instead of doing nothing.
func (a *App) saveDirtyActive() {
	if a.edit != nil {
		a.saveEdit()

		return
	}

	path := ""
	if a.mode == "changes" && a.diffSource == "working" && a.dirty[a.diffPath] {
		path = a.diffPath
	} else if doc := a.activeDoc(); doc != nil && a.dirty[doc.Path] {
		path = doc.Path
	}

	if path == "" {
		return
	}

	doc := a.docs[path]
	if doc == nil {
		return
	}

	data := []byte(joinEditLines(doc.Lines))
	if err := a.ws.writeFile(path, data); err != nil {
		a.showToast("error", "Save failed", err.Error())
		a.refresh()

		return
	}

	a.dirty[path] = false
	a.docs[path] = a.ws.openDoc(path)
	a.showToast("success", "File saved", path)
	a.refresh()
}
