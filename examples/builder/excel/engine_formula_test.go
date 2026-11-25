package main

import (
	"math"
	"testing"
	"time"
)

// xlEvalIn evaluates a formula in a scratch cell of sheet s.
func xlEvalIn(s *Sheet, formula string) Value {
	s.SetRaw(MaxRows-1, MaxCols-1, formula)
	return s.Value(MaxRows-1, MaxCols-1)
}

func xlSeeded() (*Workbook, *Sheet) {
	w := NewWorkbook()
	s := w.Sheets[1]
	data := [][]string{
		{"North", "10", "a"},
		{"South", "20", "b"},
		{"North", "30", "c"},
		{"East", "40", ""},
	}
	for r, row := range data {
		for c, v := range row {
			s.SetRaw(r, c, v)
		}
	}
	return w, s
}

func TestEngineLexer(t *testing.T) {
	toks, err := xlLex(`SUM(A1:B2, 1.5e-3) & "x""y" <> 'My Sheet'!$C$4`)
	if err != nil {
		t.Fatal(err)
	}
	var kinds []xlTokKind
	for _, k := range toks {
		kinds = append(kinds, k.kind)
	}
	want := []xlTokKind{tkFunc, tkLParen, tkRef, tkColon, tkRef, tkComma, tkNum, tkRParen, tkOp, tkStr, tkOp, tkSheet, tkRef, tkEOF}
	if len(kinds) != len(want) {
		t.Fatalf("kinds %v want %v", kinds, want)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("tok %d kind %v want %v", i, kinds[i], want[i])
		}
	}
	if toks[6].num != 1.5e-3 || toks[9].text != `x"y` || toks[11].text != "My Sheet" {
		t.Fatalf("bad token payloads %+v", toks)
	}
	for _, bad := range []string{`"open`, `'Sheet A1`, `1 @ 2`} {
		if _, err := xlLex(bad); err == nil {
			t.Errorf("expected lex error for %q", bad)
		}
	}
}

func TestEngineParserPrecedence(t *testing.T) {
	w := NewWorkbook()
	s := w.Sheet()
	cases := map[string]float64{
		"=-2^2":            4, // unary minus binds tighter than ^
		"=1+2*3":           7,
		"=(1+2)*3":         9,
		"=2^3^2":           64, // left associative
		"=50%":             0.5,
		"=10-2-3":          5,
		"=2*-3":            -6,
		"=1+2&\"3\"":       0, // checked separately below
		"=SUM(1,2)*2":      6,
		"=2^-1":            0.5,
		"=100/10/5":        2,
		"=SUM(1,SUM(2,3))": 6,
	}
	for f, want := range cases {
		if f == "=1+2&\"3\"" {
			continue
		}
		v := xlEvalIn(s, f)
		if v.Kind != KindNumber || math.Abs(v.Num-want) > 1e-12 {
			t.Errorf("%s = %+v want %v", f, v, want)
		}
	}
	if v := xlEvalIn(s, `="a"&1+2`); v.Kind != KindString || v.Str != "a3" {
		t.Errorf("concat precedence: %+v", v)
	}
	if v := xlEvalIn(s, `=1+2=3`); v.Kind != KindBool || !v.Bool {
		t.Errorf("comparison lowest: %+v", v)
	}
	for _, bad := range []string{"=1+", "=(1", "=SUM(1,", "=1 2", "=SUM(A1:)"} {
		if v := xlEvalIn(s, bad); v.Kind != KindError {
			t.Errorf("%s should error, got %+v", bad, v)
		}
	}
}

