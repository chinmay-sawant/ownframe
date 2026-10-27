package main

import "strings"

// toggleEdit switches the focused editor between the token view and a plain
// text buffer. Editing stays in the left group; the right group is a view.
func (a *App) toggleEdit() {
	if a.split && a.focusGroup == 1 {
		a.focusGroup = 0
	}

	doc := a.groupDoc(0)
	if doc == nil || doc.Err != "" || doc.Binary {
		a.showToast("info", "Edit", "Open a text file in the left pane before editing.")
		a.refresh()

		return
	}

	if a.editing {
		a.finishEdit()
		a.showToast("info", "Source view", "Edits stay in memory until you save.")
		a.refresh()

		return
	}

	raw, err := a.ws.readFile(doc.Path)
	if a.dirty[doc.Path] && doc.Err == "" {
		raw = []byte(strings.Join(doc.Lines, "\n") + "\n")
		err = nil
	}

	if err != nil {
		a.showToast("error", "Cannot edit", err.Error())
		a.refresh()

		return
	}

	a.edit = newEditBuffer(doc.Path, raw)
	a.editing = true
	a.editCaret = 0
	a.editAnchor = 0
	a.editUndo = nil
	a.editRedo = nil
	a.preview = false
	a.findOpen = false
	a.resetScroll()
	a.ensureEditCaretVisible()
	a.refresh()
}

// beginEditAtLine enters edit mode with the caret on the clicked source line.
func (a *App) beginEditAtLine(line int) {
	a.inlineCard = inlineCard{}

	if !a.editing {
		a.toggleEdit()
	}

	if a.edit == nil {
		return
	}

	runes := []rune(a.edit.Text())
	offset := 0

	for i := 0; i < line; i++ {
		next := offset
		for next < len(runes) && runes[next] != '\n' {
			next++
		}

		if next >= len(runes) {
			break
		}

		offset = next + 1
	}

	a.editCaret = offset
	a.editAnchor = offset
	a.ensureEditCaretVisible()
	a.refresh()
}

// editPath is the file with a live edit buffer.
func (a *App) editPath() string {
	if a.edit == nil {
		return ""
	}

	return a.edit.Path
}

// applyEditText copies the textarea value into the buffer. SetData runs on
// the next state change, so typing does not rebuild the page per keystroke.
func (a *App) applyEditText(text string) {
	if a.edit == nil {
		return
	}

	a.edit.SetText(text)
	a.dirty[a.edit.Path] = a.edit.Dirty
	a.view.EditDirty = a.edit.Dirty
}

// finishEdit folds the in-memory buffer back into the document and leaves
// edit mode. Unsaved text remains visible and is marked dirty.
func (a *App) finishEdit() {
	if !a.editing {
		return
	}

	a.commitEditBuffer()
	a.editing = false
	a.edit = nil
	a.editCaret = 0
	a.editAnchor = 0
	a.editUndo = nil
	a.editRedo = nil
}

// commitEditBuffer updates the document behind the buffer without writing to
// disk, so the source view shows the unsaved text.
func (a *App) commitEditBuffer() {
	if a.edit == nil {
		return
	}

	doc := a.docs[a.edit.Path]
	if doc == nil {
		return
	}

	text := a.edit.Text()
	lines := make([]string, len(a.edit.Lines))

	for i, line := range a.edit.Lines {
		lines[i] = expandTabs(strings.TrimSuffix(line, "\r"))
	}

	doc.Lines = lines
	doc.Tokens = tokenize(doc.Lang, lines)
	doc.Outline = outline(doc.Lang, lines)
	doc.ByteSize = int64(len(text))
	doc.Truncated = len(lines) > maxFileLines
	doc.Preview = nil
	doc.PreviewLine = nil
	doc.Wrap = nil
	a.dirty[doc.Path] = a.edit.Dirty
}

// saveEdit writes the live buffer to the workspace.
func (a *App) saveEdit() {
	if a.edit == nil {
		return
	}

	path := a.edit.Path
	data := a.edit.Save()

	if err := a.ws.writeFile(path, data); err != nil {
		a.showToast("error", "Save failed", err.Error())
		a.refresh()

		return
	}

	a.edit.MarkSaved()
	a.dirty[path] = false
	a.docs[path] = a.ws.openDoc(path)
	a.commitEditBuffer()
	a.showToast("success", "File saved", path)
	a.refresh()
}

// dirtyCount is the number of files with unsaved edits.
func (a *App) dirtyCount() int {
	n := 0

	for _, dirty := range a.dirty {
		if dirty {
			n++
		}
	}

	return n
}
