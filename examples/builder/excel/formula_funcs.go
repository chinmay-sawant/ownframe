package main

import (
	"math"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type xlFuncDef struct {
	min, max int // max < 0 = variadic
	fn       func(c *xlCtx, a []xlOperand) Value
	lazy     func(c *xlCtx, n []xlNode) Value
	tol      bool // receives error arguments instead of propagating them
}

type xlRegistry map[string]xlFuncDef

func (r xlRegistry) reg(name string, min, max int, fn func(c *xlCtx, a []xlOperand) Value) {
	r[name] = xlFuncDef{min: min, max: max, fn: fn}
}

func (r xlRegistry) regTol(name string, min, max int, fn func(c *xlCtx, a []xlOperand) Value) {
	r[name] = xlFuncDef{min: min, max: max, fn: fn, tol: true}
}

func (r xlRegistry) regLazy(name string, min, max int, fn func(c *xlCtx, n []xlNode) Value) {
	r[name] = xlFuncDef{min: min, max: max, lazy: fn}
}

// FunctionSignatures lists the supported functions for autocomplete.
func FunctionSignatures() map[string]string {
	sigs := xlSignatures()
	out := make(map[string]string, len(sigs))
	for k, v := range sigs {
		out[k] = v
	}
	return out
}

// FunctionNames returns the supported function names, sorted.
func FunctionNames() []string {
	table := xlBuildFuncs()
	names := make([]string, 0, len(table))
	for k := range table {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func xlSignatures() map[string]string {
	return map[string]string{
		"SUM": "SUM(number1, [number2], ...)", "AVERAGE": "AVERAGE(number1, [number2], ...)",
		"MIN": "MIN(number1, [number2], ...)", "MAX": "MAX(number1, [number2], ...)",
		"COUNT": "COUNT(value1, [value2], ...)", "COUNTA": "COUNTA(value1, [value2], ...)",
		"COUNTBLANK": "COUNTBLANK(range)", "ROUND": "ROUND(number, digits)",
		"ROUNDUP": "ROUNDUP(number, digits)", "ROUNDDOWN": "ROUNDDOWN(number, digits)",
		"ABS": "ABS(number)", "MOD": "MOD(number, divisor)", "POWER": "POWER(number, power)",
		"SQRT": "SQRT(number)", "INT": "INT(number)", "SUMIF": "SUMIF(range, criteria, [sum_range])",
		"COUNTIF": "COUNTIF(range, criteria)", "SUMIFS": "SUMIFS(sum_range, criteria_range1, criteria1, ...)",
		"COUNTIFS": "COUNTIFS(criteria_range1, criteria1, ...)", "AVERAGEIF": "AVERAGEIF(range, criteria, [average_range])",
		"MEDIAN": "MEDIAN(number1, [number2], ...)", "PRODUCT": "PRODUCT(number1, [number2], ...)",
		"IF": "IF(condition, value_if_true, [value_if_false])", "IFS": "IFS(condition1, value1, ...)",
		"AND": "AND(logical1, [logical2], ...)", "OR": "OR(logical1, [logical2], ...)",
		"NOT": "NOT(logical)", "XOR": "XOR(logical1, [logical2], ...)",
		"IFERROR": "IFERROR(value, value_if_error)", "IFNA": "IFNA(value, value_if_na)",
		"ISBLANK": "ISBLANK(value)", "ISNUMBER": "ISNUMBER(value)", "ISTEXT": "ISTEXT(value)",
		"ISERROR": "ISERROR(value)", "ISNA": "ISNA(value)",
		"VLOOKUP": "VLOOKUP(lookup_value, table, col_index, [approximate])",
		"HLOOKUP": "HLOOKUP(lookup_value, table, row_index, [approximate])",
		"XLOOKUP": "XLOOKUP(lookup_value, lookup_array, return_array, [if_not_found], [match_mode])",
		"INDEX":   "INDEX(range, row, [column])", "MATCH": "MATCH(lookup_value, range, [match_type])",
		"ROW": "ROW([reference])", "COLUMN": "COLUMN([reference])", "ROWS": "ROWS(range)", "COLUMNS": "COLUMNS(range)",
		"CONCAT": "CONCAT(text1, [text2], ...)", "CONCATENATE": "CONCATENATE(text1, [text2], ...)",
		"TEXTJOIN": "TEXTJOIN(delimiter, ignore_empty, text1, ...)", "LEFT": "LEFT(text, [count])",
		"RIGHT": "RIGHT(text, [count])", "MID": "MID(text, start, count)", "LEN": "LEN(text)",
		"TRIM": "TRIM(text)", "UPPER": "UPPER(text)", "LOWER": "LOWER(text)", "PROPER": "PROPER(text)",
		"SUBSTITUTE": "SUBSTITUTE(text, old, new, [instance])", "FIND": "FIND(find, within, [start])",
		"SEARCH": "SEARCH(find, within, [start])", "TEXT": "TEXT(value, format)", "REPT": "REPT(text, count)",
		"VALUE": "VALUE(text)", "TODAY": "TODAY()", "NOW": "NOW()", "DATE": "DATE(year, month, day)",
		"YEAR": "YEAR(date)", "MONTH": "MONTH(date)", "DAY": "DAY(date)",
		"DATEDIF": "DATEDIF(start, end, unit)", "WEEKDAY": "WEEKDAY(date, [type])",
		"CEILING": "CEILING(number, significance)", "FLOOR": "FLOOR(number, significance)",
		"LARGE": "LARGE(range, k)", "SMALL": "SMALL(range, k)", "SUMPRODUCT": "SUMPRODUCT(array1, array2, ...)",
		"PI": "PI()", "SIGN": "SIGN(number)", "EXP": "EXP(number)", "LN": "LN(number)", "LOG10": "LOG10(number)",
		"TRUNC": "TRUNC(number, [digits])", "EXACT": "EXACT(text1, text2)", "NA": "NA()",
		"TRUE": "TRUE()", "FALSE": "FALSE()",
	}
}

// Argument counts used when registering functions, and argument positions of
// the optional lookup parameters.
const (
	xlArgs2 = 2
	xlArgs3 = 3
	xlArgs4 = 4
	xlArgs5 = 5

	xlArgLookupApprox = 3 // optional approximate-match flag of VLOOKUP/HLOOKUP
	xlArgMatchMode    = 4 // optional match mode of XLOOKUP
	xlMonthsPerYear   = 12
	xlDaysPerWeek     = 7
	xlMidpointDiv     = 2 // averaging the two middle values

	xlArgThird  = 2 // zero-based index of the third argument
	xlArgFourth = 3 // zero-based index of the fourth argument

	xlWeekdayMondayIs1 = 2 // WEEKDAY type where Monday is 1
)

// ---- argument helpers ----

func xlHas(a []xlOperand, i int) bool {
	return i < len(a) && (a[i].isRef || a[i].v.Kind != KindEmpty)
}

func (c *xlCtx) numAt(a []xlOperand, i int, def float64) (float64, string) {
	if !xlHas(a, i) {
		return def, ""
	}
	return xlNumOf(c.scalar(a[i]))
}

func (c *xlCtx) strAt(a []xlOperand, i int) string {
	if i >= len(a) {
		return ""
	}
	return c.scalar(a[i]).String()
}

// nums gathers numbers for aggregates: range text/bools/empties are ignored,
// scalar arguments are coerced, errors propagate.
func (c *xlCtx) nums(args []xlOperand) ([]float64, string) {
	var out []float64
	for _, a := range args {
		if a.isRef {
			for _, v := range c.cells(a) {
				switch v.Kind {
				case KindError:
					return nil, v.Err
				case KindNumber:
					out = append(out, v.Num)
				}
			}
			continue
		}
		if a.v.Kind == KindEmpty {
			continue
		}
		f, e := xlNumOf(a.v)
		if e != "" {
			return nil, e
		}
		out = append(out, f)
	}
	return out, ""
}

// rect returns the full rectangle of values for a range operand.
func (c *xlCtx) rect(o xlOperand) [][]Value {
	if !o.isRef {
		return [][]Value{{o.v}}
	}
	out := make([][]Value, o.rng.Rows())
	for i := range out {
		out[i] = make([]Value, o.rng.Cols())
		for j := range out[i] {
			out[i][j] = o.sh.cellVal(CellRef{o.rng.Min.Row + i, o.rng.Min.Col + j})
		}
	}
	return out
}

func xlWildMatch(pat, s string) bool {
	p, t := []rune(strings.ToLower(pat)), []rune(strings.ToLower(s))
	var match func(i, j int) bool
	match = func(i, j int) bool {
		for i < len(p) {
			switch p[i] {
			case '*':
				for k := j; k <= len(t); k++ {
					if match(i+1, k) {
						return true
					}
				}
				return false
			case '?':
				if j >= len(t) {
					return false
				}
			case '~':
				if i+1 < len(p) {
					i++
				}
				fallthrough
			default:
				if j >= len(t) || t[j] != p[i] {
					return false
				}
			}
			i++
			j++
		}
		return j == len(t)
	}
	return match(0, 0)
}

// xlCriteria builds a predicate from a COUNTIF-style criterion.
func xlCriteria(crit Value) func(Value) bool {
	switch crit.Kind {
	case KindNumber:
		return func(v Value) bool { return v.Kind == KindNumber && v.Num == crit.Num }
	case KindBool:
		return func(v Value) bool { return v.Kind == KindBool && v.Bool == crit.Bool }
	case KindEmpty:
		return func(v Value) bool { return v.Kind == KindEmpty }
	}
	s := crit.Str
	op := "="
	for _, p := range []string{">=", "<=", "<>", ">", "<", "="} {
		if strings.HasPrefix(s, p) {
			op, s = p, s[len(p):]
			break
		}
	}
	cmpOK := func(c int) bool {
		switch op {
		case ">":
			return c > 0
		case "<":
			return c < 0
		case ">=":
			return c >= 0
		case "<=":
			return c <= 0
		case "<>":
			return c != 0
		}
		return c == 0
	}
	if f, ok := xlParseNumber(s); ok {
		return func(v Value) bool {
			if v.Kind != KindNumber {
				return op == "<>"
			}
			switch {
			case v.Num < f:
				return cmpOK(-1)
			case v.Num > f:
				return cmpOK(1)
			}
			return cmpOK(0)
		}
	}
	if s == "" {
		return func(v Value) bool {
			blank := v.Kind == KindEmpty || (v.Kind == KindString && v.Str == "")
			if op == "<>" {
				return !blank
			}
			return blank
		}
	}
	return func(v Value) bool {
		if v.Kind == KindEmpty {
			return op == "<>"
		}
		txt := v.String()
		if op == "=" || op == "<>" {
			m := xlWildMatch(s, txt)
			return m == (op == "=")
		}
		if v.Kind != KindString {
			return false
		}
		return cmpOK(strings.Compare(strings.ToLower(txt), strings.ToLower(s)))
	}
}

// xlMatching returns (row, col) offsets within the first criteria range where
// all criteria hold.
func (c *xlCtx) matching(a []xlOperand) ([][2]int, string) {
	if len(a) == 0 || len(a)%2 != 0 {
		return nil, ErrValue
	}
	first := a[0]
	if !first.isRef {
		return nil, ErrValue
	}
	rows, cols := first.rng.Rows(), first.rng.Cols()
	type crit struct {
		rect [][]Value
		pred func(Value) bool
	}
	var cs []crit
	for i := 0; i < len(a); i += 2 {
		if !a[i].isRef || a[i].rng.Rows() != rows || a[i].rng.Cols() != cols {
			return nil, ErrValue
		}
		cv := c.scalar(a[i+1])
		if cv.isErr() {
			return nil, cv.Err
		}
		cs = append(cs, crit{c.rect(a[i]), xlCriteria(cv)})
	}
	var out [][2]int
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			ok := true
			for _, k := range cs {
				if !k.pred(k.rect[i][j]) {
					ok = false
					break
				}
			}
			if ok {
				out = append(out, [2]int{i, j})
			}
		}
	}
	return out, ""
}

func (c *xlCtx) sumAt(sum xlOperand, pos [][2]int) (float64, int, string) {
	r := c.rect(sum)
	total, n := 0.0, 0
	for _, p := range pos {
		if p[0] < len(r) && p[1] < len(r[p[0]]) {
			v := r[p[0]][p[1]]
			if v.isErr() {
				return 0, 0, v.Err
			}
			if v.Kind == KindNumber {
				total += v.Num
				n++
			}
		}
	}
	return total, n, ""
}

func xlTruthy(v Value) (bool, string) { return xlBoolOf(v) }

func xlDateSerial(y, m, d int) float64 {
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	return math.Floor(xlTimeToSerial(t))
}

func xlSub(rs []rune, start, n int) string {
	if start < 0 {
		start = 0
	}
	if start > len(rs) {
		start = len(rs)
	}
	end := min(len(rs), start+max(n, 0))
	return string(rs[start:end])
}

func xlBuildFuncs() xlRegistry {
	r := xlRegistry{}
	// ---- math and aggregation ----
	r.reg("SUM", 1, -1, func(c *xlCtx, a []xlOperand) Value {
		ns, e := c.nums(a)
		if e != "" {
			return xlErrV(e)
		}
		s := 0.0
		for _, n := range ns {
			s += n
		}
		return xlFinite(s)
	})
	r.reg("AVERAGE", 1, -1, func(c *xlCtx, a []xlOperand) Value {
		ns, e := c.nums(a)
		if e != "" {
			return xlErrV(e)
		}
		if len(ns) == 0 {
			return xlErrV(ErrDiv0)
		}
		s := 0.0
		for _, n := range ns {
			s += n
		}
		return xlNumV(s / float64(len(ns)))
	})
	for _, name := range []string{"MIN", "MAX"} {
		isMin := name == "MIN"
		r.reg(name, 1, -1, func(c *xlCtx, a []xlOperand) Value {
			ns, e := c.nums(a)
			if e != "" {
				return xlErrV(e)
			}
			if len(ns) == 0 {
				return xlNumV(0)
			}
			m := ns[0]
			for _, n := range ns[1:] {
				if (isMin && n < m) || (!isMin && n > m) {
					m = n
				}
			}
			return xlNumV(m)
		})
	}
	r.reg("PRODUCT", 1, -1, func(c *xlCtx, a []xlOperand) Value {
		ns, e := c.nums(a)
		if e != "" {
			return xlErrV(e)
		}
		p := 1.0
		for _, n := range ns {
			p *= n
		}
		return xlFinite(p)
	})
	r.reg("MEDIAN", 1, -1, func(c *xlCtx, a []xlOperand) Value {
		ns, e := c.nums(a)
		if e != "" {
			return xlErrV(e)
		}
		if len(ns) == 0 {
			return xlErrV(ErrNum)
		}
		sort.Float64s(ns)
		if len(ns)%2 == 1 {
			return xlNumV(ns[len(ns)/2])
		}
		return xlNumV((ns[len(ns)/2-1] + ns[len(ns)/2]) / xlMidpointDiv)
	})
	for _, name := range []string{"LARGE", "SMALL"} {
		large := name == "LARGE"
		r.reg(name, xlArgs2, xlArgs2, func(c *xlCtx, a []xlOperand) Value {
			ns, e := c.nums(a[:1])
			if e != "" {
				return xlErrV(e)
			}
			k, e := c.numAt(a, 1, 1)
			if e != "" {
				return xlErrV(e)
			}
			if int(k) < 1 || int(k) > len(ns) {
				return xlErrV(ErrNum)
			}
			sort.Float64s(ns)
			if large {
				return xlNumV(ns[len(ns)-int(k)])
			}
			return xlNumV(ns[int(k)-1])
		})
	}
	r.regTol("COUNT", 1, -1, func(c *xlCtx, a []xlOperand) Value {
		n := 0
		for _, o := range a {
			if o.isRef {
				for _, v := range c.cells(o) {
					if v.Kind == KindNumber {
						n++
					}
				}
			} else if _, e := xlNumOf(o.v); e == "" && o.v.Kind != KindEmpty {
				n++
			}
		}
		return xlNumV(float64(n))
	})
	r.regTol("COUNTA", 1, -1, func(c *xlCtx, a []xlOperand) Value {
		n := 0
		for _, o := range a {
			for _, v := range c.cells(o) {
				if v.Kind != KindEmpty {
					n++
				}
			}
		}
		return xlNumV(float64(n))
	})
	r.regTol("COUNTBLANK", 1, 1, func(c *xlCtx, a []xlOperand) Value {
		if !a[0].isRef {
			return xlErrV(ErrValue)
		}
		total := a[0].rng.Rows() * a[0].rng.Cols()
		for _, v := range c.cells(a[0]) {
			if v.Kind != KindEmpty && (v.Kind != KindString || v.Str != "") {
				total--
			}
		}
		return xlNumV(float64(total))
	})
	round := func(name string, f func(x float64, d int) float64) {
		r.reg(name, xlArgs2, xlArgs2, func(c *xlCtx, a []xlOperand) Value {
			x, e := c.numAt(a, 0, 0)
			if e != "" {
				return xlErrV(e)
			}
			d, e := c.numAt(a, 1, 0)
			if e != "" {
				return xlErrV(e)
			}
			return xlFinite(f(x, int(d)))
		})
	}
	round("ROUND", xlRound)
	round("ROUNDUP", func(x float64, d int) float64 {
		p := math.Pow(xlDecimalBase, float64(d))
		return math.Copysign(math.Ceil(xlClean15(math.Abs(x)*p))/p, x)
	})
	round("ROUNDDOWN", func(x float64, d int) float64 {
		p := math.Pow(xlDecimalBase, float64(d))
		return math.Copysign(math.Floor(xlClean15(math.Abs(x)*p))/p, x)
	})
	unary := func(name string, f func(x float64) (float64, string)) {
		r.reg(name, 1, 1, func(c *xlCtx, a []xlOperand) Value {
			x, e := c.numAt(a, 0, 0)
			if e != "" {
				return xlErrV(e)
			}
			r, e := f(x)
			if e != "" {
				return xlErrV(e)
			}
			return xlFinite(r)
		})
	}
	unary("ABS", func(x float64) (float64, string) { return math.Abs(x), "" })
	unary("INT", func(x float64) (float64, string) { return math.Floor(x), "" })
	unary("SIGN", func(x float64) (float64, string) {
		switch {
		case x > 0:
			return 1, ""
		case x < 0:
			return -1, ""
		}
		return 0, ""
	})
	unary("SQRT", func(x float64) (float64, string) {
		if x < 0 {
			return 0, ErrNum
		}
		return math.Sqrt(x), ""
	})
	unary("EXP", func(x float64) (float64, string) { return math.Exp(x), "" })
	unary("LN", func(x float64) (float64, string) {
		if x <= 0 {
			return 0, ErrNum
		}
		return math.Log(x), ""
	})
	unary("LOG10", func(x float64) (float64, string) {
		if x <= 0 {
			return 0, ErrNum
		}
		return math.Log10(x), ""
	})
	r.reg("PI", 0, 0, func(c *xlCtx, a []xlOperand) Value { return xlNumV(math.Pi) })
	r.reg("TRUNC", 1, xlArgs2, func(c *xlCtx, a []xlOperand) Value {
		x, e := c.numAt(a, 0, 0)
		if e != "" {
			return xlErrV(e)
		}
		d, _ := c.numAt(a, 1, 0)
		p := math.Pow(xlDecimalBase, d)
		return xlFinite(math.Trunc(x*p) / p)
	})
	r.reg("MOD", xlArgs2, xlArgs2, func(c *xlCtx, a []xlOperand) Value {
		n, e := c.numAt(a, 0, 0)
		if e != "" {
			return xlErrV(e)
		}
		d, e := c.numAt(a, 1, 0)
		if e != "" {
			return xlErrV(e)
		}
		if d == 0 {
			return xlErrV(ErrDiv0)
		}
		return xlFinite(n - d*math.Floor(n/d))
	})
	r.reg("POWER", xlArgs2, xlArgs2, func(c *xlCtx, a []xlOperand) Value {
		n, e := c.numAt(a, 0, 0)
		if e != "" {
			return xlErrV(e)
		}
		p, e := c.numAt(a, 1, 0)
		if e != "" {
			return xlErrV(e)
		}
		return c.binary(nBinary{"^", nNum{n}, nNum{p}})
	})
	for _, name := range []string{"CEILING", "FLOOR"} {
		ceil := name == "CEILING"
		r.reg(name, xlArgs2, xlArgs2, func(c *xlCtx, a []xlOperand) Value {
			n, e := c.numAt(a, 0, 0)
			if e != "" {
				return xlErrV(e)
			}
			s, e := c.numAt(a, 1, 1)
			if e != "" {
				return xlErrV(e)
			}
			if s == 0 {
				return xlNumV(0)
			}
			if ceil {
				return xlFinite(math.Ceil(n/s) * s)
			}
			return xlFinite(math.Floor(n/s) * s)
		})
	}
	r.reg("SUMPRODUCT", 1, -1, func(c *xlCtx, a []xlOperand) Value {
		rects := make([][][]Value, len(a))
		for i, o := range a {
			rects[i] = c.rect(o)
			if len(rects[i]) != len(rects[0]) || len(rects[i][0]) != len(rects[0][0]) {
				return xlErrV(ErrValue)
			}
		}
		total := 0.0
		for i := range rects[0] {
			for j := range rects[0][i] {
				p := 1.0
				for _, r := range rects {
					if v := r[i][j]; v.Kind == KindNumber {
						p *= v.Num
					} else {
						p = 0
					}
				}
				total += p
			}
		}
		return xlFinite(total)
	})
	r.reg("SUMIF", xlArgs2, xlArgs3, func(c *xlCtx, a []xlOperand) Value {
		pos, e := c.matching(a[:2])
		if e != "" {
			return xlErrV(e)
		}
		src := a[0]
		if len(a) == xlArgs3 {
			src = a[2]
		}
		s, _, e := c.sumAt(src, pos)
		if e != "" {
			return xlErrV(e)
		}
		return xlNumV(s)
	})
	r.reg("AVERAGEIF", xlArgs2, xlArgs3, func(c *xlCtx, a []xlOperand) Value {
		pos, e := c.matching(a[:2])
		if e != "" {
			return xlErrV(e)
		}
		src := a[0]
		if len(a) == xlArgs3 {
			src = a[2]
		}
		s, n, e := c.sumAt(src, pos)
		if e != "" {
			return xlErrV(e)
		}
		if n == 0 {
			return xlErrV(ErrDiv0)
		}
		return xlNumV(s / float64(n))
	})
	r.reg("COUNTIF", xlArgs2, xlArgs2, func(c *xlCtx, a []xlOperand) Value {
		pos, e := c.matching(a)
		if e != "" {
			return xlErrV(e)
		}
		return xlNumV(float64(len(pos)))
	})
	r.reg("COUNTIFS", xlArgs2, -1, func(c *xlCtx, a []xlOperand) Value {
		pos, e := c.matching(a)
		if e != "" {
			return xlErrV(e)
		}
		return xlNumV(float64(len(pos)))
	})
	r.reg("SUMIFS", xlArgs3, -1, func(c *xlCtx, a []xlOperand) Value {
		pos, e := c.matching(a[1:])
		if e != "" {
			return xlErrV(e)
		}
		s, _, e := c.sumAt(a[0], pos)
		if e != "" {
			return xlErrV(e)
		}
		return xlNumV(s)
	})

	// ---- logical ----
	r.regLazy("IF", xlArgs2, xlArgs3, func(c *xlCtx, n []xlNode) Value {
		cond := c.val(n[0])
		b, e := xlTruthy(cond)
		if e != "" {
			return xlErrV(e)
		}
		switch {
		case b:
			return c.valOrZero(n[1])
		case len(n) == xlArgs3:
			return c.valOrZero(n[2])
		}
		return xlBoolV(false)
	})
	r.regLazy("IFS", xlArgs2, -1, func(c *xlCtx, n []xlNode) Value {
		if len(n)%2 != 0 {
			return xlErrV(ErrNA)
		}
		for i := 0; i < len(n); i += 2 {
			b, e := xlTruthy(c.val(n[i]))
			if e != "" {
				return xlErrV(e)
			}
			if b {
				return c.valOrZero(n[i+1])
			}
		}
		return xlErrV(ErrNA)
	})
	r.regLazy("IFERROR", xlArgs2, xlArgs2, func(c *xlCtx, n []xlNode) Value {
		if v := c.val(n[0]); !v.isErr() {
			return v
		}
		return c.valOrZero(n[1])
	})
	r.regLazy("IFNA", xlArgs2, xlArgs2, func(c *xlCtx, n []xlNode) Value {
		if v := c.val(n[0]); !v.isErr() || v.Err != ErrNA {
			return v
		}
		return c.valOrZero(n[1])
	})
	logic := func(name string, combine func(vals []bool) bool) {
		r.reg(name, 1, -1, func(c *xlCtx, a []xlOperand) Value {
			var vals []bool
			for _, o := range a {
				if o.isRef {
					for _, v := range c.cells(o) {
						switch v.Kind {
						case KindError:
							return v
						case KindBool:
							vals = append(vals, v.Bool)
						case KindNumber:
							vals = append(vals, v.Num != 0)
						}
					}
					continue
				}
				b, e := xlBoolOf(o.v)
				if e != "" {
					return xlErrV(e)
				}
				vals = append(vals, b)
			}
			if len(vals) == 0 {
				return xlErrV(ErrValue)
			}
			return xlBoolV(combine(vals))
		})
	}
	logic("AND", func(v []bool) bool {
		for _, b := range v {
			if !b {
				return false
			}
		}
		return true
	})
	logic("OR", func(v []bool) bool {
		for _, b := range v {
			if b {
				return true
			}
		}
		return false
	})
	logic("XOR", func(v []bool) bool {
		n := 0
		for _, b := range v {
			if b {
				n++
			}
		}
		return n%2 == 1
	})
	r.reg("NOT", 1, 1, func(c *xlCtx, a []xlOperand) Value {
		b, e := xlBoolOf(c.scalar(a[0]))
		if e != "" {
			return xlErrV(e)
		}
		return xlBoolV(!b)
	})
	r.reg("TRUE", 0, 0, func(c *xlCtx, a []xlOperand) Value { return xlBoolV(true) })
	r.reg("FALSE", 0, 0, func(c *xlCtx, a []xlOperand) Value { return xlBoolV(false) })
	r.reg("NA", 0, 0, func(c *xlCtx, a []xlOperand) Value { return xlErrV(ErrNA) })
	is := func(name string, f func(v Value) bool) {
		r.regTol(name, 1, 1, func(c *xlCtx, a []xlOperand) Value { return xlBoolV(f(c.scalar(a[0]))) })
	}
	is("ISBLANK", func(v Value) bool { return v.Kind == KindEmpty })
	is("ISNUMBER", func(v Value) bool { return v.Kind == KindNumber })
	is("ISTEXT", func(v Value) bool { return v.Kind == KindString })
	is("ISERROR", func(v Value) bool { return v.Kind == KindError })
	is("ISNA", func(v Value) bool { return v.Kind == KindError && v.Err == ErrNA })
	is("ISLOGICAL", func(v Value) bool { return v.Kind == KindBool })

	// ---- lookup and reference ----
	r.regTol("ROW", 0, 1, func(c *xlCtx, a []xlOperand) Value {
		if len(a) == 0 {
			return xlNumV(float64(c.at.Row + 1))
		}
		if !a[0].isRef {
			return xlErrV(ErrValue)
		}
		return xlNumV(float64(a[0].rng.Min.Row + 1))
	})
	r.regTol("COLUMN", 0, 1, func(c *xlCtx, a []xlOperand) Value {
		if len(a) == 0 {
			return xlNumV(float64(c.at.Col + 1))
		}
		if !a[0].isRef {
			return xlErrV(ErrValue)
		}
		return xlNumV(float64(a[0].rng.Min.Col + 1))
	})
	r.regTol("ROWS", 1, 1, func(c *xlCtx, a []xlOperand) Value {
		if !a[0].isRef {
			return xlNumV(1)
		}
		return xlNumV(float64(a[0].rng.Rows()))
	})
	r.regTol("COLUMNS", 1, 1, func(c *xlCtx, a []xlOperand) Value {
		if !a[0].isRef {
			return xlNumV(1)
		}
		return xlNumV(float64(a[0].rng.Cols()))
	})
	r.reg("INDEX", xlArgs2, xlArgs3, func(c *xlCtx, a []xlOperand) Value {
		if !a[0].isRef {
			return xlErrV(ErrValue)
		}
		r, e := c.numAt(a, 1, 0)
		if e != "" {
			return xlErrV(e)
		}
		col, e := c.numAt(a, xlArgThird, 0)
		if e != "" {
			return xlErrV(e)
		}
		rows, cols := a[0].rng.Rows(), a[0].rng.Cols()
		ri, ci := int(r), int(col)
		if len(a) < xlArgs3 || !xlHas(a, xlArgThird) {
			if rows == 1 {
				ri, ci = 1, int(r)
			} else {
				ci = 1
			}
		}
		if ri < 1 || ci < 1 || ri > rows || ci > cols {
			return xlErrV(ErrRef)
		}
		return a[0].sh.cellVal(CellRef{a[0].rng.Min.Row + ri - 1, a[0].rng.Min.Col + ci - 1})
	})
	r.reg("MATCH", xlArgs2, xlArgs3, func(c *xlCtx, a []xlOperand) Value {
		look := c.scalar(a[0])
		if !a[1].isRef {
			return xlErrV(ErrValue)
		}
		t, _ := c.numAt(a, xlArgThird, 1)
		vec := xlVector(c.rect(a[1]))
		if i := xlMatchIndex(look, vec, int(t)); i >= 0 {
			return xlNumV(float64(i + 1))
		}
		return xlErrV(ErrNA)
	})
	lookup := func(name string, byRow bool) {
		r.reg(name, xlArgs3, xlArgs4, func(c *xlCtx, a []xlOperand) Value {
			look := c.scalar(a[0])
			if !a[1].isRef {
				return xlErrV(ErrValue)
			}
			idx, e := c.numAt(a, xlArgThird, 1)
			if e != "" {
				return xlErrV(e)
			}
			approx := true
			if xlHas(a, xlArgLookupApprox) {
				b, e := xlBoolOf(c.scalar(a[3]))
				if e != "" {
					return xlErrV(e)
				}
				approx = b
			}
			t := c.rect(a[1])
			var keys []Value
			size := 0
			if byRow {
				keys, size = t[0], len(t)
			} else {
				for _, r := range t {
					keys = append(keys, r[0])
				}
				size = len(t[0])
			}
			if int(idx) < 1 || int(idx) > size {
				return xlErrV(ErrRef)
			}
			mt := 0
			if approx {
				mt = 1
			}
			i := xlMatchIndex(look, keys, mt)
			if i < 0 {
				return xlErrV(ErrNA)
			}
			if byRow {
				return t[int(idx)-1][i]
			}
			return t[i][int(idx)-1]
		})
	}
	lookup("VLOOKUP", false)
	lookup("HLOOKUP", true)
	r.regTol("XLOOKUP", xlArgs3, xlArgs5, func(c *xlCtx, a []xlOperand) Value {
		look := c.scalar(a[0])
		if look.isErr() {
			return look
		}
		if !a[1].isRef || !a[2].isRef {
			return xlErrV(ErrValue)
		}
		lr, rr := c.rect(a[1]), c.rect(a[2])
		keys := xlVector(lr)
		column := len(lr[0]) == 1
		if (column && len(rr) != len(keys)) || (!column && len(rr[0]) != len(keys)) {
			return xlErrV(ErrValue)
		}
		mode, _ := c.numAt(a, xlArgMatchMode, 0)
		i := xlExactIndex(look, keys)
		if i < 0 && mode != 0 {
			i = xlNearestIndex(look, keys, int(mode))
		}
		if i < 0 {
			if xlHas(a, xlArgLookupApprox) {
				return c.scalar(a[3])
			}
			return xlErrV(ErrNA)
		}
		if column {
			return rr[i][0]
		}
		return rr[0][i]
	})

	// ---- text ----
	r.reg("CONCAT", 1, -1, concatFn)
	r.reg("CONCATENATE", 1, -1, concatFn)
	r.reg("TEXTJOIN", xlArgs3, -1, func(c *xlCtx, a []xlOperand) Value {
		delim := c.strAt(a, 0)
		skip, e := xlBoolOf(c.scalar(a[1]))
		if e != "" {
			return xlErrV(e)
		}
		var parts []string
		for _, o := range a[2:] {
			for _, v := range c.cells(o) {
				if v.isErr() {
					return v
				}
				if skip && (v.Kind == KindEmpty || (v.Kind == KindString && v.Str == "")) {
					continue
				}
				parts = append(parts, v.String())
			}
		}
		return xlStrV(strings.Join(parts, delim))
	})
	r.reg("LEN", 1, 1, func(c *xlCtx, a []xlOperand) Value {
		return xlNumV(float64(utf8.RuneCountInString(c.strAt(a, 0))))
	})
	sub := func(name string, f func(rs []rune, n int) string) {
		r.reg(name, 1, xlArgs2, func(c *xlCtx, a []xlOperand) Value {
			n, e := c.numAt(a, 1, 1)
			if e != "" {
				return xlErrV(e)
			}
			if n < 0 {
				return xlErrV(ErrValue)
			}
			return xlStrV(f([]rune(c.strAt(a, 0)), int(n)))
		})
	}
	sub("LEFT", func(rs []rune, n int) string { return xlSub(rs, 0, n) })
	sub("RIGHT", func(rs []rune, n int) string { return xlSub(rs, len(rs)-n, n) })
	r.reg("MID", xlArgs3, xlArgs3, func(c *xlCtx, a []xlOperand) Value {
		st, e := c.numAt(a, 1, 1)
		if e != "" {
			return xlErrV(e)
		}
		n, e := c.numAt(a, xlArgThird, 0)
		if e != "" {
			return xlErrV(e)
		}
		if st < 1 || n < 0 {
			return xlErrV(ErrValue)
		}
		return xlStrV(xlSub([]rune(c.strAt(a, 0)), int(st)-1, int(n)))
	})
	textFn := func(name string, f func(string) string) {
		r.reg(name, 1, 1, func(c *xlCtx, a []xlOperand) Value { return xlStrV(f(c.strAt(a, 0))) })
	}
	textFn("UPPER", strings.ToUpper)
	textFn("LOWER", strings.ToLower)
	textFn("TRIM", func(s string) string { return strings.Join(strings.Fields(s), " ") })
	textFn("PROPER", func(s string) string {
		prev := false
		return strings.Map(func(r rune) rune {
			isL := unicode.IsLetter(r)
			defer func() { prev = isL }()
			if isL && !prev {
				return unicode.ToUpper(r)
			}
			if isL {
				return unicode.ToLower(r)
			}
			return r
		}, s)
	})
	r.reg("SUBSTITUTE", xlArgs3, xlArgs4, func(c *xlCtx, a []xlOperand) Value {
		s, old, nw := c.strAt(a, 0), c.strAt(a, 1), c.strAt(a, xlArgThird)
		if old == "" {
			return xlStrV(s)
		}
		if !xlHas(a, xlArgLookupApprox) {
			return xlStrV(strings.ReplaceAll(s, old, nw))
		}
		inst, e := c.numAt(a, xlArgFourth, 1)
		if e != "" || inst < 1 {
			return xlErrV(ErrValue)
		}
		idx, from := -1, 0
		for k := 0; k < int(inst); k++ {
			j := strings.Index(s[from:], old)
			if j < 0 {
				return xlStrV(s)
			}
			idx = from + j
			from = idx + len(old)
		}
		return xlStrV(s[:idx] + nw + s[idx+len(old):])
	})
	find := func(name string, fold bool) {
		r.reg(name, xlArgs2, xlArgs3, func(c *xlCtx, a []xlOperand) Value {
			needle, hay := c.strAt(a, 0), c.strAt(a, 1)
			if fold {
				needle, hay = strings.ToLower(needle), strings.ToLower(hay)
			}
			st, e := c.numAt(a, xlArgThird, 1)
			if e != "" || st < 1 {
				return xlErrV(ErrValue)
			}
			rs := []rune(hay)
			if int(st)-1 > len(rs) {
				return xlErrV(ErrValue)
			}
			off := len(string(rs[:int(st)-1]))
			j := strings.Index(hay[off:], needle)
			if j < 0 {
				return xlErrV(ErrValue)
			}
			return xlNumV(float64(utf8.RuneCountInString(hay[:off+j]) + 1))
		})
	}
	find("FIND", false)
	find("SEARCH", true)
	r.reg("REPT", xlArgs2, xlArgs2, func(c *xlCtx, a []xlOperand) Value {
		n, e := c.numAt(a, 1, 0)
		if e != "" || n < 0 || n > 10000 {
			return xlErrV(ErrValue)
		}
		return xlStrV(strings.Repeat(c.strAt(a, 0), int(n)))
	})
	r.reg("EXACT", xlArgs2, xlArgs2, func(c *xlCtx, a []xlOperand) Value { return xlBoolV(c.strAt(a, 0) == c.strAt(a, 1)) })
	r.reg("VALUE", 1, 1, func(c *xlCtx, a []xlOperand) Value {
		v := c.scalar(a[0])
		if v.Kind == KindNumber {
			return v
		}
		f, e := xlNumOf(v)
		if e != "" {
			return xlErrV(e)
		}
		return xlNumV(f)
	})
	r.reg("TEXT", xlArgs2, xlArgs2, func(c *xlCtx, a []xlOperand) Value {
		v := c.scalar(a[0])
		code := c.strAt(a, 1)
		if v.Kind == KindNumber {
			if strings.EqualFold(code, "general") {
				return xlStrV(xlGeneral(v.Num))
			}
			return xlStrV(xlFormatCode(v.Num, code))
		}
		return xlStrV(v.String())
	})

	// ---- date and time ----
	r.reg("TODAY", 0, 0, func(c *xlCtx, a []xlOperand) Value {
		return xlNumV(math.Floor(xlTimeToSerial(c.wb.now())))
	})
	r.reg("NOW", 0, 0, func(c *xlCtx, a []xlOperand) Value { return xlNumV(xlTimeToSerial(c.wb.now())) })
	r.reg("DATE", xlArgs3, xlArgs3, func(c *xlCtx, a []xlOperand) Value {
		var p [3]float64
		for i := range p {
			f, e := c.numAt(a, i, 0)
			if e != "" {
				return xlErrV(e)
			}
			p[i] = f
		}
		y := int(p[0])
		if y >= 0 && y < 100 {
			y += 1900
		}
		return xlNumV(xlDateSerial(y, int(p[1]), int(p[2])))
	})
	datePart := func(name string, f func(t time.Time) int) {
		r.reg(name, 1, 1, func(c *xlCtx, a []xlOperand) Value {
			s, e := c.numAt(a, 0, 0)
			if e != "" {
				return xlErrV(e)
			}
			if s < 0 {
				return xlErrV(ErrNum)
			}
			return xlNumV(float64(f(xlSerialToTime(s))))
		})
	}
	datePart("YEAR", func(t time.Time) int { return t.Year() })
	datePart("MONTH", func(t time.Time) int { return int(t.Month()) })
	datePart("DAY", func(t time.Time) int { return t.Day() })
	r.reg("WEEKDAY", 1, xlArgs2, func(c *xlCtx, a []xlOperand) Value {
		s, e := c.numAt(a, 0, 0)
		if e != "" {
			return xlErrV(e)
		}
		typ, _ := c.numAt(a, 1, 1)
		wd := int(xlSerialToTime(s).Weekday()) // Sunday=0
		if typ == xlWeekdayMondayIs1 {
			return xlNumV(float64((wd+xlDaysPerWeek-1)%xlDaysPerWeek + 1))
		}
		return xlNumV(float64(wd + 1))
	})
	r.reg("DATEDIF", xlArgs3, xlArgs3, func(c *xlCtx, a []xlOperand) Value {
		s, e := c.numAt(a, 0, 0)
		if e != "" {
			return xlErrV(e)
		}
		en, e := c.numAt(a, 1, 0)
		if e != "" {
			return xlErrV(e)
		}
		if s > en {
			return xlErrV(ErrNum)
		}
		t1, t2 := xlSerialToTime(math.Floor(s)), xlSerialToTime(math.Floor(en))
		months := (t2.Year()-t1.Year())*xlMonthsPerYear + int(t2.Month()) - int(t1.Month())
		if t2.Day() < t1.Day() {
			months--
		}
		switch strings.ToUpper(c.strAt(a, xlArgThird)) {
		case "Y":
			return xlNumV(float64(months / xlMonthsPerYear))
		case "M":
			return xlNumV(float64(months))
		case "D":
			return xlNumV(math.Floor(en) - math.Floor(s))
		case "YM":
			return xlNumV(float64(months % xlMonthsPerYear))
		}
		return xlErrV(ErrNum)
	})
	return r
}

func concatFn(c *xlCtx, a []xlOperand) Value {
	var sb strings.Builder
	for _, o := range a {
		for _, v := range c.cells(o) {
			if v.isErr() {
				return v
			}
			sb.WriteString(v.String())
		}
	}
	return xlStrV(sb.String())
}

// valOrZero evaluates a branch; an empty result reads as 0.
func (c *xlCtx) valOrZero(n xlNode) Value {
	if _, ok := n.(nEmpty); ok {
		return xlNumV(0)
	}
	v := c.val(n)
	if v.Kind == KindEmpty {
		return xlNumV(0)
	}
	return v
}

// xlVector flattens a one-dimensional rectangle.
func xlVector(r [][]Value) []Value {
	if len(r) == 1 {
		return r[0]
	}
	out := make([]Value, 0, len(r))
	for _, row := range r {
		out = append(out, row[0])
	}
	return out
}

func xlExactIndex(look Value, vec []Value) int {
	for i, v := range vec {
		if look.Kind == KindString && (strings.ContainsAny(look.Str, "*?")) {
			if v.Kind == KindString && xlWildMatch(look.Str, v.Str) {
				return i
			}
			continue
		}
		if v.Kind != KindEmpty && xlCmpRank(v) == xlCmpRank(look) && xlCompare(v, look) == 0 {
			return i
		}
	}
	return -1
}

// xlNearestIndex finds the closest match: mode 1 = smallest >= look, -1 = largest <= look.
func xlNearestIndex(look Value, vec []Value, mode int) int {
	best := -1
	for i, v := range vec {
		if v.Kind == KindEmpty || xlCmpRank(v) != xlCmpRank(look) {
			continue
		}
		c := xlCompare(v, look)
		switch {
		case mode > 0 && c >= 0 && (best < 0 || xlCompare(v, vec[best]) < 0):
			best = i
		case mode < 0 && c <= 0 && (best < 0 || xlCompare(v, vec[best]) > 0):
			best = i
		}
	}
	return best
}

// xlMatchIndex implements MATCH: 0 exact, 1 largest <= (ascending), -1 smallest >= (descending).
func xlMatchIndex(look Value, vec []Value, typ int) int {
	switch {
	case typ == 0:
		return xlExactIndex(look, vec)
	case typ > 0:
		best := -1
		for i, v := range vec {
			if v.Kind == KindEmpty || xlCmpRank(v) != xlCmpRank(look) {
				continue
			}
			if xlCompare(v, look) <= 0 {
				best = i
			} else {
				break
			}
		}
		return best
	}
	best := -1
	for i, v := range vec {
		if v.Kind == KindEmpty || xlCmpRank(v) != xlCmpRank(look) {
			continue
		}
		if xlCompare(v, look) >= 0 {
			best = i
		} else {
			break
		}
	}
	return best
}