func TestEngineOperatorsAndCoercion(t *testing.T) {
	w, s := xlSeeded()
	_ = w
	num := func(f string, want float64) {
		t.Helper()
		v := xlEvalIn(s, f)
		if v.Kind != KindNumber || math.Abs(v.Num-want) > 1e-9 {
			t.Errorf("%s = %+v want %v", f, v, want)
		}
	}
	str := func(f, want string) {
		t.Helper()
		v := xlEvalIn(s, f)
		if v.Kind != KindString || v.Str != want {
			t.Errorf("%s = %+v want %q", f, v, want)
		}
	}
	errIs := func(f, want string) {
		t.Helper()
		v := xlEvalIn(s, f)
		if v.Kind != KindError || v.Err != want {
			t.Errorf("%s = %+v want %s", f, v, want)
		}
	}
	boolIs := func(f string, want bool) {
		t.Helper()
		v := xlEvalIn(s, f)
		if v.Kind != KindBool || v.Bool != want {
			t.Errorf("%s = %+v want %v", f, v, want)
		}
	}
	num(`="5"+1`, 6)
	num(`=TRUE+TRUE`, 2)
	num(`=C9+1`, 1) // empty as 0
	num(`=B1*B2`, 200)
	str(`=A1&" "&B1`, "North 10")
	str(`=C9&"x"`, "x")
	errIs(`=1/0`, ErrDiv0)
	errIs(`="abc"+1`, ErrValue)
	errIs(`=NOPE(1)`, ErrName)
	errIs(`=Sheet9!A1`, ErrRef)
	errIs(`=#REF!+1`, ErrRef)
	errIs(`=1+1/0`, ErrDiv0)
	errIs(`=SUM(1/0,2)`, ErrDiv0)
	errIs(`=-("x")`, ErrValue)
	errIs(`=2^1000000`, ErrNum)
	boolIs(`=1<2`, true)
	boolIs(`="a"="A"`, true)
	boolIs(`=1<>1`, false)
	boolIs(`="a">1`, true) // text sorts after numbers
	boolIs(`=A1="north"`, true)
	boolIs(`=C9=0`, true)
	boolIs(`=B1>=10`, true)
}

func TestEngineMathAggregates(t *testing.T) {
	_, s := xlSeeded()
	num := func(f string, want float64) {
		t.Helper()
		v := xlEvalIn(s, f)
		if v.Kind != KindNumber || math.Abs(v.Num-want) > 1e-9 {
			t.Errorf("%s = %+v want %v", f, v, want)
		}
	}
	num(`=SUM(B1:B4)`, 100)
	num(`=SUM(A1:C4)`, 100) // text ignored in ranges
	num(`=AVERAGE(B1:B4)`, 25)
	num(`=MIN(B1:B4)`, 10)
	num(`=MAX(B1:B4)`, 40)
	num(`=COUNT(A1:B4)`, 4)
	num(`=COUNTA(A1:C4)`, 11)
	num(`=COUNTBLANK(A1:C4)`, 1)
	num(`=MEDIAN(B1:B4)`, 25)
	num(`=PRODUCT(1,2,3)`, 6)
	num(`=ROUND(2.5,0)`, 3)
	num(`=ROUND(-2.5,0)`, -3)
	num(`=ROUND(1.005,2)`, 1.01)
	num(`=ROUND(1234.567,-2)`, 1200)
	num(`=ROUNDUP(2.1,0)`, 3)
	num(`=ROUNDDOWN(2.9,0)`, 2)
	num(`=ROUNDUP(-2.1,0)`, -3)
	num(`=ABS(-3)`, 3)
	num(`=MOD(-3,2)`, 1)
	num(`=POWER(2,10)`, 1024)
	num(`=SQRT(16)`, 4)
	num(`=INT(-1.5)`, -2)
	num(`=SUMIF(A1:A4,"North",B1:B4)`, 40)
	num(`=SUMIF(B1:B4,">15")`, 90)
	num(`=SUMIF(A1:A4,"*th",B1:B4)`, 60)
	num(`=COUNTIF(A1:A4,"North")`, 2)
	num(`=COUNTIF(B1:B4,"<=20")`, 2)
	num(`=COUNTIF(A1:A4,"<>North")`, 2)
	num(`=SUMIFS(B1:B4,A1:A4,"North",B1:B4,">10")`, 30)
	num(`=COUNTIFS(A1:A4,"North",C1:C4,"c")`, 1)
	num(`=AVERAGEIF(A1:A4,"North",B1:B4)`, 20)
	num(`=LARGE(B1:B4,2)`, 30)
	num(`=SMALL(B1:B4,1)`, 10)
	num(`=SUMPRODUCT(B1:B4,B1:B4)`, 3000)
	num(`=SUM(B1:B2,B4,5)`, 75)
	if v := xlEvalIn(s, `=AVERAGE(D1:D3)`); v.Err != ErrDiv0 {
		t.Errorf("AVERAGE of nothing: %+v", v)
	}
	if v := xlEvalIn(s, `=SQRT(-1)`); v.Err != ErrNum {
		t.Errorf("SQRT(-1): %+v", v)
	}
	if v := xlEvalIn(s, `=MOD(1,0)`); v.Err != ErrDiv0 {
		t.Errorf("MOD(1,0): %+v", v)
	}
}

