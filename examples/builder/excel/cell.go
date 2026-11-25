package main

import (
	"math"
	"strconv"
	"strings"
	"time"
)

// Grid limits shared by the engine and the UI.
const (
	MaxRows = 1000
	MaxCols = 52
)

// CellRef addresses one cell with 0-based row and column.
type CellRef struct{ Row, Col int }

// ColName converts a 0-based column to letters: 0 -> "A", 26 -> "AA".
func ColName(col int) string {
	if col < 0 {
		return ""
	}
	s := ""
	for col >= 0 {
		s = string(rune('A'+col%26)) + s
		col = col/xlAlphabetLen - 1
	}
	return s
}

// A1 renders the reference in A1 notation.
func (r CellRef) A1() string { return ColName(r.Col) + strconv.Itoa(r.Row+1) }

// xlRefPart is one parsed A1 reference with its absolute markers.
type xlRefPart struct {
	Col, Row     int
	AbsCol, AbsR bool
}

func (p xlRefPart) text() string {
	s := ""
	if p.AbsCol {
		s += "$"
	}
	s += ColName(p.Col)
	if p.AbsR {
		s += "$"
	}
	return s + strconv.Itoa(p.Row+1)
}

// xlParseRefPart parses "$B$3" without bounds checks (1-3 letters, digits).
func xlParseRefPart(s string) (xlRefPart, bool) {
	var p xlRefPart
	i := 0
	if i < len(s) && s[i] == '$' {
		p.AbsCol = true
		i++
	}
	start := i
	col := 0
	for i < len(s) && (s[i]|xlASCIICaseBit) >= 'a' && (s[i]|xlASCIICaseBit) <= 'z' {
		col = col*26 + int(s[i]|0x20-'a') + 1
		i++
	}
	if i == start || i-start > 3 {
		return p, false
	}
	p.Col = col - 1
	if i < len(s) && s[i] == '$' {
		p.AbsR = true
		i++
	}
	ds := i
	row := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		row = row*xlDecimalBase + int(s[i]-'0')
		if row > 1<<20 {
			return p, false
		}
		i++
	}
	if i == ds || i != len(s) || row < 1 {
		return p, false
	}
	p.Row = row - 1
	return p, true
}

// ParseA1 parses "B3" or "$B$3" within the grid bounds.
func ParseA1(s string) (CellRef, bool) {
	p, ok := xlParseRefPart(strings.TrimSpace(s))
	if !ok || p.Row >= MaxRows || p.Col >= MaxCols {
		return CellRef{}, false
	}
	return CellRef{p.Row, p.Col}, true
}

// Range is a normalized rectangle (Min <= Max).
type Range struct{ Min, Max CellRef }

// NormRange builds a normalized range from two corners.
func NormRange(a, b CellRef) Range {
	return Range{
		Min: CellRef{min(a.Row, b.Row), min(a.Col, b.Col)},
		Max: CellRef{max(a.Row, b.Row), max(a.Col, b.Col)},
	}
}

// ParseRange parses "A1:B3" or a single cell.
func ParseRange(s string) (Range, bool) {
	a, b, found := strings.Cut(strings.TrimSpace(s), ":")
	ca, ok := ParseA1(a)
	if !ok {
		return Range{}, false
	}
	if !found {
		return Range{ca, ca}, true
	}
	cb, ok := ParseA1(b)
	if !ok {
		return Range{}, false
	}
	return NormRange(ca, cb), true
}

func (r Range) String() string {
	if r.Min == r.Max {
		return r.Min.A1()
	}
	return r.Min.A1() + ":" + r.Max.A1()
}

// Contains reports whether c lies inside the range.
func (r Range) Contains(c CellRef) bool {
	return c.Row >= r.Min.Row && c.Row <= r.Max.Row && c.Col >= r.Min.Col && c.Col <= r.Max.Col
}

// Rows and Cols return the range extent.
func (r Range) Rows() int { return r.Max.Row - r.Min.Row + 1 }
func (r Range) Cols() int { return r.Max.Col - r.Min.Col + 1 }

// Intersects reports whether two ranges overlap.
func (r Range) Intersects(o Range) bool {
	return r.Min.Row <= o.Max.Row && o.Min.Row <= r.Max.Row && r.Min.Col <= o.Max.Col && o.Min.Col <= r.Max.Col
}

// clip limits the range to the grid.
func (r Range) clip() Range {
	r.Min.Row, r.Min.Col = max(r.Min.Row, 0), max(r.Min.Col, 0)
	r.Max.Row, r.Max.Col = min(r.Max.Row, MaxRows-1), min(r.Max.Col, MaxCols-1)
	return r
}

