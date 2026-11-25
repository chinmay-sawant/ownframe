package main

import (
	"math"
	"strings"
)

// xlOperand is an evaluated node: a scalar, or a reference to a rectangle.
type xlOperand struct {
	v     Value
	sh    *Sheet
	rng   Range
	isRef bool
}

type xlCtx struct {
	wb *Workbook
	sh *Sheet
	at CellRef
}

func xlScalarOp(v Value) xlOperand { return xlOperand{v: v} }
func xlErrOp(e string) xlOperand   { return xlOperand{v: xlErrV(e)} }

func (c *xlCtx) resolve(name string, has bool) *Sheet {
	if !has {
		return c.sh
	}
	return c.wb.sheetByName(name)
}

// scalar reduces an operand to one value (a multi-cell range is #VALUE!).
func (c *xlCtx) scalar(o xlOperand) Value {
	if !o.isRef {
		return o.v
	}
	if o.rng.Min != o.rng.Max {
		// implicit intersection on the formula's own row/column
		if o.rng.Min.Col == o.rng.Max.Col && c.at.Row >= o.rng.Min.Row && c.at.Row <= o.rng.Max.Row {
			return o.sh.cellVal(CellRef{c.at.Row, o.rng.Min.Col})
		}
		if o.rng.Min.Row == o.rng.Max.Row && c.at.Col >= o.rng.Min.Col && c.at.Col <= o.rng.Max.Col {
			return o.sh.cellVal(CellRef{o.rng.Min.Row, c.at.Col})
		}
		return xlErrV(ErrValue)
	}
	return o.sh.cellVal(o.rng.Min)
}

func (c *xlCtx) val(n xlNode) Value { return c.scalar(c.eval(n)) }

// cells lists the values of a range operand in row-major order, clipped to the
// used area; a scalar yields itself.
func (c *xlCtx) cells(o xlOperand) []Value {
	if !o.isRef {
		return []Value{o.v}
	}
	r := o.rng
	u := o.sh.UsedRange()
	r.Max.Row, r.Max.Col = min(r.Max.Row, u.Max.Row), min(r.Max.Col, u.Max.Col)
	var out []Value
	for row := r.Min.Row; row <= r.Max.Row; row++ {
		for col := r.Min.Col; col <= r.Max.Col; col++ {
			out = append(out, o.sh.cellVal(CellRef{row, col}))
		}
	}
	return out
}

func (c *xlCtx) eval(n xlNode) xlOperand {
	switch t := n.(type) {
	case nNum:
		return xlScalarOp(xlNumV(t.v))
	case nStr:
		return xlScalarOp(xlStrV(t.s))
	case nBool:
		return xlScalarOp(xlBoolV(t.b))
	case nErr:
		return xlErrOp(t.e)
	case nEmpty:
		return xlScalarOp(Value{})
	case nName:
		return xlErrOp(ErrName)
	case nRef:
		sh := c.resolve(t.sheet, t.hasSheet)
		if sh == nil {
			return xlErrOp(ErrRef)
		}
		return xlOperand{sh: sh, rng: Range{t.ref, t.ref}, isRef: true}
	case nRange:
		sh := c.resolve(t.sheet, t.hasSheet)
		if sh == nil {
			return xlErrOp(ErrRef)
		}
		return xlOperand{sh: sh, rng: t.rng, isRef: true}
	case nUnary:
		v := c.val(t.x)
		if v.isErr() {
			return xlScalarOp(v)
		}
		if t.op == "+" {
			return xlScalarOp(v)
		}
		f, e := xlNumOf(v)
		if e != "" {
			return xlErrOp(e)
		}
		return xlScalarOp(xlNumV(-f))
	case nPercent:
		v := c.val(t.x)
		f, e := xlNumOf(v)
		if e != "" {
			return xlErrOp(e)
		}
		return xlScalarOp(xlNumV(f / xlPercentScale))
	case nBinary:
		return xlScalarOp(c.binary(t))
	case nCall:
		return c.call(t)
	}
	return xlErrOp(ErrValue)
}

// xlNumOf coerces a value to a number; e is "" on success.
func xlNumOf(v Value) (f float64, e string) {
	switch v.Kind {
	case KindEmpty:
		return 0, ""
	case KindNumber:
		return v.Num, ""
	case KindBool:
		if v.Bool {
			return 1, ""
		}
		return 0, ""
	case KindError:
		return 0, v.Err
	}
	if n, ok := xlParseNumber(strings.TrimSpace(v.Str)); ok {
		return n, ""
	}
	if strings.TrimSpace(v.Str) == "" {
		return 0, ErrValue
	}
	return 0, ErrValue
}

