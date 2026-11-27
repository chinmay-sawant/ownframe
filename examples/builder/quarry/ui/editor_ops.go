package ui

import (
	"context"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/chinmay-sawant/ownframe/ui"
)

// Editing operations beyond typing: word and line selection, word jumps,
// indent and comment toggles, bracket and quote auto-pairing, horizontal
// scrolling, pointer drag selection, and find.

func isWordRune(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' }

// wordLeft returns the offset of the start of the word before i.
func wordLeft(s string, i int) int {
	for i > 0 {
		r, n := utf8.DecodeLastRuneInString(s[:i])
		if isWordRune(r) {
			break
		}
		i -= n
	}
	for i > 0 {
		r, n := utf8.DecodeLastRuneInString(s[:i])
		if !isWordRune(r) {
			break
		}
		i -= n
	}
	return i
}

// wordRight returns the offset just past the word after i.
func wordRight(s string, i int) int {
	for i < len(s) {
		r, n := utf8.DecodeRuneInString(s[i:])
		if isWordRune(r) {
			break
		}
		i += n
	}
	for i < len(s) {
		r, n := utf8.DecodeRuneInString(s[i:])
		if !isWordRune(r) {
			break
		}
		i += n
	}
	return i
}

// selectWordAt selects the identifier (or the single symbol) under offset i.
func (e *editor) selectWordAt(i int) {
	i = max(0, min(i, len(e.text)))
	lo, hi := i, i
	for lo > 0 {
		r, n := utf8.DecodeLastRuneInString(e.text[:lo])
		if !isWordRune(r) {
			break
		}
		lo -= n
	}
	for hi < len(e.text) {
		r, n := utf8.DecodeRuneInString(e.text[hi:])
		if !isWordRune(r) {
			break
		}
		hi += n
	}
	if lo == hi && hi < len(e.text) { // on punctuation: take one rune
		_, n := utf8.DecodeRuneInString(e.text[hi:])
		hi += n
	}
	e.anchor, e.caret = lo, hi
}

// selectLineAt selects the whole line holding offset i, newline included.
func (e *editor) selectLineAt(i int) {
	l, _ := e.lineCol(i)
	starts := e.lineStarts()
	e.anchor = starts[l]
	if l+1 < len(starts) {
		e.caret = starts[l+1]
	} else {
		e.caret = len(e.text)
	}
}

// lineRange returns the first and last line touched by the selection; a
// selection that ends at column 0 of a line does not include that line.
func (e *editor) lineRange() (first, last int) {
	lo, hi := e.selection()
	first, _ = e.lineCol(lo)
	last, c := e.lineCol(hi)
	if hi > lo && c == 0 {
		last--
	}
	return first, max(first, last)
}

// mapLines rewrites lines first..last with fn and selects the result.
func (e *editor) mapLines(first, last int, fn func(string) string) {
	e.snapshot()
	lines := strings.Split(e.text, "\n")
	for i := first; i <= last && i < len(lines); i++ {
		lines[i] = fn(lines[i])
	}
	e.text = strings.Join(lines, "\n")
	starts := e.lineStarts()
	e.anchor = starts[first]
	endLine := min(last, len(starts)-1)
	e.caret = len(e.text)
	if endLine+1 < len(starts) {
		e.caret = starts[endLine+1] - 1
	}
}

const indentUnit = "  "

// indent adds one indent unit to every selected line (or inserts it at the caret).
func (e *editor) indent() {
	lo, hi := e.selection()
	if lo == hi {
		e.insert(indentUnit)
		return
	}
	first, last := e.lineRange()
	e.mapLines(first, last, func(s string) string { return indentUnit + s })
}

// outdent removes up to one indent unit from every selected line.
func (e *editor) outdent() {
	first, last := e.lineRange()
	lo, hi := e.selection()
	e.mapLines(first, last, func(s string) string {
		switch {
		case strings.HasPrefix(s, indentUnit):
			return s[len(indentUnit):]
		case strings.HasPrefix(s, " "), strings.HasPrefix(s, "\t"):
			return s[1:]
		}
		return s
	})
	if lo == hi { // a bare caret keeps its place on the line
		e.anchor = min(e.caret, e.anchor)
		e.caret = e.anchor
	}
	_ = hi
}

// toggleComment comments the selected lines with "-- ", or uncomments them
// when every non-blank line is already commented.
func (e *editor) toggleComment() {
	first, last := e.lineRange()
	lines := strings.Split(e.text, "\n")
	all, any := true, false
	for i := first; i <= last && i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		if t == "" {
			continue
		}
		any = true
		if !strings.HasPrefix(t, "--") {
			all = false
		}
	}
	if any && all {
		e.mapLines(first, last, func(s string) string {
			i := strings.Index(s, "--")
			if i < 0 {
				return s
			}
			rest := s[i+2:]
			rest = strings.TrimPrefix(rest, " ")
			return s[:i] + rest
		})
		return
	}
	e.mapLines(first, last, func(s string) string {
		if strings.TrimSpace(s) == "" {
			return s
		}
		return "-- " + s
	})
}

