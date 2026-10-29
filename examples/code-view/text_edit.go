package main

import (
	"context"
	"math"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	editUndoLimit = 100
	editPageLines = 20
)

// editSnapshot is one undo point. The buffer text is cheap to copy for the
// files this viewer opens, and it keeps the editor model simple.
type editSnapshot struct {
	Text   string
	Caret  int
	Anchor int
}

// onType inserts printable text at the edit caret.
func (a *App) onType(ctx context.Context, text string) error {
	if a.settingsOpen || !a.editing || a.edit == nil || text == "" {
		return nil
	}

	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.ReplaceAll(text, "\t", tabSpaces)

	return a.insertEditText(ctx, text)
}

// onBackspace deletes the selection or the rune before the caret.
func (a *App) onBackspace(ctx context.Context) error {
	if a.settingsOpen || !a.editing || a.edit == nil {
		return nil
	}

	text, start, end := a.editSelection()
	if start == end && start > 0 {
		start--
	}

	return a.replaceEditRange(ctx, text, start, end, "")
}

// onDeleteWord deletes the word behind the caret.
func (a *App) onDeleteWord(ctx context.Context) error {
	if a.settingsOpen || !a.editing || a.edit == nil {
		return nil
	}

	text, start, end := a.editSelection()
	if start == end {
		start = wordStart([]rune(text), start)
	}

	return a.replaceEditRange(ctx, text, start, end, "")
}

// onCopy returns the selected edit text.
func (a *App) onCopy(context.Context) (string, bool, error) {
	if a.settingsOpen || !a.editing || a.edit == nil {
		return "", false, nil
	}

	text, start, end := a.editSelection()
	if start == end {
		return "", false, nil
	}

	return string([]rune(text)[start:end]), true, nil
}

// onCut returns and removes the selected edit text.
func (a *App) onCut(ctx context.Context) (string, bool, error) {
	if a.settingsOpen || !a.editing || a.edit == nil {
		return "", false, nil
	}

	text, start, end := a.editSelection()
	if start == end {
		return "", false, nil
	}

	out := string([]rune(text)[start:end])

	if err := a.replaceEditRange(ctx, text, start, end, ""); err != nil {
		return "", false, err
	}

	return out, true, nil
}

// onPaste inserts clipboard text at the caret. Outside the editor it pastes
// files into the explorer through the shared Ctrl+V helper, so the chord and
// the menu row behave the same way.
func (a *App) onPaste(ctx context.Context, text string) error {
	if a.settingsOpen || text == "" {
		return nil
	}

	if a.editing && a.edit != nil {
		return a.onType(ctx, text)
	}

	if a.editing {
		return nil
	}

	// Explorer paste shares the Ctrl+V helper. Overlays and text fields keep
	// the event.
	if !a.pasteTargetReady() {
		return nil
	}

	return a.pasteClipboardText(ctx, text)
}

// onSelectAll selects the whole edit buffer.
func (a *App) onSelectAll(ctx context.Context) error {
	if a.settingsOpen || !a.editing || a.edit == nil {
		return nil
	}

	a.editAnchor = 0
	a.editCaret = utf8.RuneCountInString(a.edit.Text())
	a.refresh()

	return a.page.Redraw(ctx)
}

// onUndo restores the previous snapshot, or the last reverted hunk when the
// editor is closed. Hunk undo lives on its own stack so text and hunk
// histories never clobber each other.
func (a *App) onUndo(ctx context.Context) error {
	if a.editing {
		if len(a.editUndo) == 0 {
			return nil
		}

		a.pushRedo()
		last := a.editUndo[len(a.editUndo)-1]
		a.editUndo = a.editUndo[:len(a.editUndo)-1]
		a.restoreEdit(last)
		a.refresh()

		return a.page.Redraw(ctx)
	}

	if a.confirmOpen || a.settingsOpen || a.paletteOpen || a.branchOpen || a.searchOpen {
		return nil
	}

	if !a.canUndoRevert() {
		return nil
	}

	return a.undoRevertHunk(ctx)
}

// onRedo restores a snapshot taken by Undo, or re-applies an undone hunk
// revert when the editor is closed.
func (a *App) onRedo(ctx context.Context) error {
	if a.editing {
		if len(a.editRedo) == 0 {
			return nil
		}

		a.pushUndo()
		next := a.editRedo[len(a.editRedo)-1]
		a.editRedo = a.editRedo[:len(a.editRedo)-1]
		a.restoreEdit(next)
		a.refresh()

		return a.page.Redraw(ctx)
	}

	if a.confirmOpen || a.settingsOpen || a.paletteOpen || a.branchOpen || a.searchOpen {
		return nil
	}

	if !a.canRedoRevert() {
		return nil
	}

	return a.redoRevertHunk(ctx)
}

