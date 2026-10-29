package main

import "sort"

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

// showInlineCard opens the line card for a changed line in the left editor.
func (a *App) showInlineCard(line int) {
	doc := a.groupDoc(0)
	if doc == nil || line < 1 {
		return
	}

	changes := a.changedLinesFor(doc.Path)
	change, ok := inlineChangeAt(changes, line)
	if !ok {
		return
	}

	lines := inlineChangeLines(changes)
	index := -1

	for i, changed := range lines {
		if changed == line {
			index = i
			break
		}
	}

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
	card.Left = a.zx(chromeLeft) + inlineCardLeft
	card.HasPrev = index > 0
	card.HasNext = index >= 0 && index < len(lines)-1

	a.inlineCard = card
	a.refresh()
}

// inlineCardStep moves the line card to the previous or next changed line in
// the same file. It stops at the first and last change.
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
