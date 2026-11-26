package ui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/virtual"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/sqlkit"
)

// Editor metrics. The editor draws each line itself (the engine has no
// highlighted textarea), so these must match base.css.
const (
	edLineH   = 18
	edCharW   = 7.2
	edPadLeft = 0
	edGutterW = 36
)

type editSnap struct {
	text  string
	caret int
}

// editor is the SQL buffer with caret, selection, scroll, and undo history.
type editor struct {
	text     string
	caret    int
	anchor   int // selection anchor; equal to caret when nothing is selected
	scroll   int // first visible line
	hcol     int // first visible column
	wrapCols int // word-wrap width in characters, 0 when wrapping is off
	undo     []editSnap
	redo     []editSnap

	tokText string
	tokens  []sqlkit.Token
}

func newEditor(text string) *editor {
	return &editor{text: text, caret: len(text), anchor: len(text)}
}

func (e *editor) snapshot() {
	e.undo = append(e.undo, editSnap{e.text, e.caret})
	if len(e.undo) > undoDepth {
		e.undo = e.undo[1:]
	}
	e.redo = nil
}

func (e *editor) selection() (lo, hi int) {
	lo, hi = e.anchor, e.caret
	if lo > hi {
		lo, hi = hi, lo
	}
	return lo, hi
}

func (e *editor) hasSelection() bool { return e.anchor != e.caret }

// SelectionText returns the selected text, or "".
func (e *editor) selectionText() string {
	lo, hi := e.selection()
	return e.text[lo:hi]
}

// insert replaces the selection with s and moves the caret after it.
func (e *editor) insert(s string) {
	if s == "" && !e.hasSelection() {
		return
	}
	e.snapshot()
	lo, hi := e.selection()
	e.text = e.text[:lo] + s + e.text[hi:]
	e.caret = lo + len(s)
	e.anchor = e.caret
}

func (e *editor) backspace() {
	if e.hasSelection() {
		e.insert("")
		return
	}
	if e.caret == 0 {
		return
	}
	e.snapshot()
	_, n := utf8.DecodeLastRuneInString(e.text[:e.caret])
	e.text = e.text[:e.caret-n] + e.text[e.caret:]
	e.caret -= n
	e.anchor = e.caret
}

func (e *editor) deleteForward() {
	if e.hasSelection() {
		e.insert("")
		return
	}
	if e.caret >= len(e.text) {
		return
	}
	e.snapshot()
	_, n := utf8.DecodeRuneInString(e.text[e.caret:])
	e.text = e.text[:e.caret] + e.text[e.caret+n:]
}

func (e *editor) undoStep() {
	if len(e.undo) == 0 {
		return
	}
	e.redo = append(e.redo, editSnap{e.text, e.caret})
	s := e.undo[len(e.undo)-1]
	e.undo = e.undo[:len(e.undo)-1]
	e.text, e.caret, e.anchor = s.text, s.caret, s.caret
}

func (e *editor) redoStep() {
	if len(e.redo) == 0 {
		return
	}
	e.undo = append(e.undo, editSnap{e.text, e.caret})
	s := e.redo[len(e.redo)-1]
	e.redo = e.redo[:len(e.redo)-1]
	e.text, e.caret, e.anchor = s.text, s.caret, s.caret
}

// lineStarts returns the byte offset of each line start.
func (e *editor) lineStarts() []int {
	starts := []int{0}
	for i := 0; i < len(e.text); i++ {
		if e.text[i] == '\n' {
			starts = append(starts, i+1)
		}
	}
	return starts
}

func (e *editor) lineCount() int { return strings.Count(e.text, "\n") + 1 }

// lineCol returns the zero-based line and rune column of a byte offset.
func (e *editor) lineCol(off int) (int, int) {
	l, c := sqlkit.LineCol(e.text, off)
	return l - 1, c - 1
}