// editKey handles caret movement and edit-only keys. Printable text arrives in
// onType, so an untouched key returns nil.
func (a *App) editKey(ctx context.Context, key string) error {
	if a.settingsOpen || !a.editing || a.edit == nil {
		return nil
	}

	if a.ctrl {
		switch key {
		case "s":
			a.saveEdit()

			return a.page.Redraw(ctx)
		default:
			return nil
		}
	}

	text, _, _ := a.editSelection()
	runes := []rune(text)
	caret := clampEditPos(a.editCaret, len(runes))
	anchor := clampEditPos(a.editAnchor, len(runes))

	switch key {
	case "arrowleft":
		caret = max(0, caret-1)
	case "arrowright":
		caret = min(len(runes), caret+1)
	case "arrowup":
		caret = a.editMoveVertical(runes, caret, -1)
	case "arrowdown":
		caret = a.editMoveVertical(runes, caret, 1)
	case "home":
		caret = lineStart(runes, caret)
	case "end":
		caret = lineEnd(runes, caret)
	case "pageup":
		caret = editMoveLines(runes, caret, -editPageLines)
	case "pagedown":
		caret = editMoveLines(runes, caret, editPageLines)
	case "delete":
		if caret == anchor && caret < len(runes) {
			return a.replaceEditRange(ctx, text, caret, caret+1, "")
		}

		return a.replaceEditRange(ctx, text, min(anchor, caret), max(anchor, caret), "")
	case "escape":
		a.finishEdit()
		a.refresh()

		return a.page.Redraw(ctx)
	default:
		return nil
	}

	if !a.shift {
		anchor = caret
	}

	a.editCaret, a.editAnchor = caret, anchor
	a.ensureEditCaretVisible()
	a.refresh()

	return a.page.Redraw(ctx)
}

// insertEditText replaces the selection with text and moves the caret.
func (a *App) insertEditText(ctx context.Context, insert string) error {
	text, start, end := a.editSelection()

	return a.replaceEditRange(ctx, text, start, end, insert)
}

// replaceEditRange applies one text change, records undo, and redraws.
func (a *App) replaceEditRange(ctx context.Context, text string, start, end int, insert string) error {
	runes := []rune(text)
	start = clampEditPos(start, len(runes))
	end = clampEditPos(end, len(runes))
	if start > end {
		start, end = end, start
	}

	a.pushUndo()
	a.editRedo = nil

	next := make([]rune, 0, len(runes)-(end-start)+len([]rune(insert)))
	next = append(next, runes[:start]...)
	next = append(next, []rune(insert)...)
	next = append(next, runes[end:]...)

	a.edit.SetText(string(next))
	a.dirty[a.edit.Path] = a.edit.Dirty
	a.editCaret = start + utf8.RuneCountInString(insert)
	a.editAnchor = a.editCaret
	line, _ := a.editDisplayLineCol()
	if doc := a.docs[a.edit.Path]; doc != nil && doc.Syntax != nil {
		doc.Syntax.markDirty(line)
	}
	a.commitEditBuffer()
	a.ensureEditCaretVisible()
	a.refresh()

	return a.page.Redraw(ctx)
}

// editSelection returns the buffer text and the ordered selection offsets.
func (a *App) editSelection() (string, int, int) {
	if a.edit == nil {
		return "", 0, 0
	}

	text := a.edit.Text()
	n := utf8.RuneCountInString(text)
	anchor := clampEditPos(a.editAnchor, n)
	caret := clampEditPos(a.editCaret, n)

	if anchor > caret {
		anchor, caret = caret, anchor
	}

	return text, anchor, caret
}

// pushUndo records the current text before a change.
func (a *App) pushUndo() {
	if a.edit == nil {
		return
	}

	a.editUndo = append(a.editUndo, editSnapshot{
		Text:   a.edit.Text(),
		Caret:  a.editCaret,
		Anchor: a.editAnchor,
	})
	if len(a.editUndo) > editUndoLimit {
		a.editUndo = a.editUndo[1:]
	}
}

