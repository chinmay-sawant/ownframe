package main

import "sort"

const (
	// inlineCardBottomStrip is the visible strip clampCardTop keeps above
	// the status bar, so a dragged card never slides under it.
	inlineCardBottomStrip = 40
)

// changedLinesFor returns the working-tree line changes for one path. Sample
// and live workspaces both use the parsed change set.
func (a *App) changedLinesFor(path string) map[int]inlineLine {
	file := findFile(a.changes.Files, path)
	if file == nil {
		return nil
	}

	return inlineChanges(*file)
}

// decorateCodeRows adds gutter marks and a subtle line tint to source rows.
func (a *App) decorateCodeRows(doc *fileDoc, rows []codeRow, group int) {
	if doc == nil || len(rows) == 0 {
		return
	}

	changes := a.changedLinesFor(doc.Path)
	if len(changes) == 0 {
		return
	}

	marks := gutterMarks(changes)
	lines := inlineChangeLines(changes)
	scroll := a.scrollY
	if a.split {
		scroll = a.scrollGroup[group]
	}
	current := scroll / a.zx(rowH)
	if a.changeNav.valid && a.changeNav.path == doc.Path && a.changeNav.mode == a.mode &&
		a.changeNav.group == group && a.changeNav.scroll == scroll {
		current = a.changeNav.row
	}
	selected := lines[0]
	for i, line := range lines {
		if i > 0 && line == lines[i-1]+1 {
			continue
		}
		visual := line - 1
		if a.wrap {
			visual = doc.wrapped(a.codeColsForGroup(group)).firstRow(visual)
		}
		if visual > current {
			break
		}
		selected = line
	}
	selectedEnd := selected
	for _, line := range lines {
		if line > selectedEnd+1 {
			break
		}
		if line == selectedEnd+1 {
			selectedEnd = line
		}
	}

	for i := range rows {
		row := &rows[i]
		line := row.Line + 1

		change, ok := inlineChangeAt(changes, line)
		if !ok {
			continue
		}

		row.ChangeLine = line
		row.ChangeClass = inlineRowClass(change.Kind)
		row.ChangeCurrent = line >= selected && line <= selectedEnd

		if row.No == 0 {
			continue
		}

		if mark, markOK := marks[line]; markOK {
			row.Mark = mark.Glyph
			row.MarkClass = mark.Class
			row.Hint = mark.Hint
		}
	}
}

// inlineRowClass is the row background class for one change kind.
func inlineRowClass(kind inlineKind) string {
	switch kind {
	case inlineAdded:
		return "line-add"
	case inlineDeleted:
		return "line-del"
	case inlineModified:
		return "line-mod"
	default:
		return ""
	}
}

// showInlineCard opens the line card for one line in the left editor. Changed
// lines show their old and new sides; any other line shows the current text
// under the Line kind. When the line sits inside a multi-line edit selection,
// the card shows the whole selection instead of just that line.
func (a *App) showInlineCard(line int) {
	doc := a.groupDoc(0)
	if doc == nil || line < 1 {
		return
	}

	if line > len(doc.Lines) {
		return
	}

	if start, end, ok := a.selectionRange(doc); ok && line >= start && line <= end {
		a.showSelectionCard(doc, start, end)
		return
	}

	changes := a.changedLinesFor(doc.Path)
	change, ok := inlineChangeAt(changes, line)
	if !ok {
		change = inlineLine{Kind: inlineNone}
	}

	lines := inlineChangeLines(changes)
	index := -1

	for i, changed := range lines {
		if changed == line {
			index = i
			break
		}
	}

	hasPrev, hasNext := inlineCardBounds(lines, line, index)

	visual := line - 1
	if a.wrap {
		plan := doc.wrapped(a.codeColsForGroup(0))
		visual = plan.firstRow(line - 1)
	}

	rowTop := a.zx(editorBodyTop) + (visual-a.scrollY/a.zx(rowH))*a.zx(rowH)
	newText := ""
	if line-1 < len(doc.Lines) {
		newText = doc.Lines[line-1]
	}

	card := inlineCardFor(rowTop, a.pageWidth(), a.viewportH, line, change, doc.Path, newText)
	card.Line = line
	card.Left = a.chromeLeftPx() + inlineCardLeft
	card.HasPrev = hasPrev
	card.HasNext = hasNext

	a.inlineCard = card
	a.refresh()
}