// offsetAt returns the byte offset of line, rune column col (clamped).
func (e *editor) offsetAt(line, col int) int {
	starts := e.lineStarts()
	line = max(0, min(line, len(starts)-1))
	s := starts[line]
	end := len(e.text)
	if line+1 < len(starts) {
		end = starts[line+1] - 1
	}
	off := s
	for i := 0; i < col && off < end; i++ {
		_, n := utf8.DecodeRuneInString(e.text[off:])
		off += n
	}
	return off
}

func (e *editor) move(line, col int, extend bool) {
	e.caret = e.offsetAt(line, col)
	if !extend {
		e.anchor = e.caret
	}
}

// key applies a caret key and reports whether it was one.
func (e *editor) key(key string, shift bool) bool {
	extend := shift
	name := strings.TrimPrefix(key, "shift+")
	if strings.HasPrefix(key, "shift+") {
		extend = true
	}
	ctrl := strings.HasPrefix(name, "ctrl+")
	name = strings.TrimPrefix(name, "ctrl+")
	l, c := e.lineCol(e.caret)
	switch name {
	case "arrowleft":
		switch {
		case ctrl:
			e.caret = wordLeft(e.text, e.caret)
			if !extend {
				e.anchor = e.caret
			}
		case !extend && e.hasSelection():
			lo, _ := e.selection()
			e.caret, e.anchor = lo, lo
		case e.caret > 0:
			_, n := utf8.DecodeLastRuneInString(e.text[:e.caret])
			e.caret -= n
			if !extend {
				e.anchor = e.caret
			}
		}
	case "arrowright":
		switch {
		case ctrl:
			e.caret = wordRight(e.text, e.caret)
			if !extend {
				e.anchor = e.caret
			}
		case !extend && e.hasSelection():
			_, hi := e.selection()
			e.caret, e.anchor = hi, hi
		case e.caret < len(e.text):
			_, n := utf8.DecodeRuneInString(e.text[e.caret:])
			e.caret += n
			if !extend {
				e.anchor = e.caret
			}
		}
	case "arrowup":
		if e.wrapCols > 0 {
			e.moveRow(-1, extend)
		} else {
			e.move(l-1, c, extend)
		}
	case "arrowdown":
		if e.wrapCols > 0 {
			e.moveRow(1, extend)
		} else {
			e.move(l+1, c, extend)
		}
	case "pageup":
		e.move(l-editorPageLines, c, extend)
	case "pagedown":
		e.move(l+editorPageLines, c, extend)
	case "home":
		if ctrl {
			e.move(0, 0, extend)
		} else {
			e.move(l, 0, extend)
		}
	case "end":
		if ctrl {
			e.caret = len(e.text)
			if !extend {
				e.anchor = e.caret
			}
		} else {
			e.move(l, editorEndColumn, extend)
		}
	default:
		return false
	}
	return true
}

// ensureVisible scrolls so the caret line shows inside rows lines.
func (e *editor) ensureVisible(rows int) {
	l, _ := e.rowOf(e.caret)
	if l < e.scroll {
		e.scroll = l
	}
	if l >= e.scroll+rows {
		e.scroll = l - rows + 1
	}
	e.scroll = max(0, min(e.scroll, max(0, e.rowCount()-rows)))
}

func (e *editor) tokensFor(d sqlkit.Dialect) []sqlkit.Token {
	if e.tokens == nil || e.tokText != e.text {
		e.tokens = sqlkit.Tokenize(e.text, d)
		e.tokText = e.text
	}
	return e.tokens
}

// ---- rendering ----

func synClass(k sqlkit.TokenKind) string {
	switch k {
	case sqlkit.KindKeyword:
		return "qr-syn-kw"
	case sqlkit.KindFunction:
		return "qr-syn-fn"
	case sqlkit.KindType:
		return "qr-syn-type"
	case sqlkit.KindString:
		return "qr-syn-str"
	case sqlkit.KindNumber:
		return "qr-syn-num"
	case sqlkit.KindComment:
		return "qr-syn-comment"
	case sqlkit.KindOperator, sqlkit.KindParam:
		return "qr-syn-op"
	}
	return ""
}