// pushRedo records the current text before an undo.
func (a *App) pushRedo() {
	if a.edit == nil {
		return
	}

	a.editRedo = append(a.editRedo, editSnapshot{
		Text:   a.edit.Text(),
		Caret:  a.editCaret,
		Anchor: a.editAnchor,
	})
	if len(a.editRedo) > editUndoLimit {
		a.editRedo = a.editRedo[1:]
	}
}

// restoreEdit installs one snapshot without adding a new undo point.
func (a *App) restoreEdit(snap editSnapshot) {
	if a.edit == nil {
		return
	}

	a.edit.SetText(snap.Text)
	a.dirty[a.edit.Path] = a.edit.Dirty
	a.editCaret = snap.Caret
	a.editAnchor = snap.Anchor
	a.commitEditBuffer()
}

// ensureEditCaretVisible scrolls the page so the caret's visual row is inside
// the editor body.
func (a *App) ensureEditCaretVisible() {
	doc := a.groupDoc(0)
	if doc == nil || !a.editing {
		return
	}

	line, col := a.editDisplayLineCol()
	row := line
	if a.wrap {
		row = caretVisualRow(doc.wrapped(a.codeColsForGroup(0)).Rows, line, col)
	}

	top := a.scrollY / a.zx(rowH)
	visible := a.bodyHeight() / a.zx(rowH)
	if visible < 1 {
		visible = 1
	}

	switch {
	case row < top:
		a.scrollTo(row * a.zx(rowH))
	case row >= top+visible:
		a.scrollTo((row - visible + 1) * a.zx(rowH))
	}
}

// editDisplayLineCol returns the caret line and display column, with tabs
// expanded the same way the code view expands them.
func (a *App) editDisplayLineCol() (int, int) {
	if a.edit == nil {
		return 0, 0
	}

	text := a.edit.Text()
	line, col := offsetLineCol(text, a.editCaret)
	raw := []rune(lineTextAt(text, line))
	if col > len(raw) {
		col = len(raw)
	}

	return line, utf8.RuneCountInString(expandTabs(string(raw[:col])))
}

// decorateEditRows marks the caret and selection on the visible rows.
func (a *App) decorateEditRows(rows []codeRow) {
	if !a.editing || a.edit == nil || len(rows) == 0 {
		return
	}

	text := a.edit.Text()
	caretLine, caretCol := a.editDisplayLineCol()
	caretStart := 0
	if a.wrap {
		if doc := a.groupDoc(0); doc != nil {
			rows := doc.wrapped(a.codeColsForGroup(0)).Rows
			if len(rows) > 0 {
				caretStart = rows[caretVisualRow(rows, caretLine, caretCol)].Start
			}
		}
	}
	selStart, selEnd := a.editAnchor, a.editCaret
	if selStart > selEnd {
		selStart, selEnd = selEnd, selStart
	}

	startLine, startCol := offsetLineCol(text, selStart)
	endLine, endCol := offsetLineCol(text, selEnd)

	// Map the raw selection columns to display columns for the line they sit
	// on. The prefix is expanded because the code view prints no tabs.
	startCol = displayColumn(text, startLine, startCol)
	endCol = displayColumn(text, endLine, endCol)

	for i := range rows {
		row := &rows[i]
		length := segmentRuneLen(row.Segments)

		if row.Line == caretLine && row.Start == caretStart && caretCol >= row.Start && caretCol <= row.Start+length {
			row.Caret = true
			row.CaretLeft = int(float64(caretCol-row.Start) * a.zf(codeCellW))
			if !a.wrap && a.scrollX > 0 {
				row.CaretLeft -= a.scrollX
			}
		}

		if selStart == selEnd || row.Line < startLine || row.Line > endLine {
			continue
		}

		from, to := 0, row.Start+length
		if row.Line == startLine {
			from = max(from, startCol)
		}

		if row.Line == endLine {
			to = min(to, endCol)
		}

		from = max(from, row.Start)
		to = min(to, row.Start+length)
		lineLength := utf8.RuneCountInString(expandTabs(lineTextAt(text, row.Line)))
		newline := row.Line < endLine && row.Start+length >= lineLength
		if from >= to && !newline {
			continue
		}

		left, width := selectionSpan(from, to, row.Start, newline, a.zf(codeCellW))
		row.Selection = true
		row.SelectionLeft = left
		row.SelectionWidth = width
		if !a.wrap {
			row.SelectionLeft -= a.scrollX
		}
	}
}

