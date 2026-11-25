package main

import (
	"fmt"
	"strconv"
	"strings"
)

type xlTokKind int

const (
	tkEOF xlTokKind = iota
	tkNum
	tkStr
	tkBool
	tkRef   // A1, $A$1
	tkSheet // Sheet2! or 'Sales Q1'! (text is the unquoted name)
	tkIdent // a bare name
	tkFunc  // name followed by "("
	tkOp
	tkComma
	tkLParen
	tkRParen
	tkColon
	tkErr // #REF! literal
)

// xlTwoCharOp is the width of <=, <> and >=.
const xlTwoCharOp = 2

type xlToken struct {
	kind       xlTokKind
	text       string
	num        float64
	start, end int
}

func xlIsWordByte(c byte) bool {
	return c == '_' || c == '.' || c == '$' || (c >= '0' && c <= '9') || (c|0x20 >= 'a' && c|0x20 <= 'z') || c >= 0x80
}

// xlLex tokenizes a formula body (without the leading "=").
func xlLex(src string) ([]xlToken, error) {
	var toks []xlToken
	i := 0
	for i < len(src) {
		c := src[i]
		switch {
		case c == ' ' || c == '\t':
			i++
		case c >= '0' && c <= '9' || (c == '.' && i+1 < len(src) && src[i+1] >= '0' && src[i+1] <= '9'):
			j := i
			for j < len(src) && (src[j] >= '0' && src[j] <= '9' || src[j] == '.') {
				j++
			}
			if j < len(src) && (src[j] == 'e' || src[j] == 'E') {
				k := j + 1
				if k < len(src) && (src[k] == '+' || src[k] == '-') {
					k++
				}
				if k < len(src) && src[k] >= '0' && src[k] <= '9' {
					for k < len(src) && src[k] >= '0' && src[k] <= '9' {
						k++
					}
					j = k
				}
			}
			f, err := strconv.ParseFloat(src[i:j], 64)
			if err != nil {
				return nil, fmt.Errorf("bad number %q", src[i:j])
			}
			toks = append(toks, xlToken{kind: tkNum, text: src[i:j], num: f, start: i, end: j})
			i = j
		case c == '"':
			var sb strings.Builder
			j := i + 1
			closed := false
			for j < len(src) {
				if src[j] == '"' {
					if j+1 < len(src) && src[j+1] == '"' {
						sb.WriteByte('"')
						j += 2
						continue
					}
					closed = true
					j++
					break
				}
				sb.WriteByte(src[j])
				j++
			}
			if !closed {
				return nil, fmt.Errorf("unterminated string")
			}
			toks = append(toks, xlToken{kind: tkStr, text: sb.String(), start: i, end: j})
			i = j
		case c == '\'':
			var sb strings.Builder
			j := i + 1
			closed := false
			for j < len(src) {
				if src[j] == '\'' {
					if j+1 < len(src) && src[j+1] == '\'' {
						sb.WriteByte('\'')
						j += 2
						continue
					}
					closed = true
					j++
					break
				}
				sb.WriteByte(src[j])
				j++
			}
			if !closed || j >= len(src) || src[j] != '!' {
				return nil, fmt.Errorf("bad quoted sheet name")
			}
			toks = append(toks, xlToken{kind: tkSheet, text: sb.String(), start: i, end: j + 1})
			i = j + 1
		case c == '#':
			j := i + 1
			for j < len(src) && (src[j] == '/' || src[j] == '!' || src[j] == '?' || src[j]|0x20 >= 'a' && src[j]|0x20 <= 'z' || src[j] >= '0' && src[j] <= '9') {
				j++
				if src[j-1] == '!' || src[j-1] == '?' {
					break
				}
			}
			toks = append(toks, xlToken{kind: tkErr, text: strings.ToUpper(src[i:j]), start: i, end: j})
			i = j
		case xlIsWordByte(c):
			j := i
			for j < len(src) && xlIsWordByte(src[j]) {
				j++
			}
			word := src[i:j]
			if j < len(src) && src[j] == '!' {
				toks = append(toks, xlToken{kind: tkSheet, text: word, start: i, end: j + 1})
				i = j + 1
				continue
			}
			k := j
			for k < len(src) && src[k] == ' ' {
				k++
			}
			tok := xlToken{text: word, start: i, end: j}
			switch up := strings.ToUpper(word); {
			case k < len(src) && src[k] == '(':
				tok.kind, tok.text = tkFunc, up
			case up == "TRUE" || up == "FALSE":
				tok.kind, tok.text = tkBool, up
			default:
				if _, ok := xlParseRefPart(word); ok {
					tok.kind = tkRef
				} else {
					tok.kind = tkIdent
				}
			}
			toks = append(toks, tok)
			i = j
		default:
			tok := xlToken{start: i, end: i + 1, text: string(c)}
			switch c {
			case '+', '-', '*', '/', '^', '&', '=', '%':
				tok.kind = tkOp
			case '<':
				tok.kind = tkOp
				if i+1 < len(src) && (src[i+1] == '=' || src[i+1] == '>') {
					tok.text, tok.end = src[i:i+xlTwoCharOp], i+xlTwoCharOp
				}
			case '>':
				tok.kind = tkOp
				if i+1 < len(src) && src[i+1] == '=' {
					tok.text, tok.end = ">=", i+xlTwoCharOp
				}
			case ',':
				tok.kind = tkComma
			case '(':
				tok.kind = tkLParen
			case ')':
				tok.kind = tkRParen
			case ':':
				tok.kind = tkColon
			default:
				return nil, fmt.Errorf("unexpected %q at %d", c, i)
			}
			toks = append(toks, tok)
			i = tok.end
		}
	}
	toks = append(toks, xlToken{kind: tkEOF, start: len(src), end: len(src)})
	return toks, nil
}