// editorRows is how many lines fit in the editor body.
func (a *App) editorRows() int {
	h := a.editorH - edToolbarH - edLineH/edHalf
	if a.find.open {
		h -= 36
	}
	return max(editorMinRows, h/edLineH)
}

// renderEditorBody draws the visible window of lines with gutter, active-line
// highlight, selection, caret, and syntax colors.
func (a *App) renderEditorBody(t *workTab) *ui.Element {
	ed := t.ed
	d := a.dialect()
	toks := ed.tokensFor(d)
	a.syncWrap(t)
	vr := ed.vrows()
	total := len(vr)
	rows := a.editorRows()
	win := virtual.Window(float64(ed.scroll*edLineH), float64(rows*edLineH), edLineH, total, 1)
	caretLine, _ := ed.lineCol(ed.caret)
	caretRow, _ := ed.rowOf(ed.caret)
	selLo, selHi := ed.selection()
	cols := a.editorViewCols()
	focus := a.focus == "editor"
	body := ui.Div().ID("editor-body").Class("qr-ed-body").Attr("role", "textbox").
		Attr("aria-multiline", "true").Attr("aria-label", "SQL editor")
	for vi := win.Start; vi < win.End; vi++ {
		ls, le, ln := vr[vi].start, vr[vi].end, vr[vi].line
		row := ui.Div().ID(fmt.Sprintf("ed-line-%d", vi)).Class("qr-ed-line")
		switch ln {
		case t.rowErrLine:
			row.Class("qr-ed-line qr-ed-line-err")
		case caretLine:
			row.Class("qr-ed-line qr-ed-line-active")
		}
		if ed.wrapCols == 0 {
			ls, le = colWindow(ed.text, ls, le, ed.hcol, cols)
		}
		num := ""
		if vr[vi].first {
			num = fmt.Sprint(ln + 1)
		}
		row.Children(ui.Span().Class("qr-ed-gutter").Text(num))
		caretAt := -1
		if vi == caretRow {
			caretAt = ed.caret
		}
		code := ui.Div().Class("qr-ed-code")
		for _, r := range lineRuns(ed.text, toks, ls, le, selLo, selHi, caretAt) {
			if r.caret {
				if focus {
					code.Children(ui.Span().Class("qr-caret"))
				}
				continue
			}
			cls := "qr-run " + r.class
			if r.selected {
				cls += " qr-sel"
			}
			code.Children(ui.Span().Class(cls).Text(r.text))
		}
		if ls == le {
			code.Children(ui.Span().Class("qr-run").Text(" "))
		}
		row.Children(code)
		body.Children(row)
	}
	return body
}

type run struct {
	text     string
	class    string
	selected bool
	caret    bool
}

// lineRuns splits [ls, le) at token, selection, and caret boundaries.
func lineRuns(text string, toks []sqlkit.Token, ls, le, selLo, selHi, caret int) []run {
	cuts := []int{ls, le}
	i := sort.Search(len(toks), func(i int) bool { return toks[i].End > ls })
	for ; i < len(toks) && toks[i].Start < le; i++ {
		cuts = append(cuts, max(ls, toks[i].Start), min(le, toks[i].End))
	}
	for _, c := range []int{selLo, selHi, caret} {
		if c > ls && c < le {
			cuts = append(cuts, c)
		}
	}
	sort.Ints(cuts)
	var out []run
	caretDone := false
	for k := 0; k+1 < len(cuts); k++ {
		s, e := cuts[k], cuts[k+1]
		if s == e {
			continue
		}
		if caret == s && !caretDone {
			out = append(out, run{caret: true})
			caretDone = true
		}
		cls := ""
		j := sort.Search(len(toks), func(i int) bool { return toks[i].End > s })
		if j < len(toks) && toks[j].Start <= s {
			cls = synClass(toks[j].Kind)
		}
		out = append(out, run{text: text[s:e], class: cls, selected: selLo != selHi && s >= selLo && e <= selHi})
	}
	if caret >= ls && caret <= le && !caretDone {
		out = append(out, run{caret: true})
	}
	return out
}

