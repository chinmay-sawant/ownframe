package main

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
func (a *App) decorateCodeRows(doc *fileDoc, rows []codeRow) {
	if doc == nil || len(rows) == 0 {
		return
	}

	changes := a.changedLinesFor(doc.Path)
	if len(changes) == 0 {
		return
	}

	marks := gutterMarks(changes)

	for i := range rows {
		row := &rows[i]
		line := row.Line + 1

		change, ok := inlineChangeAt(changes, line)
		if !ok {
			continue
		}

		row.ChangeLine = line
		row.ChangeClass = inlineRowClass(change.Kind)

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

	visual := line - 1
	if a.wrap {
		plan := doc.wrapped(a.codeColsForGroup(0))
		visual = plan.firstRow(line - 1)
	}

	rowTop := editorBodyTop + (visual-a.scrollY/rowH)*rowH
	newText := ""
	if line-1 < len(doc.Lines) {
		newText = doc.Lines[line-1]
	}

	card := inlineCardFor(rowTop, a.pageWidth(), a.viewportH, line, change, doc.Path, newText)
	card.Left = chromeLeft + inlineCardLeft

	a.inlineCard = card
	a.refresh()
}

// closeInlineCard hides the line card.
func (a *App) closeInlineCard() {
	if !a.inlineCard.Show {
		return
	}

	a.inlineCard = inlineCard{}
	a.refresh()
}