// xlRefSpan is a reference (optionally sheet-qualified, optionally a range) in formula text.
type xlRefSpan struct {
	from, to, refFrom int // from..to is the full text; refFrom skips the sheet prefix
	sheet             string
	hasSheet          bool
	a, b              xlRefPart
	isRange           bool
}

// xlScanRefs finds reference spans in formula text (without "="), tolerating
// incomplete formulas.
func xlScanRefs(body string) []xlRefSpan {
	toks, err := xlLex(body)
	if err != nil {
		return nil
	}
	var out []xlRefSpan
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		var sp xlRefSpan
		switch {
		case t.kind == tkSheet && i+1 < len(toks) && toks[i+1].kind == tkRef:
			sp = xlRefSpan{from: t.start, refFrom: toks[i+1].start, sheet: t.text, hasSheet: true}
			i++
		case t.kind == tkRef:
			sp = xlRefSpan{from: t.start, refFrom: t.start}
		default:
			continue
		}
		p, ok := xlParseRefPart(toks[i].text)
		if !ok {
			continue
		}
		sp.a, sp.b, sp.to = p, p, toks[i].end
		if i+2 < len(toks) && toks[i+1].kind == tkColon && toks[i+2].kind == tkRef {
			if q, ok := xlParseRefPart(toks[i+2].text); ok {
				sp.b, sp.isRange, sp.to = q, true, toks[i+2].end
				i += 2
			}
		}
		out = append(out, sp)
	}
	return out
}

// RefToken is a reference located in formula text, for highlighting.
type RefToken struct {
	Start, End int // byte offsets in the full formula string (including "=")
	Range      Range
	Sheet      string // "" = current sheet
}