// selectionSpan is the highlight box for one visual row. The right edge is
// the ceiling of the cell grid, so the last selected glyph stays inside the
// box when the cell width is not a whole pixel.
func selectionSpan(from, to, rowStart int, newline bool, cellW float64) (int, int) {
	cols := to - rowStart
	if newline {
		cols++
	}

	if cols < 0 {
		cols = 0
	}

	left := int(math.Floor(float64(from-rowStart) * cellW))
	right := int(math.Ceil(float64(cols) * cellW))
	if right < left {
		right = left
	}

	return left, right - left
}

// caretVisualRow finds the wrapped visual row that owns a line and column.
func caretVisualRow(rows []codeRow, line, col int) int {
	match := -1
	for i := range rows {
		row := rows[i]
		length := segmentRuneLen(row.Segments)
		if row.Line != line {
			continue
		}

		if col >= row.Start && col <= row.Start+length {
			match = i
		}
	}
	if match >= 0 {
		return match
	}

	return max(0, min(line, len(rows)-1))
}

func segmentRuneLen(segs []segment) int {
	n := 0

	for _, seg := range segs {
		n += utf8.RuneCountInString(seg.Text)
	}

	return n
}

func displayColumn(text string, line, col int) int {
	raw := []rune(lineTextAt(text, line))
	if col > len(raw) {
		col = len(raw)
	}

	return utf8.RuneCountInString(expandTabs(string(raw[:col])))
}

func lineTextAt(text string, line int) string {
	lines := strings.Split(text, "\n")
	if line < 0 || line >= len(lines) {
		return ""
	}

	return lines[line]
}

// offsetLineCol turns a rune offset into a zero-based line and column.
func offsetLineCol(text string, offset int) (int, int) {
	runes := []rune(text)
	offset = clampEditPos(offset, len(runes))
	line, col := 0, 0

	for _, r := range runes[:offset] {
		if r == '\n' {
			line++
			col = 0

			continue
		}

		col++
	}

	return line, col
}

// editMoveVertical follows visual rows when the editor wraps lines.
func (a *App) editMoveVertical(runes []rune, caret, delta int) int {
	if !a.wrap || a.groupDoc(0) == nil {
		return editMoveLine(runes, caret, delta)
	}
	rows := a.groupDoc(0).wrapped(a.codeColsForGroup(0)).Rows
	if len(rows) == 0 {
		return caret
	}
	line, col := a.editDisplayLineCol()
	current := caretVisualRow(rows, line, col)
	next := current + delta
	if next < 0 || next >= len(rows) {
		return caret
	}
	row := rows[next]
	display := min(row.Start+col-rows[current].Start, row.Start+segmentRuneLen(row.Segments))
	text := string(runes)
	rawCol := rawColumnAt(lineTextAt(text, row.Line), display)
	return offsetAtPoint(selectionLines(text), row.Line, rawCol)
}

// editMoveLine keeps the raw column while moving one source line.
func editMoveLine(runes []rune, caret, delta int) int {
	line, col := offsetLineCol(string(runes), caret)
	lines := selectionLines(string(runes))
	line += delta
	if line < 0 || line >= len(lines) {
		return caret
	}

	return offsetAtPoint(lines, line, col)
}

func editMoveLines(runes []rune, caret, delta int) int {
	for range absInt(delta) {
		next := editMoveLine(runes, caret, signInt(delta))
		if next == caret {
			return caret
		}

		caret = next
	}

	return caret
}

func lineStart(runes []rune, caret int) int {
	for i := caret - 1; i >= 0; i-- {
		if runes[i] == '\n' {
			return i + 1
		}
	}

	return 0
}

func lineEnd(runes []rune, caret int) int {
	for i := caret; i < len(runes); i++ {
		if runes[i] == '\n' {
			return i
		}
	}

	return len(runes)
}

func wordStart(runes []rune, caret int) int {
	i := clampEditPos(caret, len(runes))
	for i > 0 && unicode.IsSpace(runes[i-1]) {
		i--
	}

	for i > 0 && !unicode.IsSpace(runes[i-1]) && !unicode.IsPunct(runes[i-1]) {
		i--
	}

	return i
}

func clampEditPos(pos, n int) int {
	if pos < 0 {
		return 0
	}

	if pos > n {
		return n
	}

	return pos
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}

	return n
}

func signInt(n int) int {
	if n < 0 {
		return -1
	}

	return 1
}