func TestEngineLogicalAndShortCircuit(t *testing.T) {
	_, s := xlSeeded()
	str := func(f, want string) {
		t.Helper()
		if v := xlEvalIn(s, f); v.String() != want {
			t.Errorf("%s = %+v want %q", f, v, want)
		}
	}
	str(`=IF(1>0,"y","n")`, "y")
	str(`=IF(1<0,"y")`, "FALSE")
	str(`=IF(1>0,"ok",1/0)`, "ok") // untaken branch is never evaluated
	str(`=IF(0,1/0,"fine")`, "fine")
	str(`=IFERROR(1/0,"bad")`, "bad")
	str(`=IFERROR(5,1/0)`, "5")
	str(`=IFNA(NA(),"gone")`, "gone")
	str(`=IFNA(1/0,"gone")`, "#DIV/0!")
	str(`=IFS(1>2,"a",2>1,"b")`, "b")
	str(`=IFS(1>2,"a")`, "#N/A")
	str(`=AND(TRUE,1,B1>5)`, "TRUE")
	str(`=AND(TRUE,0)`, "FALSE")
	str(`=OR(FALSE,B1>50)`, "FALSE")
	str(`=XOR(TRUE,TRUE)`, "FALSE")
	str(`=NOT(FALSE)`, "TRUE")
	str(`=ISBLANK(C9)`, "TRUE")
	str(`=ISBLANK(A1)`, "FALSE")
	str(`=ISNUMBER(B1)`, "TRUE")
	str(`=ISNUMBER(A1)`, "FALSE")
	str(`=ISTEXT(A1)`, "TRUE")
	str(`=ISERROR(1/0)`, "TRUE")
	str(`=ISERROR(B1)`, "FALSE")
	str(`=IF("maybe",1,2)`, "#VALUE!")
}

func TestEngineLookup(t *testing.T) {
	w := NewWorkbook()
	s := w.Sheets[1]
	rows := [][]string{{"1", "one", "x"}, {"3", "three", "y"}, {"5", "five", "z"}}
	for r, row := range rows {
		for c, v := range row {
			s.SetRaw(r, c, v)
		}
	}
	str := func(f, want string) {
		t.Helper()
		if v := xlEvalIn(s, f); v.String() != want {
			t.Errorf("%s = %+v want %q", f, v, want)
		}
	}
	str(`=VLOOKUP(3,A1:C3,2,FALSE)`, "three")
	str(`=VLOOKUP(4,A1:C3,2)`, "three") // approximate: largest <= 4
	str(`=VLOOKUP(4,A1:C3,2,FALSE)`, "#N/A")
	str(`=VLOOKUP(0,A1:C3,2)`, "#N/A")
	str(`=VLOOKUP(3,A1:C3,9,FALSE)`, "#REF!")
	str(`=HLOOKUP(5,A1:C3,2,FALSE)`, "#N/A")
	s.SetRaw(5, 0, "a")
	s.SetRaw(5, 1, "b")
	s.SetRaw(6, 0, "10")
	s.SetRaw(6, 1, "20")
	str(`=HLOOKUP("b",A6:B7,2,FALSE)`, "20")
	str(`=INDEX(A1:C3,2,2)`, "three")
	str(`=INDEX(A1:A3,3)`, "5")
	str(`=INDEX(A1:C3,4,1)`, "#REF!")
	str(`=MATCH(3,A1:A3,0)`, "2")
	str(`=MATCH(4,A1:A3,1)`, "2")
	str(`=MATCH(4,A1:A3,0)`, "#N/A")
	str(`=MATCH("fi*",B1:B3,0)`, "3")
	str(`=XLOOKUP(5,A1:A3,B1:B3)`, "five")
	str(`=XLOOKUP(6,A1:A3,B1:B3,"none")`, "none")
	str(`=XLOOKUP(6,A1:A3,B1:B3)`, "#N/A")
	str(`=XLOOKUP(4,A1:A3,B1:B3,"n",-1)`, "three")
	str(`=ROW(B7)`, "7")
	str(`=COLUMN(C1)`, "3")
	str(`=ROWS(A1:C3)`, "3")
	str(`=COLUMNS(A1:C3)`, "3")
}