// editorPane builds toolbar, body, and the autocomplete popover.
func (a *App) editorPane(t *workTab) *ui.Element {
	a.syncWrap(t)
	rows := a.editorRows()
	cols := a.editorViewCols()
	longest := t.ed.longestLine()
	if a.wrap {
		longest = 0
	}
	pane := ui.Div().ID("editor-pane").Class("qr-editor").Height(ui.Px(float64(a.editorH))).Children(
		a.editorToolbar(t),
		a.findBar(),
		ui.Div().ID("editor-wrap").Class("qr-ed-wrap").Height(ui.Px(float64(rows*edLineH+edLineH/2))).Children(
			a.renderEditorBody(t),
			a.scrollbarFor("editor", "Editor scroll", t.ed.rowCount()*edLineH, rows*edLineH, t.ed.scroll*edLineH,
				func(o int) { t.ed.scroll = o / edLineH }),
			a.hbar("editor", "Editor horizontal scroll", int(float64(longest)*edCharW), int(float64(cols)*edCharW), int(float64(t.ed.hcol)*edCharW),
				func(o int) { t.ed.hcol = int(float64(o)/edCharW + edRound) }),
			a.autocompletePopover(t),
		),
	)
	return pane
}

func rowLimits() []int { return []int{100, thousandRows, tenThousandRows, 0} }

func limitLabel(n int) string {
	switch n {
	case 0:
		return "Unlimited"
	case thousandRows:
		return "1,000"
	case tenThousandRows:
		return "10,000"
	}
	return fmt.Sprint(n)
}

func (a *App) editorToolbar(t *workTab) *ui.Element {
	busy := a.running != nil
	tb := ui.Div().ID("editor-toolbar").Class("qr-toolbar")
	tb.Children(
		a.tbButton("ed-run", "Run", "qr-btn qr-btn-primary", busy, func(ctx context.Context) error { return a.runFromKey(ctx, false) }),
	)
	tb.Children(a.overflowGroup("ed", a.mainWidth()-edToolbarFixed-tbRunWidth, []tbItem{
		{id: "ed-run-sel", label: "Run Selection", class: "qr-btn", disabled: busy, fn: a.runSelection},
		{id: "ed-explain", label: "Explain", class: "qr-btn", disabled: busy, fn: func(ctx context.Context) error { return a.startExplain(ctx, false) }},
		{id: "ed-analyze", label: "Explain Analyze", class: "qr-btn", disabled: busy, fn: func(ctx context.Context) error { return a.startExplain(ctx, true) }},
		{id: "ed-format", label: "Format", class: "qr-btn", fn: a.formatEditor},
		{id: "ed-snippet", label: "Save Snippet", class: "qr-btn", fn: a.saveSnippet},
	}, []tbItem{{id: "ed-wrap", label: "Word wrap", check: true, on: a.wrap, fn: a.toggleWrap}})...)
	if busy {
		tb.Children(a.tbButton("ed-cancel", "Cancel", "qr-btn qr-btn-danger", false, func(ctx context.Context) error {
			a.cancelRunning()
			return a.redraw(ctx)
		}))
		tb.Children(ui.Span().ID("ed-running").Class("qr-running").Text(fmt.Sprintf("Running %s", fmtDur(a.clock().Sub(a.running.started)))))
	}
	tb.Children(ui.Div().Class("qr-spacer"), ui.Span().Class("qr-muted").Text("Limit"))
	seg := ui.Div().Class("qr-seg").Attr("role", ariaRole("group")).Attr("aria-label", "Row limit")
	for _, n := range rowLimits() {
		n := n
		cls := "qr-seg-item"
		if t.rowLimit == n {
			cls += " qr-seg-on"
		}
		seg.Children(ui.Div().ID(fmt.Sprintf("ed-limit-%d", n)).Class(cls).Text(limitLabel(n)).
			Attr("role", "button").OnClick(func(ctx context.Context) error {
			t.rowLimit = n
			return a.redraw(ctx)
		}))
	}
	tb.Children(seg)
	return tb
}

