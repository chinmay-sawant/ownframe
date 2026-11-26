package sqlkit

import "strings"

// Span is a half-open byte range [Start, End) of one statement. It excludes
// leading whitespace and the terminating semicolon.
type Span struct{ Start, End int }

// Text returns the statement text within src.
func (s Span) Text(src string) string { return src[s.Start:s.End] }

// chunk is a statement plus the byte range it "owns" for caret lookup.
type chunk struct {
	span Span
	// own is the end of the owned range: the byte after the semicolon, or len(src).
	own int
}

// Split returns every non-empty statement in src. Semicolons inside strings,
// quoted identifiers, comments, and SQLite CREATE TRIGGER bodies do not split.
// Chunks holding only comments are dropped.
func Split(src string, d Dialect) []Span {
	cs := chunks(src, d)
	out := make([]Span, len(cs))
	for i, c := range cs {
		out[i] = c.span
	}
	return out
}

func chunks(src string, d Dialect) []chunk {
	toks := Tokenize(src, d)
	var out []chunk
	first, last := -1, -1 // first and last significant (non-space) token of the chunk
	hasCode := false      // chunk has a non-comment token
	var w triggerWatch
	flush := func(own int) {
		if first >= 0 && hasCode {
			out = append(out, chunk{span: Span{toks[first].Start, toks[last].End}, own: own})
		}
		first, last, hasCode = -1, -1, false
		w = triggerWatch{}
	}
	for i, t := range toks {
		if t.Kind == KindSpace {
			continue
		}
		if t.Kind == KindOperator && src[t.Start:t.End] == ";" && !w.inBody() {
			flush(t.End)
			continue
		}
		if first < 0 {
			first = i
		}
		last = i
		if t.Kind != KindComment {
			hasCode = true
			if d == SQLite && (t.Kind == KindKeyword || t.Kind == KindIdent) {
				w.word(strings.ToUpper(src[t.Start:t.End]))
			}
		}
	}
	// Trailing statement without a semicolon owns the rest of the buffer.
	flush(len(src))
	// Comment-only chunks were dropped; make the previous chunk own up to the
	// next one so every caret maps somewhere sensible.
	return out
}

// triggerWatch tracks CREATE TRIGGER ... BEGIN ... END so inner semicolons do
// not split the statement (SQLite only).
type triggerWatch struct {
	words     int
	create    bool
	trigger   bool
	begun     bool
	caseDepth int
	closed    bool
}

func (w *triggerWatch) word(u string) {
	w.words++
	switch {
	case w.words == 1:
		w.create = u == "CREATE"
	case w.create && !w.begun && u == "TRIGGER":
		w.trigger = true
	case w.trigger && !w.begun && u == "BEGIN":
		w.begun = true
	case w.begun && !w.closed:
		switch u {
		case "CASE":
			w.caseDepth++
		case "END":
			if w.caseDepth > 0 {
				w.caseDepth--
			} else {
				w.closed = true
			}
		}
	}
}

func (w *triggerWatch) inBody() bool { return w.begun && !w.closed }

// StatementAt returns the statement the caret (a byte offset) belongs to.
// A caret directly after a statement's semicolon, or anywhere inside the
// statement, selects it; a caret in the gap between statements selects the next
// statement, or the last one when none follows. ok is false for an empty script.
func StatementAt(src string, caret int, d Dialect) (Span, bool) {
	cs := chunks(src, d)
	if len(cs) == 0 {
		return Span{}, false
	}
	caret = max(0, min(caret, len(src)))
	for _, c := range cs {
		if caret <= c.own {
			return c.span, true
		}
	}
	return cs[len(cs)-1].span, true
}

// Target is what a run action should execute.
type Target struct {
	SQL  string
	Span Span
}

// RunTarget picks the SQL for "Run": the non-blank selection [selStart, selEnd)
// when present, otherwise the statement at the caret. ok is false when there is
// nothing to run.
func RunTarget(src string, caret, selStart, selEnd int, d Dialect) (Target, bool) {
	if selStart > selEnd {
		selStart, selEnd = selEnd, selStart
	}
	selStart, selEnd = max(0, selStart), min(len(src), selEnd)
	if sel := src[selStart:selEnd]; selEnd > selStart && strings.TrimSpace(sel) != "" {
		return Target{SQL: strings.TrimSpace(sel), Span: Span{selStart, selEnd}}, true
	}
	sp, ok := StatementAt(src, caret, d)
	if !ok {
		return Target{}, false
	}
	return Target{SQL: sp.Text(src), Span: sp}, true
}

// Script returns every statement text in order for "Run All".
func Script(src string, d Dialect) []string {
	spans := Split(src, d)
	out := make([]string, len(spans))
	for i, s := range spans {
		out[i] = s.Text(src)
	}
	return out
}

// LineCol converts a byte offset into a 1-based line and column (in runes).
func LineCol(src string, off int) (line, col int) {
	off = max(0, min(off, len(src)))
	line, col = 1, 1
	for _, r := range src[:off] {
		if r == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

// OffsetOfLine returns the byte offset of the start of 1-based line n, clamped.
func OffsetOfLine(src string, n int) int {
	off := 0
	for ; n > 1; n-- {
		i := strings.IndexByte(src[off:], '\n')
		if i < 0 {
			return len(src)
		}
		off += i + 1
	}
	return off
}