// FormulaRefTokens lists references in order of appearance with text offsets.
func FormulaRefTokens(formula string) []RefToken {
	body, ok := strings.CutPrefix(formula, "=")
	if !ok {
		return nil
	}
	var out []RefToken
	for _, sp := range xlScanRefs(body) {
		a := CellRef{sp.a.Row, sp.a.Col}
		b := CellRef{sp.b.Row, sp.b.Col}
		if a.Row >= MaxRows || b.Row >= MaxRows || a.Col >= MaxCols || b.Col >= MaxCols {
			continue
		}
		out = append(out, RefToken{Start: sp.from + 1, End: sp.to + 1, Range: NormRange(a, b), Sheet: sp.sheet})
	}
	return out
}

// FormulaRefs returns the ranges referenced on the current sheet, in order of
// appearance (cross-sheet references are skipped).
func FormulaRefs(formula string) []Range {
	var out []Range
	for _, t := range FormulaRefTokens(formula) {
		if t.Sheet == "" {
			out = append(out, t.Range)
		}
	}
	return out
}

// CycleRefAbs rotates a reference through $A$1, A$1, $A1, A1 (the F4 order).
func CycleRefAbs(ref string) string {
	p, ok := xlParseRefPart(ref)
	if !ok {
		return ref
	}
	switch {
	case !p.AbsCol && !p.AbsR:
		p.AbsCol, p.AbsR = true, true
	case p.AbsCol && p.AbsR:
		p.AbsCol, p.AbsR = false, true
	case !p.AbsCol && p.AbsR:
		p.AbsCol, p.AbsR = true, false
	default:
		p.AbsCol, p.AbsR = false, false
	}
	return p.text()
}

// ShiftFormula moves relative references by (dCol, dRow); $-locked parts stay.
// References pushed off the grid become #REF!. Non-formulas are returned as-is.
func ShiftFormula(raw string, dCol, dRow int) string {
	body, ok := strings.CutPrefix(raw, "=")
	if !ok || (dCol == 0 && dRow == 0) {
		return raw
	}
	shift := func(p xlRefPart) (xlRefPart, bool) {
		if !p.AbsCol {
			p.Col += dCol
		}
		if !p.AbsR {
			p.Row += dRow
		}
		return p, p.Col >= 0 && p.Col < MaxCols && p.Row >= 0 && p.Row < MaxRows
	}
	return "=" + xlRewriteSpans(body, func(sp xlRefSpan) (string, bool) {
		a, ok1 := shift(sp.a)
		b, ok2 := shift(sp.b)
		if !ok1 || !ok2 {
			return ErrRef, true
		}
		return xlSpanText(sp, a, b), true
	})
}

func xlSpanText(sp xlRefSpan, a, b xlRefPart) string {
	s := a.text()
	if sp.isRange {
		s += ":" + b.text()
	}
	return s
}

// xlRewriteSpans replaces each reference span using fn; the sheet prefix is kept
// unless fn returns the #REF! error text.
func xlRewriteSpans(body string, fn func(sp xlRefSpan) (string, bool)) string {
	spans := xlScanRefs(body)
	if len(spans) == 0 {
		return body
	}
	var sb strings.Builder
	last := 0
	for _, sp := range spans {
		sb.WriteString(body[last:sp.from])
		rep, ok := fn(sp)
		switch {
		case !ok:
			sb.WriteString(body[sp.from:sp.to])
		case rep == ErrRef:
			sb.WriteString(ErrRef)
		default:
			sb.WriteString(body[sp.from:sp.refFrom])
			sb.WriteString(rep)
		}
		last = sp.to
	}
	sb.WriteString(body[last:])
	return sb.String()
}

// xlQuoteSheet renders a sheet name for use in a reference prefix.
func xlQuoteSheet(name string) string {
	plain := name != ""
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c != '_' && (c < '0' || c > '9') && (c|xlASCIICaseBit < 'a' || c|xlASCIICaseBit > 'z') {
			plain = false
		}
	}
	if _, isRef := xlParseRefPart(name); isRef || (plain && name[0] >= '0' && name[0] <= '9') {
		plain = false
	}
	if plain {
		return name
	}
	return "'" + strings.ReplaceAll(name, "'", "''") + "'"
}