// buttonIcon names the icon shown before a toolbar button's label.
func buttonIcon(id string) string {
	switch id {
	case "tbl-freeze", "rs-freeze":
		return "lock"
	case "ed-run", "ed-run-sel", "empty-run":
		return "play"
	case "ed-explain":
		return "search"
	case "ed-analyze", "slow-open-0":
		return "zap"
	case "ed-format":
		return "align-left"
	case "ed-snippet":
		return "save"
	case "ed-cancel":
		return "square"
	case "rs-export", "tbl-export":
		return "download"
	case "tbl-import":
		return "upload"
	case "rs-inspector", "tbl-inspector":
		return "panel-right"
	case "tbl-refresh", "mon-refresh":
		return "refresh"
	case "tbl-add":
		return "plus"
	case "tbl-apply", "cs-commit", "insp-apply":
		return "check"
	case "cs-discard":
		return "x"
	case "cs-preview":
		return "eye"
	case "mon-freeze":
		return "pause"
	case "insp-copy", "ddl-copy":
		return "copy"
	case "tbl-prev":
		return "chevron-left"
	}
	return ""
}

func (a *App) tbButton(id, label, class string, disabled bool, fn func(context.Context) error) *ui.Element {
	b := ui.Button("").ID(id).Class(class).Attr("type", "button").Attr("aria-label", label)
	if ic := buttonIcon(id); ic != "" {
		tone := toneText
		switch {
		case strings.Contains(class, "qr-btn-primary"):
			tone = toneOnBrand
		case strings.Contains(class, "qr-btn-danger"):
			tone = toneErr
		}
		b.Children(a.icon(ic, tone, iconSmall))
	}
	b.Children(ui.Span().Text(label))
	if disabled {
		b.Class(class+" qr-btn-off").Attr("aria-disabled", "true")
		return b
	}
	return b.OnClick(func(ctx context.Context) error {
		a.ac.open = false // a toolbar action is a blur for the completion popup
		return fn(ctx)
	})
}

// ---- editor input ----

// editorFocused reports whether typing belongs to the SQL editor.
func (a *App) editorFocused() bool {
	t := a.activeTabPtr()
	return a.focus == "editor" && t != nil && t.kind == tabQuery && t.ed != nil &&
		a.modal == "" && a.page.Form().FocusID() == ""
}

func (a *App) editorType(ctx context.Context, text string) error {
	t := a.activeTabPtr()
	t.ed.typeText(text)
	t.dirty = true
	t.ed.ensureVisible(a.editorRows())
	t.ed.ensureVisibleCol(a.editorViewCols())
	a.updateAutocomplete(true)
	return a.redraw(ctx)
}

func (a *App) editorBackspace(ctx context.Context) error {
	t := a.activeTabPtr()
	t.ed.backspace()
	t.dirty = true
	t.ed.ensureVisible(a.editorRows())
	a.updateAutocomplete(false)
	return a.redraw(ctx)
}

func (a *App) editorKey(ctx context.Context, key string, shift bool) (bool, error) {
	t := a.activeTabPtr()
	ed := t.ed
	a.syncWrap(t)
	switch key {
	case "enter":
		a.suppressSubmit = true
		ed.insert("\n" + leadingIndent(ed.text, ed.caret))
	case "delete":
		ed.deleteForward()
	case "tab":
		if shift {
			ed.outdent()
		} else {
			ed.indent()
		}
	default:
		if !ed.key(key, shift) {
			return false, nil
		}
		a.ac.open = false
	}
	if key == "enter" || key == "delete" || key == "tab" {
		t.dirty = true
		a.ac.open = false
	}
	if a.ac.open && (ed.caret < a.ac.comp.ReplaceStart || ed.caret > a.ac.comp.ReplaceEnd+1) {
		a.ac.open = false // the caret left the identifier
	}
	ed.ensureVisible(a.editorRows())
	ed.ensureVisibleCol(a.editorViewCols())
	return true, a.redraw(ctx)
}

