package main

import "strings"

const paneMinWidth = 220

// activeGroup returns the group that receives file opens and shortcuts.
func (a *App) activeGroup() int {
	if !a.split || a.focusGroup != 1 {
		return 0
	}

	return 1
}

// groupTabs returns one group's tab paths.
func (a *App) groupTabs(group int) []string {
	if group == 1 {
		return a.tabs2
	}

	return a.tabs
}

// groupActive returns one group's active tab index.
func (a *App) groupActive(group int) int {
	if group == 1 {
		return a.active2
	}

	return a.active
}

// groupDoc returns the document behind one group's active tab.
func (a *App) groupDoc(group int) *fileDoc {
	tabs := a.groupTabs(group)
	active := a.groupActive(group)
	if active < 0 || active >= len(tabs) {
		return nil
	}

	return a.docs[tabs[active]]
}

// groupPaths returns every path open in either group.
func (a *App) groupPaths() []string {
	out := make([]string, 0, len(a.tabs)+len(a.tabs2))
	out = append(out, a.tabs...)
	out = append(out, a.tabs2...)

	return out
}

// paneGeometry returns the left edge and width of the first pane.
func (a *App) paneGeometry() (int, int) {
	total := a.editorWidth()
	if !a.split || a.mode != "explorer" {
		return chromeLeft, total
	}

	left := int(float64(total) * a.splitRatio)
	if left < paneMinWidth {
		left = paneMinWidth
	}

	if left > total-paneMinWidth {
		left = total - paneMinWidth
	}

	if left < 1 {
		left = total
	}

	return chromeLeft, left
}

// secondGeometry returns the left edge and width of the second pane.
func (a *App) secondGeometry() (int, int) {
	left, width := a.paneGeometry()
	total := a.editorWidth()

	secondLeft := left + width + 1
	secondWidth := total - width - 1
	if secondWidth < 1 {
		secondWidth = 1
	}

	return secondLeft, secondWidth
}

// codeColsFor is the monospace cell count inside a pane width.
func (a *App) codeColsFor(width int) int {
	avail := width - codeGutter - codePadRight - editorBorderPad
	cols := int(float64(avail) / codeCellW)
	if cols < wrapMinCols {
		cols = wrapMinCols
	}

	return cols
}

// codeColsForGroup returns the cell count for one editor group.
func (a *App) codeColsForGroup(group int) int {
	width := a.editorWidth()
	if group == 1 {
		_, width = a.secondGeometry()
	} else if a.split {
		_, width = a.paneGeometry()
	}

	return a.codeColsFor(width)
}

// previewCols is the column budget for rendered Markdown.
func (a *App) previewCols() int {
	return a.codeColsForGroup(a.activeGroup())
}

// secondView builds the right-hand editor pane.
func (a *App) secondView() secondaryView {
	s := secondaryView{Open: true, Tabs: a.tabRowsFor(1), ActiveTab: a.active2}
	s.Left, s.Width = a.secondGeometry()
	doc := a.groupDoc(1)
	if doc == nil {
		s.Welcome = true

		return s
	}

	s.Active = doc.Path
	s.ActiveDir = dirName(doc.Path)
	s.ActiveName = baseName(doc.Path)
	s.Lang = doc.Lang
	s.LangLabel = languageInfo(doc.Lang).Label
	s.Badge = languageInfo(doc.Lang).Badge
	s.LineCount = len(doc.Lines)
	s.ByteSize = humanSize(doc.ByteSize)
	s.Status = a.statuses[doc.Path]
	s.StatusClass = "st-" + strings.ToLower(s.Status)
	s.Truncated = doc.Truncated
	s.CanEdit = doc.Err == "" && !doc.Binary
	s.Editing = a.editing && a.editPath() == doc.Path
	s.EditDirty = a.dirty[doc.Path]

	if doc.Err != "" {
		s.Empty = true
		s.EmptyKind = "error"
		s.EmptyTitle = "Cannot open " + baseName(doc.Path)
		s.EmptyBody = doc.Err

		return s
	}

	if doc.Binary {
		s.Empty = true
		s.EmptyKind = "binary"
		s.EmptyTitle = "Binary file"
		s.EmptyBody = "The viewer reads text files only. Pick a source file in the explorer."

		return s
	}

	s.Previewable = doc.Lang == langMarkdown
	s.Preview = a.preview && s.Previewable && a.activeGroup() == 1

	cols := a.codeColsForGroup(1)

	if s.Preview {
		doc.ensurePreview(cols)
		start, end, track := a.windowSlice(len(doc.Preview))
		s.TrackTop = track

		if len(doc.Preview) > 0 {
			s.PreviewRows = doc.Preview[start:end]
		}

		return s
	}

	if a.wrap {
		plan := doc.wrapped(cols)
		start, end, track := a.windowSlice(len(plan.Rows))
		s.TrackTop = track
		s.Rows = plan.Rows[start:end]

		return s
	}

	start, end, track := a.windowSlice(len(doc.Lines))
	s.TrackTop = track

	for i := start; i < end; i++ {
		s.Rows = append(s.Rows, codeRow{No: i + 1, Line: i, Segments: doc.Tokens[i]})
	}

	return s
}

// paneRowCount is the visual row count for one document in one pane.
func (a *App) paneRowCount(group int, doc *fileDoc) int {
	if doc == nil || doc.Err != "" || doc.Binary {
		return 0
	}

	cols := a.codeColsForGroup(group)

	if a.activeGroup() == group && a.preview && doc.Lang == langMarkdown {
		doc.ensurePreview(cols)

		return len(doc.Preview)
	}

	if a.wrap {
		return len(doc.wrapped(cols).Rows)
	}

	return len(doc.Lines)
}