// pairs are the opener/closer pairs that auto-close.
func pairCloser(r rune) (rune, bool) {
	switch r {
	case '(':
		return ')', true
	case '[':
		return ']', true
	case '\'':
		return '\'', true
	case '"':
		return '"', true
	}
	return 0, false
}

// typeText inserts typed text with bracket and quote pairing. It reports
// whether it handled the text.
func (e *editor) typeText(s string) {
	r, n := utf8.DecodeRuneInString(s)
	if n != len(s) {
		e.insert(s)
		return
	}
	// Typing a closer over an identical next character steps over it.
	if (r == ')' || r == ']' || r == '\'' || r == '"') && !e.hasSelection() && e.caret < len(e.text) {
		if nr, _ := utf8.DecodeRuneInString(e.text[e.caret:]); nr == r {
			e.caret++
			e.anchor = e.caret
			return
		}
	}
	closer, ok := pairCloser(r)
	if !ok {
		e.insert(s)
		return
	}
	if e.hasSelection() { // wrap the selection
		sel := e.selectionText()
		lo, _ := e.selection()
		e.insert(s + sel + string(closer))
		e.anchor, e.caret = lo+n, lo+n+len(sel)
		return
	}
	if (r == '\'' || r == '"') && e.caret > 0 {
		if pr, _ := utf8.DecodeLastRuneInString(e.text[:e.caret]); isWordRune(pr) {
			e.insert(s) // an apostrophe after a word is not a quote opener
			return
		}
	}
	e.insert(s + string(closer))
	e.caret -= len(string(closer))
	e.anchor = e.caret
}

// ---- horizontal scroll and geometry ----

// editorViewCols is how many character columns fit beside the gutter.
func (a *App) editorViewCols() int {
	w := a.width - railWidth - edGutterW - edColMargin
	if a.sidebarOpen {
		w -= a.sidebarW + edSidebarGap
	}
	return max(edMinCols, int(float64(w)/edCharW))
}

// longestLine is the length in runes of the widest line.
func (e *editor) longestLine() int {
	best := 0
	for _, ln := range strings.Split(e.text, "\n") {
		best = max(best, utf8.RuneCountInString(ln))
	}
	return best
}

// ensureVisibleCol scrolls horizontally so the caret column is visible.
func (e *editor) ensureVisibleCol(cols int) {
	if e.wrapCols > 0 {
		e.hcol = 0
		return
	}
	_, c := e.lineCol(e.caret)
	if c < e.hcol {
		e.hcol = c
	}
	if c > e.hcol+cols-edHScrollLead {
		e.hcol = c - cols + edHScrollLead
	}
	e.hcol = max(0, min(e.hcol, max(0, e.longestLine()-cols+edHScrollLead)))
}

