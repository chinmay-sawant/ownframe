package main

import (
	"strings"
)

const paneMinWidth = 220

// codeClipSlack keeps the last glyph inside the row. The painted advance
// matches a.zf(codeCellW); without this slack a full row still enters the right padding.
const codeClipSlack = 8

// Preview rows do not use the source gutter. .prow pads both sides, .pmark
// takes an 18 px marker whose 7 px right padding is ink headroom, and a quote
// row adds a 2 px left border. The budget subtracts the marker box, repeats
// the marker padding on the right, and rounds the measured 7.321 px advance of
// the 12.2 px body face up to 7.5 px, so a full row keeps at least one glyph
// clear of the right padding.
const (
	previewPadX        = 22
	previewMarkW       = 18
	previewMarkPad     = 7
	previewQuoteBorder = 2
	previewCellW       = 7.5
	previewSlack       = 8

	// previewHeadingCellW covers the 14 px heading face. A heading row drops
	// the marker column, so its budget starts after the row padding only.
	previewHeadingCellW = 8.5
)

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
		return a.zx(chromeLeft), total
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

	return a.zx(chromeLeft), left
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
	avail := width - a.zx(codeGutter) - a.zx(codePadRight) - a.zx(editorBorderPad) - a.zx(codeClipSlack)
	cols := int(float64(avail) / a.zf(codeCellW))
	if cols < wrapMinCols {
		cols = wrapMinCols
	}

	return cols
}

// groupEditorWidth is the pane width for one editor group.
func (a *App) groupEditorWidth(group int) int {
	width := a.editorWidth()
	if group == 1 {
		_, width = a.secondGeometry()
	} else if a.split {
		_, width = a.paneGeometry()
	}

	return width
}

// codeColsForGroup returns the cell count for one editor group. A group that
// is showing Markdown preview uses the preview row's own padding and marker.
func (a *App) codeColsForGroup(group int) int {
	width := a.groupEditorWidth(group)
	if a.preview && a.activeGroup() == group {
		return a.previewColsForDoc(width, a.groupDoc(group))
	}

	return a.codeColsFor(width)
}

// previewCols is the column budget for rendered Markdown.
func (a *App) previewCols() int {
	return a.previewColsForDoc(a.groupEditorWidth(a.activeGroup()), a.activeDoc())
}

// previewColsFor is the cell count inside a preview pane. The budget is
// conservative: every subtracted length comes from the preview stylesheet and
// a.zf(previewCellW) is wider than the painted advance, so a full row cannot reach
// the right padding even when the face resolves a little wide.
func (a *App) previewColsFor(width int) int {
	avail := width - a.zx(previewPadX) - a.zx(previewPadX) - a.zx(previewMarkW) - a.zx(previewMarkPad) - a.zx(previewQuoteBorder) - a.zx(previewSlack)
	cols := int(float64(avail) / a.zf(previewCellW))
	if cols < wrapMinCols {
		cols = wrapMinCols
	}

	return cols
}

// previewColsForDoc narrows the whole preview to the heading budget when a
// heading uses the larger face and would not fit the body budget.
func (a *App) previewColsForDoc(width int, doc *fileDoc) int {
	cols := a.previewColsFor(width)
	if doc == nil {
		return cols
	}

	headingCols := a.previewHeadingColsFor(width)
	if headingCols >= cols || !hasWideHeading(doc.Lines, headingCols) {
		return cols
	}

	return headingCols
}

// previewHeadingColsFor is the heading cell count inside a preview pane. A
// heading row has no marker column, so only the row padding and slack apply.
func (a *App) previewHeadingColsFor(width int) int {
	avail := width - a.zx(previewPadX) - a.zx(previewPadX) - a.zx(previewSlack)
	cols := int(float64(avail) / a.zf(previewHeadingCellW))
	if cols < wrapMinCols {
		cols = wrapMinCols
	}

	return cols
}

// hasWideHeading reports whether a heading's visible text is longer than cols.
// Fenced code is skipped so a comment starting with # cannot narrow the page.
func hasWideHeading(lines []string, cols int) bool {
	inFence := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if isFenceLine(trimmed) {
			inFence = !inFence

			continue
		}

		if inFence || !strings.HasPrefix(trimmed, "#") {
			continue
		}

		r := []rune(trimmed)
		level, offset := headingLevel(r)
		if level == 0 {
			continue
		}

		if len([]rune(plainInline(strings.TrimSpace(string(r[offset:]))))) > cols {
			return true
		}
	}

	return false
}

// secondView builds the right-hand editor pane.
func (a *App) secondView() secondaryView {
	s := secondaryView{Open: true, ActiveTab: a.active2}
	s.Left, s.Width = a.secondGeometry()
	s.Tabs, s.TabsPrev, s.TabsNext = a.windowTabRows(1)
	doc := a.groupDoc(1)
	if doc == nil {
		s.Welcome = true

		return s
	}

	s.Active = doc.Path
	s.ActiveDir = dirName(doc.Path)
	s.ActiveName = baseName(doc.Path)
	winW, _ := a.page.Size()
	s.ActiveNameShort, s.ActiveDirShort = clipSecondCrumbPath(s.Width, winW, s.ActiveName, s.ActiveDir)
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
		start, end, track := a.windowSliceForGroup(1, len(doc.Preview))
		s.TrackTop = track

		if len(doc.Preview) > 0 {
			s.PreviewRows = doc.Preview[start:end]
		}

		return s
	}

	if a.wrap {
		plan := doc.wrapped(cols)
		start, end, track := a.windowSliceForGroup(1, len(plan.Rows))
		s.TrackTop = track
		s.Rows = cloneCodeRows(plan.Rows[start:end])
		a.decorateCodeRows(doc, s.Rows, 1)

		return s
	}

	start, end, track := a.windowSliceForGroup(1, len(doc.Lines))
	s.TrackTop = track
	if !a.wrap {
		s.ScrollX = a.scrollX
	}

	for i := start; i < end; i++ {
		row := codeRow{No: i + 1, Line: i, Segments: doc.Tokens[i]}
		if !a.wrap {
			row.ScrollX = a.scrollX
		}
		s.Rows = append(s.Rows, row)
	}

	a.decorateCodeRows(doc, s.Rows, 1)

	return s
}

// cloneCodeRows copies a visible slice so gutter marks never leak back into a
// cached wrap plan.
func cloneCodeRows(rows []codeRow) []codeRow {
	if len(rows) == 0 {
		return nil
	}

	out := make([]codeRow, len(rows))
	copy(out, rows)

	return out
}

// paneRowCount is the visual row count for one document in one pane.
func (a *App) paneRowCount(group int, doc *fileDoc) int {
	if doc == nil || doc.Err != "" || doc.Binary {
		return 0
	}

	cols := a.codeColsForGroup(group)

	if a.activeGroup() == group && a.preview && doc.Lang == langMarkdown {
		doc.ensurePreview(cols)
		n := len(doc.Preview)
		if n == 0 {
			return 0
		}

		return n + a.editorTailRows()
	}

	n := len(doc.Lines)
	if a.wrap {
		n = len(doc.wrapped(cols).Rows)
	}

	if n == 0 {
		return 0
	}

	// Blank rows under the file let the last real row scroll above the
	// status bar and the file navigator. The slice never draws them.
	return n + a.editorTailRows()
}
