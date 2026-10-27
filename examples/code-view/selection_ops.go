package main

import (
	"context"
	"unicode/utf8"

	"github.com/chinmay-sawant/ownframe"
)

const editTextLeft = 16 + 44

// onPress records the selection anchor for a drag inside the edit view.
func (a *App) onPress(ctx context.Context, box ownframe.Box, x, y float64) (bool, error) {
	if !a.editing || a.edit == nil || box.ID != "code-body" {
		return false, nil
	}

	line, col, ok := a.editPointAt(x, y)
	if !ok {
		return false, nil
	}

	a.selX, a.selY = x, y
	text := a.edit.Text()
	sel := selectionFromPoints(text, line, col, line, col)
	a.editAnchor = sel.Anchor
	a.editCaret = sel.Caret
	a.dragSelect = true
	a.refresh()

	return false, a.page.Redraw(ctx)
}

// onDragStart claims a drag in the edit body.
func (a *App) onDragClaim(box ownframe.Box) bool {
	return a.editing && box.ID == "code-body"
}

// extendEditSelection extends the caret from the drag start to x,y.
func (a *App) extendEditSelection(x, y float64) {
	line, col, ok := a.editPointAt(x, y)
	if !ok || a.edit == nil {
		return
	}

	text := a.edit.Text()
	anchorLine, anchorCol := offsetLineCol(text, a.editAnchor)
	sel := selectionFromPoints(text, anchorLine, anchorCol, line, col)
	a.editAnchor = sel.Anchor
	a.editCaret = sel.Caret
}

// editPointAt maps a page point in the code body to a source line and raw
// rune column.
func (a *App) editPointAt(x, y float64) (int, int, bool) {
	doc := a.groupDoc(0)
	if doc == nil || a.edit == nil {
		return 0, 0, false
	}

	box, ok := findBoxID(a.page.Boxes(), "code-body")
	if !ok {
		return 0, 0, false
	}

	cols := a.codeColsForGroup(0)
	visual := visualRowAt(int(y), int(box.Y), a.scrollY, rowH)
	line := sourceLineAt(doc, cols, visual)
	plan := doc.wrapped(cols)

	rowStart := 0
	if a.wrap && visual >= 0 && visual < len(plan.Rows) {
		line = plan.Rows[visual].Line
		rowStart = plan.Rows[visual].Start
	}

	textLeft := box.X + editTextLeft
	displayCol := rowStart + int((x-textLeft)/codeCellW+roundHalf)
	length := 0
	if line >= 0 && line < len(doc.Lines) {
		length = utf8.RuneCountInString(expandTabs(doc.Lines[line]))
	}

	displayCol = max(rowStart, min(displayCol, length))
	raw := utf8.RuneCountInString(expandTabs(lineTextAt(a.edit.Text(), line)))
	rawCol := rawColumnAt(lineTextAt(a.edit.Text(), line), displayCol)
	rawCol = max(0, min(rawCol, raw))

	return line, rawCol, true
}

// rawColumnAt turns a display column, with tabs expanded, back into a raw rune
// column in the buffer line.
func rawColumnAt(line string, displayCol int) int {
	if displayCol <= 0 {
		return 0
	}

	cells := 0
	for i, r := range []rune(line) {
		width := 1
		if r == '\t' {
			width = len(tabSpaces)
		}

		if cells+width > displayCol {
			return i
		}

		cells += width
		if cells == displayCol {
			return i + 1
		}
	}

	return utf8.RuneCountInString(line)
}