// colWindow returns the byte range of line [ls, le) covering columns
// hcol..hcol+cols, so only visible text is rendered.
func colWindow(text string, ls, le, hcol, cols int) (int, int) {
	i := ls
	for c := 0; c < hcol && i < le; c++ {
		_, n := utf8.DecodeRuneInString(text[i:])
		i += n
	}
	j := i
	for c := 0; c < cols && j < le; c++ {
		_, n := utf8.DecodeRuneInString(text[j:])
		j += n
	}
	return i, j
}

// editorWrapBox returns the editor-wrap box origin.
func (a *App) editorWrapOrigin() (x, y float64, ok bool) {
	for _, b := range a.page.Boxes() {
		if b.ID == "editor-wrap" {
			return b.X, b.Y, true
		}
	}
	return 0, 0, false
}

// pointToOffset maps a window point inside the editor to a text offset.
func (a *App) pointToOffset(px, py float64) (int, bool) {
	t := a.activeTabPtr()
	wx, wy, ok := a.editorWrapOrigin()
	if t == nil || t.ed == nil || !ok {
		return 0, false
	}
	a.syncWrap(t)
	rows := t.ed.vrows()
	row := max(0, min(t.ed.scroll+int((py-wy-edBodyPad)/edLineH), len(rows)-1))
	col := t.ed.hcol + int((px-wx-edGutterW+edCharW/2)/edCharW)
	return t.ed.offsetInRow(rows, row, max(0, col)), true
}

// ---- pointer clicks, drag selection ----

const multiClickWindow = 450 // milliseconds

// editorPress handles a pointer press: place the caret, extend with Shift, and
// turn a quick second or third press into word or line selection.
func (a *App) editorPress(x, y float64, shift bool) {
	t := a.activeTabPtr()
	off, ok := a.pointToOffset(x, y)
	if t == nil || t.ed == nil || !ok {
		return
	}
	now := a.clock()
	if now.Sub(a.edClickAt).Milliseconds() < multiClickWindow && a.edClickOff == off {
		a.edClicks++
	} else {
		a.edClicks = 1
	}
	a.edClickAt, a.edClickOff = now, off
	a.edPX, a.edPY, a.edSumX, a.edSumY = x, y, 0, 0
	switch {
	case a.edClicks == edDoubleClick:
		t.ed.selectWordAt(off)
	case a.edClicks >= edTripleClick:
		t.ed.selectLineAt(off)
	case shift:
		t.ed.caret = off
	default:
		t.ed.caret, t.ed.anchor = off, off
	}
	a.ac.open = false
}

// editorDrag extends the selection to the pointer as it moves, scrolling when
// the pointer leaves the editor.
func (a *App) editorDrag(ctx context.Context, dx, dy float64) error {
	t := a.activeTabPtr()
	if t == nil || t.ed == nil {
		return nil
	}
	a.edSumX += dx
	a.edSumY += dy
	px, py := a.edPX+a.edSumX, a.edPY+a.edSumY
	if wx, wy, ok := a.editorWrapOrigin(); ok {
		rows := a.editorRows()
		switch {
		case py < wy:
			t.ed.scroll = max(0, t.ed.scroll-1)
		case py > wy+float64(rows*edLineH):
			t.ed.scroll = min(max(0, t.ed.rowCount()-rows), t.ed.scroll+1)
		}
		cols := a.editorViewCols()
		switch {
		case px < wx+edGutterW:
			t.ed.hcol = max(0, t.ed.hcol-1)
		case px > wx+edGutterW+float64(cols)*edCharW:
			t.ed.hcol = min(max(0, t.ed.longestLine()-cols+edHScrollLead), t.ed.hcol+1)
		}
	}
	if off, ok := a.pointToOffset(px, py); ok {
		t.ed.caret = off
	}
	return a.redraw(ctx)
}

