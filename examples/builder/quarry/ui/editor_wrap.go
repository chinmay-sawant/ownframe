package ui

import (
	"strings"
	"unicode/utf8"
)

// vrow is one painted row of the editor: a whole logical line without word
// wrap, or one chunk of it with wrap on.
type vrow struct {
	line       int  // logical line index
	start, end int  // byte range of the row's text, newline excluded
	first      bool // first row of its logical line (shows the line number)
	last       bool // last row of its logical line
}

// vrows lays the buffer out as painted rows. With wrapCols > 0 a line longer
// than wrapCols characters breaks after its last space that fits (or hard at
// the column limit when it has none).
func (e *editor) vrows() []vrow {
	var rows []vrow
	text := e.text
	ls, line := 0, 0
	for {
		le := strings.IndexByte(text[ls:], '\n')
		if le < 0 {
			le = len(text)
		} else {
			le += ls
		}
		rows = e.wrapLine(rows, line, ls, le)
		if le >= len(text) {
			return rows
		}
		ls, line = le+1, line+1
	}
}

func (e *editor) wrapLine(rows []vrow, line, ls, le int) []vrow {
	pos, first := ls, true
	for {
		if e.wrapCols <= 0 || utf8.RuneCountInString(e.text[pos:le]) <= e.wrapCols {
			return append(rows, vrow{line: line, start: pos, end: le, first: first, last: true})
		}
		end := pos
		for c := 0; c < e.wrapCols; c++ {
			_, n := utf8.DecodeRuneInString(e.text[end:])
			end += n
		}
		brk := end
		if sp := strings.LastIndexByte(e.text[pos:end], ' '); sp > 0 {
			brk = pos + sp + 1
		}
		rows = append(rows, vrow{line: line, start: pos, end: brk, first: first})
		pos, first = brk, false
	}
}

// rowCount is the number of painted rows, the unit of vertical scrolling.
func (e *editor) rowCount() int {
	if e.wrapCols <= 0 {
		return e.lineCount()
	}
	return len(e.vrows())
}

// rowOf returns the painted row holding byte offset off and the rune column
// inside it. A caret at a wrap point belongs to the row that follows.
func (e *editor) rowOf(off int) (row, col int) {
	rows := e.vrows()
	for i, r := range rows {
		if off < r.end || (off == r.end && r.last) {
			return i, utf8.RuneCountInString(e.text[r.start:min(max(off, r.start), r.end)])
		}
	}
	last := len(rows) - 1
	return last, utf8.RuneCountInString(e.text[rows[last].start:rows[last].end])
}

// offsetInRow returns the offset at rune column col of painted row row,
// clamped so the caret stays on that row.
func (e *editor) offsetInRow(rows []vrow, row, col int) int {
	row = max(0, min(row, len(rows)-1))
	r := rows[row]
	if !r.last {
		col = min(col, max(0, utf8.RuneCountInString(e.text[r.start:r.end])-1))
	}
	off := r.start
	for i := 0; i < col && off < r.end; i++ {
		_, n := utf8.DecodeRuneInString(e.text[off:])
		off += n
	}
	return off
}

// moveRow moves the caret d painted rows up or down, keeping the column.
func (e *editor) moveRow(d int, extend bool) {
	rows := e.vrows()
	row, col := e.rowOf(e.caret)
	e.caret = e.offsetInRow(rows, row+d, col)
	if !extend {
		e.anchor = e.caret
	}
}

// syncWrap applies the word-wrap preference to the active editor.
func (a *App) syncWrap(t *workTab) {
	if t == nil || t.ed == nil {
		return
	}
	if a.wrap {
		t.ed.wrapCols = a.editorViewCols()
		t.ed.hcol = 0
	} else {
		t.ed.wrapCols = 0
	}
}