// selectionRange returns the 1-based inclusive source lines of the edit
// selection when it spans more than one line. A collapsed caret reports false
// so the caller keeps the single-line card.
func (a *App) selectionRange(doc *fileDoc) (int, int, bool) {
	if doc == nil || !a.editing || a.edit == nil || a.edit.Path != doc.Path {
		return 0, 0, false
	}
	text, start, end := a.editSelection()
	if start == end {
		return 0, 0, false
	}
	startLine, _ := offsetLineCol(text, start)
	endLine, _ := offsetLineCol(text, end)
	startLine++
	endLine++
	if endLine <= startLine {
		return 0, 0, false
	}
	if startLine < 1 {
		startLine = 1
	}
	if endLine > len(doc.Lines) {
		endLine = len(doc.Lines)
	}
	if endLine <= startLine {
		return 0, 0, false
	}
	return startLine, endLine, true
}

// showSelectionCard opens the card for a multi-line selection, with one row
// per selected source line. Prev/next still step through the file's changed
// lines from the selection start.
func (a *App) showSelectionCard(doc *fileDoc, start, end int) {
	if doc == nil || start < 1 || end < start || end > len(doc.Lines) {
		return
	}
	visual := start - 1
	if a.wrap {
		plan := doc.wrapped(a.codeColsForGroup(0))
		visual = plan.firstRow(start - 1)
	}
	rowTop := a.zx(editorBodyTop) + (visual-a.scrollY/a.zx(rowH))*a.zx(rowH)
	lines := append([]string(nil), doc.Lines[start-1:end]...)
	card := inlineCardForSelection(rowTop, a.pageWidth(), a.viewportH, start, end, doc.Path, lines)
	card.Left = a.chromeLeftPx() + inlineCardLeft
	changes := a.changedLinesFor(doc.Path)
	changed := inlineChangeLines(changes)
	index := -1
	for i, l := range changed {
		if l == start {
			index = i
			break
		}
	}
	card.HasPrev, card.HasNext = inlineCardBounds(changed, start, index)
	a.inlineCard = card
	a.refresh()
}

// refreshInlineCardSelection rebuilds an open card's selection rows in place,
// keeping the dragged position. It returns false when there is no selection
// card to refresh.
func (a *App) refreshInlineCardSelection() bool {
	if !a.inlineCard.Show || a.inlineCard.Kind != cardKindSelected {
		return false
	}
	doc := a.groupDoc(0)
	if doc == nil {
		return false
	}
	start, end, ok := a.selectionRange(doc)
	if !ok {
		return false
	}
	left, top := a.inlineCard.Left, a.inlineCard.Top
	offX, offY := a.inlineCard.OffX, a.inlineCard.OffY
	visual := start - 1
	if a.wrap {
		plan := doc.wrapped(a.codeColsForGroup(0))
		visual = plan.firstRow(start - 1)
	}
	_ = visual
	rowTop := a.zx(editorBodyTop) + (visual-a.scrollY/a.zx(rowH))*a.zx(rowH)
	lines := append([]string(nil), doc.Lines[start-1:end]...)
	card := inlineCardForSelection(rowTop, a.pageWidth(), a.viewportH, start, end, doc.Path, lines)
	card.Left = left
	card.Top = a.clampCardTop(top)
	card.OffX = offX
	card.OffY = offY
	changes := a.changedLinesFor(doc.Path)
	changed := inlineChangeLines(changes)
	index := -1
	for i, l := range changed {
		if l == start {
			index = i
			break
		}
	}
	card.HasPrev, card.HasNext = inlineCardBounds(changed, start, index)
	a.inlineCard = card
	return true
}

// cardAnchor recomputes the card's attached position from its line: the
// editor's left edge plus the gap, and the clamped top above the line's row.
// It reports false when the card has no line to attach to.
func (a *App) cardAnchor() (int, int, bool) {
	if !a.inlineCard.Show {
		return 0, 0, false
	}
	doc := a.groupDoc(0)
	if doc == nil || len(doc.Lines) == 0 {
		return 0, 0, false
	}
	line := a.inlineCard.Line
	if line < 1 || line > len(doc.Lines) {
		return 0, 0, false
	}
	if a.inlineCard.Kind == cardKindSelected {
		if a.inlineCard.EndLine < line {
			return 0, 0, false
		}
	}
	left := a.chromeLeftPx() + inlineCardLeft
	visual := line - 1
	if a.wrap {
		plan := doc.wrapped(a.codeColsForGroup(0))
		visual = plan.firstRow(line - 1)
	}
	rowTop := a.zx(editorBodyTop) + (visual-a.scrollY/a.zx(rowH))*a.zx(rowH)
	return left, inlineCardTop(rowTop, a.viewportH), true
}

// repositionInlineCard keeps an open card on its line across scrolls, zoom,
// and wrap changes, preserving the user's drag offset. The card only closes
// when its file or line is gone.
func (a *App) repositionInlineCard() {
	if !a.inlineCard.Show {
		return
	}
	left, top, ok := a.cardAnchor()
	if !ok {
		a.inlineCard = inlineCard{}
		return
	}
	a.inlineCard.Left = a.clampCardLeft(left + a.inlineCard.OffX)
	a.inlineCard.Top = a.clampCardTop(top + a.inlineCard.OffY)
}