// ---- find ----

type findState struct {
	open    bool
	query   string
	matches []int // byte offsets
	idx     int
}

func (a *App) openFind(ctx context.Context) error {
	t := a.activeTabPtr()
	if t == nil || t.kind != tabQuery || t.ed == nil {
		return nil
	}
	a.find.open = true
	if t.ed.hasSelection() && !strings.Contains(t.ed.selectionText(), "\n") {
		a.find.query = t.ed.selectionText()
	}
	a.refreshFind()
	a.page.Form().SetFormValue("find-input", a.find.query)
	a.page.Form().RequestFocus("find-input")
	return a.redraw(ctx)
}

func (a *App) closeFind(ctx context.Context) error {
	a.find = findState{}
	_ = a.page.Form().Focus(ctx, "")
	a.focus = "editor"
	return a.redraw(ctx)
}

// refreshFind recomputes the case-insensitive matches of the query.
func (a *App) refreshFind() {
	a.find.matches = a.find.matches[:0]
	t := a.activeTabPtr()
	if t == nil || t.ed == nil || a.find.query == "" {
		return
	}
	text, q := strings.ToLower(t.ed.text), strings.ToLower(a.find.query)
	for i := 0; ; {
		j := strings.Index(text[i:], q)
		if j < 0 {
			break
		}
		a.find.matches = append(a.find.matches, i+j)
		i += j + max(1, len(q))
	}
}

// findStep selects the next (dir 1) or previous (dir -1) match.
func (a *App) findStep(ctx context.Context, dir int) error {
	t := a.activeTabPtr()
	if t == nil || t.ed == nil {
		return nil
	}
	a.refreshFind()
	n := len(a.find.matches)
	if n == 0 {
		return a.redraw(ctx)
	}
	pos := t.ed.caret
	if dir < 0 {
		pos = min(t.ed.anchor, t.ed.caret)
	}
	idx := -1
	if dir > 0 {
		for i, m := range a.find.matches {
			if m >= pos {
				idx = i
				break
			}
		}
		if idx < 0 {
			idx = 0
		}
	} else {
		for i := n - 1; i >= 0; i-- {
			if a.find.matches[i] < pos {
				idx = i
				break
			}
		}
		if idx < 0 {
			idx = n - 1
		}
	}
	a.find.idx = idx
	m := a.find.matches[idx]
	t.ed.anchor, t.ed.caret = m, m+len(a.find.query)
	t.ed.ensureVisible(a.editorRows())
	t.ed.ensureVisibleCol(a.editorViewCols())
	return a.redraw(ctx)
}

// findBar is the search strip above the editor body.
func (a *App) findBar() *ui.Element {
	if !a.find.open {
		return nil
	}
	count := "No matches"
	if n := len(a.find.matches); n > 0 {
		count = strconv.Itoa(min(a.find.idx, n-1)+1) + " of " + strconv.Itoa(n)
	}
	return ui.Div().ID("find-bar").Class("qr-findbar").Children(
		a.icon("search", toneMuted, iconSmall),
		ui.Input().ID("find-input").Class("qr-input qr-find-input").Attr("type", "text").Attr("placeholder", "Find").
			Attr("aria-label", "Find in query").Attr("value", a.find.query),
		ui.Span().ID("find-count").Class("qr-muted qr-nowrap").Text(count),
		a.tbButton("find-prev", "Prev", "qr-btn qr-btn-ghost", false, func(ctx context.Context) error { return a.findStep(ctx, -1) }),
		a.tbButton("find-next", "Next", "qr-btn qr-btn-ghost", false, func(ctx context.Context) error { return a.findStep(ctx, 1) }),
		ui.Button("").ID("find-close").Class("qr-iconbtn").Attr("type", "button").Attr("aria-label", "Close find").
			Children(a.icon("x", toneMuted, iconSmall)).OnClick(a.closeFind),
	)
}