// xlBoolOf coerces a value to a boolean.
func xlBoolOf(v Value) (b bool, e string) {
	switch v.Kind {
	case KindEmpty:
		return false, ""
	case KindBool:
		return v.Bool, ""
	case KindNumber:
		return v.Num != 0, ""
	case KindError:
		return false, v.Err
	}
	switch strings.ToUpper(strings.TrimSpace(v.Str)) {
	case "TRUE":
		return true, ""
	case "FALSE":
		return false, ""
	}
	return false, ErrValue
}

func xlFinite(f float64) Value {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return xlErrV(ErrNum)
	}
	return xlNumV(f)
}

// xlCompare orders two values: numbers < text < booleans; empty matches the other side.
func xlCompare(a, b Value) int {
	if a.Kind == KindEmpty {
		switch b.Kind {
		case KindString:
			a = xlStrV("")
		case KindBool:
			a = xlBoolV(false)
		default:
			a = xlNumV(0)
		}
	}
	if b.Kind == KindEmpty {
		switch a.Kind {
		case KindString:
			b = xlStrV("")
		case KindBool:
			b = xlBoolV(false)
		default:
			b = xlNumV(0)
		}
	}
	ra, rb := xlCmpRank(a), xlCmpRank(b)
	if ra != rb {
		if ra < rb {
			return -1
		}
		return 1
	}
	return xlCompareScalar(a, b)
}

func xlCmpRank(v Value) int {
	switch v.Kind {
	case KindNumber:
		return xlRankNumber
	case KindString:
		return xlRankString
	case KindBool:
		return xlRankBool
	}
	return xlRankError
}

// Sort ranks: numbers, text, bools, errors, then empty cells.
const (
	xlRankNumber = iota
	xlRankString
	xlRankBool
	xlRankError
	xlRankEmpty
)

func (c *xlCtx) binary(b nBinary) Value {
	l, r := c.val(b.l), c.val(b.r)
	if l.isErr() {
		return l
	}
	if r.isErr() {
		return r
	}
	switch b.op {
	case "&":
		return xlStrV(l.String() + r.String())
	case "=", "<>", "<", ">", "<=", ">=":
		cmp := xlCompare(l, r)
		var ok bool
		switch b.op {
		case "=":
			ok = cmp == 0
		case "<>":
			ok = cmp != 0
		case "<":
			ok = cmp < 0
		case ">":
			ok = cmp > 0
		case "<=":
			ok = cmp <= 0
		default:
			ok = cmp >= 0
		}
		return xlBoolV(ok)
	}
	x, e := xlNumOf(l)
	if e != "" {
		return xlErrV(e)
	}
	y, e := xlNumOf(r)
	if e != "" {
		return xlErrV(e)
	}
	switch b.op {
	case "+":
		return xlFinite(x + y)
	case "-":
		return xlFinite(x - y)
	case "*":
		return xlFinite(x * y)
	case "/":
		if y == 0 {
			return xlErrV(ErrDiv0)
		}
		return xlFinite(x / y)
	case "^":
		if x == 0 && y == 0 {
			return xlErrV(ErrNum)
		}
		if x == 0 && y < 0 {
			return xlErrV(ErrDiv0)
		}
		return xlFinite(math.Pow(x, y))
	}
	return xlErrV(ErrValue)
}

// call dispatches a function node.
func (c *xlCtx) call(n nCall) xlOperand {
	def, ok := c.wb.funcTable()[n.name]
	if !ok {
		return xlErrOp(ErrName)
	}
	if len(n.args) < def.min || (def.max >= 0 && len(n.args) > def.max) {
		return xlErrOp(ErrValue)
	}
	if def.lazy != nil {
		return xlScalarOp(def.lazy(c, n.args))
	}
	args := make([]xlOperand, len(n.args))
	for i, a := range n.args {
		args[i] = c.eval(a)
		if !def.tol {
			if !args[i].isRef && args[i].v.isErr() {
				return args[i]
			}
			if args[i].isRef && args[i].rng.Min == args[i].rng.Max {
				if v := args[i].sh.cellVal(args[i].rng.Min); v.isErr() {
					return xlScalarOp(v)
				}
			}
		}
	}
	return xlScalarOp(def.fn(c, args))
}