func leadingIndent(text string, caret int) string {
	ls := strings.LastIndexByte(text[:caret], '\n') + 1
	n := 0
	for ls+n < caret && (text[ls+n] == ' ' || text[ls+n] == '\t') {
		n++
	}
	return text[ls : ls+n]
}

// editorChanged is called after a form control edit; the SQL editor is not a
// form control, so only the filter bar paths land here.
func (a *App) editorChanged(ctx context.Context) error { return nil }

// pressEditor positions the caret from a pointer press inside the editor.
func (a *App) pressEditor(_ ownframe.Box, x, y float64, shift bool) { a.editorPress(x, y, shift) }

// formatEditor uppercases keywords and breaks major clauses onto new lines.
func (a *App) formatEditor(ctx context.Context) error {
	t := a.activeTabPtr()
	if t == nil || t.ed == nil {
		return errNoTab
	}
	out := formatSQL(t.ed.text, a.dialect())
	if out != t.ed.text {
		t.ed.snapshot()
		t.ed.text = out
		t.ed.caret = min(t.ed.caret, len(out))
		t.ed.anchor = t.ed.caret
		t.dirty = true
	}
	return a.redraw(ctx)
}

func isClauseStart(up string) bool {
	switch up {
	case "FROM", "WHERE", "GROUP", "ORDER", "HAVING", "LIMIT", "JOIN", "LEFT", "RIGHT", "INNER", "UNION", "VALUES", "SET":
		return true
	}
	return false
}

func isJoinModifier(w string) bool {
	switch w {
	case "LEFT", "RIGHT", "INNER", "OUTER", "CROSS":
		return true
	}
	return false
}

// formatSQL uppercases keywords and starts each major clause on its own line.
func formatSQL(src string, d sqlkit.Dialect) string {
	toks := sqlkit.Tokenize(src, d)
	var b strings.Builder
	prevWord := ""
	for _, tk := range toks {
		txt := tk.Text(src)
		switch tk.Kind {
		case sqlkit.KindKeyword:
			up := strings.ToUpper(txt)
			if isClauseStart(up) && b.Len() > 0 && prevWord != "BY" && (up != "JOIN" || !isJoinModifier(prevWord)) && (up != "LEFT" || prevWord != "ON") {
				s := strings.TrimRight(b.String(), " ")
				b.Reset()
				b.WriteString(s)
				if !strings.HasSuffix(s, "\n") {
					b.WriteString("\n")
				}
			}
			b.WriteString(up)
			prevWord = up
		case sqlkit.KindSpace:
			b.WriteString(txt)
		default:
			b.WriteString(txt)
			if tk.Kind == sqlkit.KindIdent {
				prevWord = strings.ToUpper(txt)
			}
		}
	}
	return b.String()
}

// ---- autocomplete ----

type acState struct {
	open  bool
	items []sqlkit.Item
	idx   int
	comp  sqlkit.Completion
}

func (a *App) dialect() sqlkit.Dialect {
	if a.cfg.Engine == "postgres" {
		return sqlkit.Postgres
	}
	return sqlkit.SQLite
}

