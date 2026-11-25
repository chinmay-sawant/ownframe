package main

import (
	"fmt"
	"strings"
)

// xlNode is one AST node.
type xlNode interface{}

type (
	nNum   struct{ v float64 }
	nStr   struct{ s string }
	nBool  struct{ b bool }
	nErr   struct{ e string }
	nEmpty struct{}
	nName  struct{ name string }
	nRef   struct {
		sheet    string
		hasSheet bool
		ref      CellRef
	}
	nRange struct {
		sheet    string
		hasSheet bool
		rng      Range
	}
	nUnary struct {
		op string
		x  xlNode
	}
	nPercent struct{ x xlNode }
	nBinary  struct {
		op   string
		l, r xlNode
	}
	nCall struct {
		name string
		args []xlNode
	}
)

type xlParser struct {
	toks []xlToken
	pos  int
}

// xlParse parses a formula body (without the leading "=").
func xlParse(src string) (xlNode, error) {
	toks, err := xlLex(src)
	if err != nil {
		return nil, err
	}
	p := &xlParser{toks: toks}
	n, err := p.cmp()
	if err != nil {
		return nil, err
	}
	if p.peek().kind != tkEOF {
		return nil, fmt.Errorf("unexpected %q", p.peek().text)
	}
	return n, nil
}

func (p *xlParser) peek() xlToken { return p.toks[p.pos] }
func (p *xlParser) next() xlToken {
	t := p.toks[p.pos]
	if t.kind != tkEOF {
		p.pos++
	}
	return t
}

func (p *xlParser) isOp(set ...string) (string, bool) {
	t := p.peek()
	if t.kind != tkOp {
		return "", false
	}
	for _, s := range set {
		if t.text == s {
			return s, true
		}
	}
	return "", false
}

func (p *xlParser) binLevel(sub func() (xlNode, error), ops ...string) (xlNode, error) {
	l, err := sub()
	if err != nil {
		return nil, err
	}
	for {
		op, ok := p.isOp(ops...)
		if !ok {
			return l, nil
		}
		p.next()
		r, err := sub()
		if err != nil {
			return nil, err
		}
		l = nBinary{op, l, r}
	}
}

func (p *xlParser) cmp() (xlNode, error) {
	return p.binLevel(p.concat, "=", "<>", "<", ">", "<=", ">=")
}
func (p *xlParser) concat() (xlNode, error) { return p.binLevel(p.add, "&") }
func (p *xlParser) add() (xlNode, error)    { return p.binLevel(p.mul, "+", "-") }
func (p *xlParser) mul() (xlNode, error)    { return p.binLevel(p.pow, "*", "/") }
func (p *xlParser) pow() (xlNode, error)    { return p.binLevel(p.unary, "^") }

func (p *xlParser) unary() (xlNode, error) {
	if op, ok := p.isOp("-", "+"); ok {
		p.next()
		x, err := p.unary()
		if err != nil {
			return nil, err
		}
		return nUnary{op, x}, nil
	}
	x, err := p.primary()
	if err != nil {
		return nil, err
	}
	for {
		if _, ok := p.isOp("%"); !ok {
			return x, nil
		}
		p.next()
		x = nPercent{x}
	}
}

func xlCellNode(sheet string, hasSheet bool, a, b xlRefPart, isRange bool) xlNode {
	if a.Row >= MaxRows || a.Col >= MaxCols || b.Row >= MaxRows || b.Col >= MaxCols {
		return nErr{ErrRef}
	}
	ra, rb := CellRef{a.Row, a.Col}, CellRef{b.Row, b.Col}
	if !isRange {
		return nRef{sheet, hasSheet, ra}
	}
	return nRange{sheet, hasSheet, NormRange(ra, rb)}
}

func (p *xlParser) primary() (xlNode, error) {
	t := p.next()
	switch t.kind {
	case tkNum:
		return nNum{t.num}, nil
	case tkStr:
		return nStr{t.text}, nil
	case tkBool:
		return nBool{t.text == "TRUE"}, nil
	case tkErr:
		return nErr{t.text}, nil
	case tkLParen:
		n, err := p.cmp()
		if err != nil {
			return nil, err
		}
		if p.next().kind != tkRParen {
			return nil, fmt.Errorf("missing )")
		}
		return n, nil
	case tkFunc:
		return p.call(t.text)
	case tkSheet:
		r := p.next()
		if r.kind != tkRef {
			return nil, fmt.Errorf("expected reference after %q", t.text)
		}
		return p.refTail(t.text, true, r)
	case tkRef:
		return p.refTail("", false, t)
	case tkIdent:
		// whole-column range A:C
		if p.peek().kind == tkColon && p.pos+1 < len(p.toks) && p.toks[p.pos+1].kind == tkIdent {
			c1, ok1 := xlColLetters(t.text)
			c2, ok2 := xlColLetters(p.toks[p.pos+1].text)
			if ok1 && ok2 {
				p.pos += 2
				if c1 >= MaxCols || c2 >= MaxCols {
					return nErr{ErrRef}, nil
				}
				return nRange{rng: NormRange(CellRef{0, c1}, CellRef{MaxRows - 1, c2})}, nil
			}
		}
		return nName{strings.ToUpper(t.text)}, nil
	}
	if t.kind == tkEOF {
		return nil, fmt.Errorf("unexpected end")
	}
	return nil, fmt.Errorf("unexpected %q", t.text)
}

func xlColLetters(s string) (int, bool) {
	if s == "" || len(s) > 3 {
		return 0, false
	}
	col := 0
	for i := 0; i < len(s); i++ {
		c := s[i] | xlASCIICaseBit
		if c < 'a' || c > 'z' {
			return 0, false
		}
		col = col*26 + int(c-'a') + 1
	}
	return col - 1, true
}

func (p *xlParser) refTail(sheet string, hasSheet bool, first xlToken) (xlNode, error) {
	a, ok := xlParseRefPart(first.text)
	if !ok {
		return nil, fmt.Errorf("bad reference %q", first.text)
	}
	if p.peek().kind == tkColon {
		save := p.pos
		p.next()
		if p.peek().kind == tkSheet && p.pos+1 < len(p.toks) {
			p.next()
		}
		if r := p.peek(); r.kind == tkRef {
			p.next()
			b, ok := xlParseRefPart(r.text)
			if !ok {
				return nil, fmt.Errorf("bad reference %q", r.text)
			}
			return xlCellNode(sheet, hasSheet, a, b, true), nil
		}
		p.pos = save
		return nil, fmt.Errorf("incomplete range")
	}
	return xlCellNode(sheet, hasSheet, a, a, false), nil
}

func (p *xlParser) call(name string) (xlNode, error) {
	if p.next().kind != tkLParen {
		return nil, fmt.Errorf("expected ( after %s", name)
	}
	var args []xlNode
	if p.peek().kind == tkRParen {
		p.next()
		return nCall{name, nil}, nil
	}
	for {
		if k := p.peek().kind; k == tkComma || k == tkRParen {
			args = append(args, nEmpty{})
		} else {
			a, err := p.cmp()
			if err != nil {
				return nil, err
			}
			args = append(args, a)
		}
		switch t := p.next(); t.kind {
		case tkComma:
		case tkRParen:
			return nCall{name, args}, nil
		default:
			return nil, fmt.Errorf("expected , or ) in %s", name)
		}
	}
}