// ValueKind tags a Value.
type ValueKind int

const (
	KindEmpty ValueKind = iota
	KindNumber
	KindString
	KindBool
	KindError
)

// Value is a computed cell value.
type Value struct {
	Kind ValueKind
	Num  float64
	Str  string
	Bool bool
	Err  string
}

// Formula error codes.
const (
	ErrDiv0  = "#DIV/0!"
	ErrName  = "#NAME?"
	ErrValue = "#VALUE!"
	ErrRef   = "#REF!"
	ErrNA    = "#N/A"
	ErrNum   = "#NUM!"
	ErrCirc  = "#CIRC!"
	ErrSyn   = "#ERROR!"
)

func xlNumV(f float64) Value { return Value{Kind: KindNumber, Num: f} }
func xlStrV(s string) Value  { return Value{Kind: KindString, Str: s} }
func xlBoolV(b bool) Value   { return Value{Kind: KindBool, Bool: b} }
func xlErrV(e string) Value  { return Value{Kind: KindError, Err: e} }
func (v Value) isErr() bool  { return v.Kind == KindError }

// Style is the base formatting of one cell.
type Style struct {
	Bold, Italic, Underline, Strike, Wrap bool
	Align                                 string // "", "left", "center", "right"
	FG, BG                                string // css colors, "" = default
	// NumFmt is "General","Number","Currency","Accounting","Percent","Date","Scientific","Text",
	// or a custom format code such as "0.00" or "yyyy-mm-dd".
	NumFmt string
	// Decimals: 0 = the format's default (Number/Currency/Accounting/Scientific 2, Percent 0,
	// General none), n>0 = n decimals, negative = zero decimals.
	Decimals int
	FontSize int // 0 = default
	// FontFamily is "", "sans", "mono" or "serif" ("" = the sheet default).
	FontFamily string
	// VAlign is "", "top", "middle" or "bottom" ("" = bottom, as in Excel).
	VAlign string
	// Indent is the left indent in steps of one level (0..15).
	Indent int
	// Edges holds the per-edge borders. Border is the pre-edges enum
	// ("all", "outline", "bottom", "top"); it is still read (see BorderEdges)
	// and migrated into Edges when a file is loaded.
	Edges  Edges
	Border string
}

// Edge line styles.
const (
	EdgeNone   = ""
	EdgeThin   = "thin"
	EdgeThick  = "thick"
	EdgeDouble = "double"
)

// Edges are the four cell borders, each "", "thin", "thick" or "double".
type Edges struct{ Left, Right, Top, Bottom string }

// Any reports whether at least one edge is drawn.
func (e Edges) Any() bool { return e.Left != "" || e.Right != "" || e.Top != "" || e.Bottom != "" }

// BorderEdges resolves the legacy Border enum and the explicit edges into the
// borders to paint. Explicit edges win over the legacy value.
func (st Style) BorderEdges() Edges {
	e := st.Edges
	set := func(dst *string, v string) {
		if *dst == "" {
			*dst = v
		}
	}
	switch st.Border {
	case "all", "outline":
		set(&e.Left, EdgeThin)
		set(&e.Right, EdgeThin)
		set(&e.Top, EdgeThin)
		set(&e.Bottom, EdgeThin)
	case "bottom":
		set(&e.Bottom, EdgeThin)
	case "top":
		set(&e.Top, EdgeThin)
	}
	return e
}

// HasBorder reports whether any border is drawn for the style.
func (st Style) HasBorder() bool { return st.BorderEdges().Any() }

// migrated folds the legacy Border enum into Edges.
func (st Style) migrated() Style {
	if st.Border != "" {
		st.Edges = st.BorderEdges()
		st.Border = ""
	}
	return st
}

// xlLiteral is the result of interpreting typed text.
type xlLiteral struct {
	val    Value
	numFmt string
	dec    int
}

// xlEpoch returns the spreadsheet serial-date origin (serial 0).
func xlEpoch() time.Time { return time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC) }

const (
	secondsPerDay    = 86400
	xlAlphabetLen    = 26   // letters in a column name digit
	xlASCIICaseBit   = 0x20 // OR-ing this folds ASCII upper case to lower case
	xlThousandsGroup = 3    // digits between thousands separators
)

func xlTimeToSerial(t time.Time) float64 {
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
	return t.Sub(xlEpoch()).Seconds() / secondsPerDay
}

func xlSerialToTime(f float64) time.Time {
	return xlEpoch().Add(time.Duration(math.Round(f*secondsPerDay)) * time.Second)
}