// updateAutocomplete refreshes the popover after typing. typed is true when a
// character was inserted; the popover opens only for a non-empty identifier.
func (a *App) updateAutocomplete(typed bool) {
	t := a.activeTabPtr()
	if t == nil || t.ed == nil {
		return
	}
	if !typed && !a.ac.open {
		return
	}
	comp := sqlkit.Complete(t.ed.text, t.ed.caret, a.sqlSchema, a.dialect())
	if len(comp.Items) == 0 || (comp.Prefix == "" && !a.ac.open && typed) {
		a.ac.open = false
		return
	}
	if typed && comp.Prefix == "" {
		a.ac.open = false
		return
	}
	a.ac.comp = comp
	a.ac.items = comp.Items
	if len(a.ac.items) > autocompleteMax {
		a.ac.items = a.ac.items[:autocompleteMax]
	}
	if !a.ac.open {
		a.ac.idx = 0
	}
	a.ac.idx = min(a.ac.idx, len(a.ac.items)-1)
	a.ac.open = true
}

func (a *App) openAutocomplete(ctx context.Context) error {
	if !a.editorFocused() {
		return nil
	}
	t := a.activeTabPtr()
	comp := sqlkit.Complete(t.ed.text, t.ed.caret, a.sqlSchema, a.dialect())
	a.ac.comp = comp
	a.ac.items = comp.Items
	if len(a.ac.items) > autocompleteMax {
		a.ac.items = a.ac.items[:autocompleteMax]
	}
	a.ac.idx = 0
	a.ac.open = len(a.ac.items) > 0
	return a.redraw(ctx)
}

// autocompleteKey handles popover navigation. It reports whether the key was used.
func (a *App) autocompleteKey(ctx context.Context, key string) (bool, error) {
	switch key {
	case "arrowdown":
		a.ac.idx = (a.ac.idx + 1) % len(a.ac.items)
	case "arrowup":
		a.ac.idx = (a.ac.idx - 1 + len(a.ac.items)) % len(a.ac.items)
	case "enter", "tab":
		a.suppressSubmit = true
		return true, a.acceptCompletion(ctx, a.ac.idx)
	default:
		return false, nil
	}
	return true, a.redraw(ctx)
}

func (a *App) acceptCompletion(ctx context.Context, i int) error {
	t := a.activeTabPtr()
	if t == nil || t.ed == nil || i < 0 || i >= len(a.ac.items) {
		return nil
	}
	c := a.ac.comp
	ed := t.ed
	ed.snapshot()
	ed.text = ed.text[:c.ReplaceStart] + a.ac.items[i].Insert + ed.text[c.ReplaceEnd:]
	ed.caret = c.ReplaceStart + len(a.ac.items[i].Insert)
	ed.anchor = ed.caret
	t.dirty = true
	a.ac.open = false
	return a.redraw(ctx)
}

func (a *App) autocompletePopover(t *workTab) *ui.Element {
	if !a.ac.open || len(a.ac.items) == 0 || t.ed == nil {
		return nil
	}
	l, c := t.ed.rowOf(a.ac.comp.ReplaceStart)
	top := (l-t.ed.scroll+1)*edLineH + acGap
	left := edGutterW + edPadLeft + int(float64(c-t.ed.hcol)*edCharW)
	pop := ui.Div().ID("ac-popover").Class("qr-ac").Attr("role", ariaRole("listbox")).Attr("aria-label", "Completions").
		Position(ui.PositionAbsolute).Top(ui.Px(float64(top))).Left(ui.Px(float64(left)))
	for i, it := range a.ac.items {
		i := i
		cls := "qr-ac-row"
		if i == a.ac.idx {
			cls += " qr-ac-row-active"
		}
		pop.Children(ui.Div().ID(fmt.Sprintf("ac-%d", i)).Class(cls).Attr("role", ariaRole("option")).Children(
			ui.Span().Class("qr-badge qr-badge-brand").Text(it.Kind.Badge()),
			ui.Span().Class("qr-ac-label").Text(it.Label),
			ui.Span().Class("qr-muted qr-ac-detail").Text(it.Detail),
		).OnClick(func(ctx context.Context) error { return a.acceptCompletion(ctx, i) }))
	}
	return pop
}