func TestEngineTextFunctions(t *testing.T) {
	_, s := xlSeeded()
	s.SetRaw(0, 5, "  Hello   World ")
	str := func(f, want string) {
		t.Helper()
		if v := xlEvalIn(s, f); v.String() != want {
			t.Errorf("%s = %+v want %q", f, v, want)
		}
	}
	str(`=LEN("héllo")`, "5")
	str(`=LEFT("abcdef",2)`, "ab")
	str(`=LEFT("abc")`, "a")
	str(`=RIGHT("abcdef",3)`, "def")
	str(`=MID("abcdef",2,3)`, "bcd")
	str(`=MID("abc",5,2)`, "")
	str(`=UPPER("abC")`, "ABC")
	str(`=LOWER("AbC")`, "abc")
	str(`=PROPER("hello wORLD")`, "Hello World")
	str(`=TRIM(F1)`, "Hello World")
	str(`=CONCAT("a",1,TRUE)`, "a1TRUE")
	str(`=CONCATENATE(A1,"-",B1)`, "North-10")
	str(`=CONCAT(A1:A2)`, "NorthSouth")
	str(`=TEXTJOIN(", ",TRUE,A1:A4)`, "North, South, North, East")
	str(`=TEXTJOIN("-",TRUE,C1:C4)`, "a-b-c")
	str(`=TEXTJOIN("-",FALSE,C1:C4)`, "a-b-c-")
	str(`=SUBSTITUTE("a-b-c","-","+")`, "a+b+c")
	str(`=SUBSTITUTE("a-b-c","-","+",2)`, "a-b+c")
	str(`=FIND("b","abcb")`, "2")
	str(`=FIND("b","abcb",3)`, "4")
	str(`=FIND("B","abc")`, "#VALUE!")
	str(`=SEARCH("B","abc")`, "2")
	str(`=REPT("ab",3)`, "ababab")
	str(`=VALUE("12.5")`, "12.5")
	str(`=TEXT(1234.5,"#,##0.00")`, "1,234.50")
	str(`=TEXT(0.256,"0.0%")`, "25.6%")
	str(`=TEXT(5,"000")`, "005")
	str(`=TEXT(DATE(2026,11,25),"yyyy-mm-dd")`, "2026-11-25")
	str(`=TEXT(DATE(2026,3,5),"d mmm yyyy")`, "5 Mar 2026")
	str(`=EXACT("a","A")`, "FALSE")
}

func TestEngineDateFunctions(t *testing.T) {
	w, s := xlSeeded()
	w.clock = func() time.Time { return time.Date(2026, 11, 25, 14, 30, 0, 0, time.UTC) }
	str := func(f, want string) {
		t.Helper()
		if v := xlEvalIn(s, f); v.String() != want {
			t.Errorf("%s = %+v want %q", f, v, want)
		}
	}
	str(`=YEAR(TODAY())`, "2026")
	str(`=MONTH(TODAY())`, "11")
	str(`=DAY(TODAY())`, "25")
	str(`=TEXT(TODAY(),"yyyy-mm-dd")`, "2026-11-25")
	str(`=TEXT(NOW(),"hh:mm")`, "14:30")
	str(`=TODAY()-DATE(2026,11,1)`, "24")
	str(`=DATE(2026,14,1)=DATE(2027,2,1)`, "TRUE")
	str(`=DATEDIF(DATE(2020,1,15),DATE(2026,11,25),"Y")`, "6")
	str(`=DATEDIF(DATE(2020,1,15),DATE(2026,11,25),"M")`, "82")
	str(`=DATEDIF(DATE(2026,1,1),DATE(2026,1,31),"D")`, "30")
	str(`=DATEDIF(DATE(2026,2,1),DATE(2026,1,1),"D")`, "#NUM!")
	str(`=WEEKDAY(DATE(2026,11,25))`, "4") // Wednesday
}

func TestEngineCrossSheetRefs(t *testing.T) {
	w := NewWorkbook()
	a, b := w.Sheets[0], w.Sheets[1]
	b.SetRaw(0, 0, "5")
	b.SetRaw(1, 0, "7")
	w.RenameSheet(1, "Sales Q1-Q4")
	a.SetRaw(20, 0, "=SUM('Sales Q1-Q4'!A1:A2)*2")
	a.SetRaw(20, 1, "=Summary!A1+1")
	if v := a.Value(20, 0); v.Num != 24 {
		t.Fatalf("quoted cross-sheet sum = %+v", v)
	}
	w.Sheets[2].SetRaw(0, 0, "9")
	if v := a.Value(20, 1); v.Num != 10 {
		t.Fatalf("unquoted cross-sheet = %+v", v)
	}
	b.SetRaw(0, 0, "100")
	if v := a.Value(20, 0); v.Num != 214 {
		t.Fatalf("recalc after other-sheet edit = %+v", v)
	}
	w.RenameSheet(1, "Data")
	if got := a.Raw(20, 0); got != "=SUM(Data!A1:A2)*2" {
		t.Fatalf("rename rewrote formula to %q", got)
	}
	if v := a.Value(20, 0); v.Num != 214 {
		t.Fatalf("value after rename = %+v", v)
	}
	w.DeleteSheet(1)
	if v := a.Value(20, 0); v.Err != ErrRef {
		t.Fatalf("deleted sheet should give #REF!, got %+v", v)
	}
}