// xlIsNumText reports whether s is a plain decimal or scientific number.
func xlIsNumText(s string) bool {
	i, digits := 0, 0
	if i < len(s) && (s[i] == '-' || s[i] == '+') {
		i++
	}
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
		digits++
	}
	if i < len(s) && s[i] == '.' {
		i++
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
			digits++
		}
	}
	if digits == 0 {
		return false
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i < len(s) && (s[i] == '-' || s[i] == '+') {
			i++
		}
		e := 0
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
			e++
		}
		if e == 0 {
			return false
		}
	}
	return i == len(s)
}

// xlStripThousands removes valid thousands separators ("1,250.5" -> "1250.5").
func xlStripThousands(s string) (string, bool) {
	if !strings.Contains(s, ",") {
		return s, true
	}
	body := strings.TrimLeft(s, "+-")
	intPart, frac, _ := strings.Cut(body, ".")
	groups := strings.Split(intPart, ",")
	if len(groups[0]) == 0 || len(groups[0]) > xlThousandsGroup {
		return s, false
	}
	for _, g := range groups[1:] {
		if len(g) != xlThousandsGroup {
			return s, false
		}
	}
	out := strings.Join(groups, "")
	if strings.Contains(body, ".") {
		out += "." + frac
	}
	return s[:len(s)-len(body)] + out, true
}

func xlDecimalsOf(s string) int {
	_, frac, ok := strings.Cut(s, ".")
	if !ok {
		return 0
	}
	n := 0
	for n < len(frac) && frac[n] >= '0' && frac[n] <= '9' {
		n++
	}
	return n
}

// xlParseNumber parses plain numbers, thousands groups, percents and currency.
func xlParseNumber(s string) (float64, bool) {
	l := xlInfer(s)
	if l.val.Kind == KindNumber && l.numFmt != "Date" {
		return l.val.Num, true
	}
	return 0, false
}

// xlInfer interprets typed text as number, percent, currency, bool, date or string.
func xlInfer(raw string) xlLiteral {
	if raw == "" {
		return xlLiteral{}
	}
	if raw[0] == '\'' {
		return xlLiteral{val: xlStrV(raw[1:])}
	}
	t := strings.TrimSpace(raw)
	switch strings.ToUpper(t) {
	case "TRUE":
		return xlLiteral{val: xlBoolV(true)}
	case "FALSE":
		return xlLiteral{val: xlBoolV(false)}
	}
	if t == "" {
		return xlLiteral{val: xlStrV(raw)}
	}
	if c := t[0]; c == '-' || c == '+' || c == '.' || c == '$' || (c >= '0' && c <= '9') {
		if n, ok := xlStripThousands(t); ok && xlIsNumText(n) {
			f, err := strconv.ParseFloat(n, 64)
			if err == nil {
				return xlLiteral{val: xlNumV(f)}
			}
		}
		if body, ok := strings.CutSuffix(t, "%"); ok {
			if n, ok := xlStripThousands(body); ok && xlIsNumText(n) {
				if f, err := strconv.ParseFloat(n, 64); err == nil {
					return xlLiteral{val: xlNumV(f / xlPercentScale), numFmt: "Percent", dec: xlDecOrNeg(xlDecimalsOf(n))}
				}
			}
		}
		neg := false
		cur := t
		if strings.HasPrefix(cur, "-$") {
			neg, cur = true, cur[1:]
		}
		if body, ok := strings.CutPrefix(cur, "$"); ok {
			if n, ok := xlStripThousands(body); ok && xlIsNumText(n) {
				if f, err := strconv.ParseFloat(n, 64); err == nil {
					if neg {
						f = -f
					}
					return xlLiteral{val: xlNumV(f), numFmt: "Currency", dec: xlCurrencyDec(xlDecimalsOf(n))}
				}
			}
		}
		for _, layout := range []string{"2006-01-02", "1/2/2006", "2006/1/2", "2006-1-2"} {
			if d, err := time.Parse(layout, t); err == nil {
				return xlLiteral{val: xlNumV(math.Floor(xlTimeToSerial(d))), numFmt: "Date"}
			}
		}
	}
	return xlLiteral{val: xlStrV(raw)}
}

func xlDecOrNeg(n int) int {
	if n == 0 {
		return -1
	}
	return n
}

// xlCurrencyDec keeps "$5" at zero decimals and pads partial decimals to cents.
func xlCurrencyDec(n int) int {
	if n == 0 {
		return -1
	}
	return max(n, xlDefaultDecimals)
}