// stampCardOffset records the drag offset between the card's current position
// and its line anchor, so later scrolls keep the card where the user put it.
func (a *App) stampCardOffset() {
	left, top, ok := a.cardAnchor()
	if !ok {
		a.inlineCard.OffX = 0
		a.inlineCard.OffY = 0
		return
	}
	a.inlineCard.OffX = a.inlineCard.Left - left
	a.inlineCard.OffY = a.inlineCard.Top - top
}

// clampCardLeft keeps a dragged card inside the page width.
func (a *App) clampCardLeft(left int) int {
	width := a.inlineCard.Width
	if width < 1 {
		width = inlineCardWidth
	}
	maxLeft := a.pageWidth() - width
	if maxLeft < 0 {
		maxLeft = 0
	}
	if left < 0 {
		return 0
	}
	if left > maxLeft {
		return maxLeft
	}
	return left
}

// clampCardTop keeps a dragged card between the page top and a visible strip
// above the status bar.
func (a *App) clampCardTop(top int) int {
	if top < inlineCardGap {
		return inlineCardGap
	}
	maxTop := a.viewportH - inlineCardBottomStrip
	if maxTop < inlineCardGap {
		maxTop = inlineCardGap
	}
	if top > maxTop {
		return maxTop
	}
	return top
}

// inlineCardBounds reports whether a card at line can step to a previous or
// next changed line. index is the position of line inside the sorted changed
// lines, or -1 when line itself is unchanged.
func inlineCardBounds(lines []int, line, index int) (bool, bool) {
	if index >= 0 {
		return index > 0, index < len(lines)-1
	}

	hasPrev := false
	hasNext := false

	for _, changed := range lines {
		if changed < line {
			hasPrev = true
		} else if changed > line {
			hasNext = true

			break
		}
	}

	return hasPrev, hasNext
}

// currentSourceLine is the 1-based source line Ctrl+I names in the left
// editor. While editing it is the caret line; otherwise it is the top visible
// source line, so line 100 on screen names line 100.
func (a *App) currentSourceLine() int {
	doc := a.groupDoc(0)
	if doc == nil || len(doc.Lines) == 0 {
		return 0
	}

	if a.editing && a.edit != nil && a.edit.Path == doc.Path {
		line, _ := offsetLineCol(a.edit.Text(), a.editCaret)
		line++

		return min(max(line, 1), len(doc.Lines))
	}

	scroll := a.scrollY
	if a.split {
		scroll = a.scrollGroup[0]
	}

	visual := scroll / a.zx(rowH)

	if a.wrap {
		plan := doc.wrapped(a.codeColsForGroup(0))
		if len(plan.Rows) == 0 {
			return 0
		}

		if visual < 0 {
			visual = 0
		}

		if visual >= len(plan.Rows) {
			visual = len(plan.Rows) - 1
		}

		line := plan.Rows[visual].Line + 1

		return min(max(line, 1), len(doc.Lines))
	}

	return min(max(visual+1, 1), len(doc.Lines))
}

// showInlineCardAtCurrent opens the line card at the current source line.
func (a *App) showInlineCardAtCurrent() {
	if line := a.currentSourceLine(); line > 0 {
		a.showInlineCard(line)
	}
}

// inlineCardStep moves the line card to the previous or next changed line in
// the same file. From an unchanged line it jumps to the nearest changed line
// in that direction. It stops at the first and last change.
func (a *App) inlineCardStep(delta int) {
	doc := a.groupDoc(0)
	if doc == nil || !a.inlineCard.Show {
		return
	}

	lines := inlineChangeLines(a.changedLinesFor(doc.Path))
	if len(lines) == 0 {
		return
	}

	line := a.inlineCard.Line
	index := -1

	for i, changed := range lines {
		if line == changed {
			index = i
			break
		}
	}

	if index < 0 {
		if delta > 0 {
			for _, changed := range lines {
				if changed > line {
					a.showInlineCard(changed)

					return
				}
			}
		} else {
			for i := len(lines) - 1; i >= 0; i-- {
				if lines[i] < line {
					a.showInlineCard(lines[i])

					return
				}
			}
		}

		return
	}

	next := index + delta
	if next < 0 || next >= len(lines) {
		return
	}

	a.showInlineCard(lines[next])
}

// inlineChangeLines sorts the changed 1-based lines of one file.
func inlineChangeLines(changes map[int]inlineLine) []int {
	lines := make([]int, 0, len(changes))

	for line := range changes {
		lines = append(lines, line)
	}

	sort.Ints(lines)

	return lines
}

// closeInlineCard hides the line card.
func (a *App) closeInlineCard() {
	if !a.inlineCard.Show {
		return
	}

	a.inlineCard = inlineCard{}
	a.refresh()
}
