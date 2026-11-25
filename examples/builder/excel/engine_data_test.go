package main

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEngineA1Parsing(t *testing.T) {
	for _, c := range []struct {
		in   string
		want CellRef
		ok   bool
	}{{"A1", CellRef{0, 0}, true}, {"$B$3", CellRef{2, 1}, true}, {"AA10", CellRef{9, 26}, true},
		{"AZ1000", CellRef{999, 51}, true}, {"BA1", CellRef{}, false}, {"A1001", CellRef{}, false}, {"A0", CellRef{}, false}, {"1A", CellRef{}, false}, {"", CellRef{}, false}} {
		got, ok := ParseA1(c.in)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("ParseA1(%q) = %v,%v", c.in, got, ok)
		}
	}
	if ColName(0) != "A" || ColName(25) != "Z" || ColName(26) != "AA" || ColName(51) != "AZ" {
		t.Error("ColName")
	}
	r, ok := ParseRange("C3:A1")
	if !ok || r.String() != "A1:C3" || !r.Contains(CellRef{1, 1}) || r.Contains(CellRef{3, 0}) {
		t.Errorf("range %v", r)
	}
	if r, _ := ParseRange("B2"); r.String() != "B2" {
		t.Error("single cell range")
	}
}

func TestEngineLiteralInference(t *testing.T) {
	w := NewWorkbook()
	s := w.Sheets[1]
	type tc struct {
		raw     string
		kind    ValueKind
		num     float64
		str     string
		numFmt  string
		display string
	}
	for i, c := range []tc{
		{"42", KindNumber, 42, "", "", "42"},
		{"-3.14", KindNumber, -3.14, "", "", "-3.14"},
		{"1e3", KindNumber, 1000, "", "", "1000"},
		{"1,250", KindNumber, 1250, "", "", "1250"},
		{"12.5%", KindNumber, 0.125, "", "Percent", "12.5%"},
		{"$1,250.00", KindNumber, 1250, "", "Currency", "$1,250.00"},
		{"-$5", KindNumber, -5, "", "Currency", "-$5"},
		{"$5.5", KindNumber, 5.5, "", "Currency", "$5.50"},
		{"TRUE", KindBool, 0, "", "", "TRUE"},
		{"false", KindBool, 0, "", "", "FALSE"},
		{"2026-11-25", KindNumber, 46351, "", "Date", "2026-11-25"},
		{"11/25/2026", KindNumber, 46351, "", "Date", "2026-11-25"},
		{"'00123", KindString, 0, "00123", "", "00123"},
		{"hello", KindString, 0, "hello", "", "hello"},
		{"1,2", KindString, 0, "1,2", "", "1,2"},
		{"12abc", KindString, 0, "12abc", "", "12abc"},
		{"-", KindString, 0, "-", "", "-"},
	} {
		s.SetRaw(i, 0, c.raw)
		v := s.Value(i, 0)
		if v.Kind != c.kind || (c.kind == KindNumber && math.Abs(v.Num-c.num) > 1e-9) || (c.kind == KindString && v.Str != c.str) {
			t.Errorf("%q -> %+v", c.raw, v)
		}
		if got := s.BaseStyle(i, 0).NumFmt; got != c.numFmt {
			t.Errorf("%q numfmt %q want %q", c.raw, got, c.numFmt)
		}
		if got := s.Display(i, 0); got != c.display {
			t.Errorf("%q display %q want %q", c.raw, got, c.display)
		}
	}
	if s.Raw(0, 0) != "42" || s.Raw(12, 0) != "'00123" {
		t.Error("raw must stay as typed")
	}
}

func TestEngineNumberFormats(t *testing.T) {
	n := xlNumV
	for _, c := range []struct {
		v    Value
		st   Style
		want string
	}{
		{n(0.1 + 0.2), Style{}, "0.3"},
		{n(1234567.891), Style{}, "1234567.891"},
		{n(1e12), Style{}, "1E+12"},
		{n(0.00000001234), Style{}, "1.234E-08"},
		{n(1234.5), Style{NumFmt: "Number"}, "1234.50"},
		{n(1234.5), Style{NumFmt: "Number", Decimals: -1}, "1235"},
		{n(1234.5), Style{NumFmt: "Number", Decimals: 3}, "1234.500"},
		{n(1234.5), Style{NumFmt: "Currency"}, "$1,234.50"},
		{n(-1234.5), Style{NumFmt: "Currency"}, "-$1,234.50"},
		{n(1234.5), Style{NumFmt: "Currency", Decimals: -1}, "$1,235"},
		{n(1234.5), Style{NumFmt: "Accounting"}, "$ 1,234.50"},
		{n(-1234.5), Style{NumFmt: "Accounting"}, "($ 1,234.50)"},
		{n(0), Style{NumFmt: "Accounting"}, "$ -"},
		{n(0.256), Style{NumFmt: "Percent"}, "26%"},
		{n(0.256), Style{NumFmt: "Percent", Decimals: 1}, "25.6%"},
		{n(12345.678), Style{NumFmt: "Scientific"}, "1.23E+04"},
		{n(46351), Style{NumFmt: "Date"}, "2026-11-25"},
		{n(46351), Style{NumFmt: "LongDate"}, "Wed, Nov 25, 2026"},
		{n(46351.75), Style{NumFmt: "Time"}, "18:00:00"},
		{n(46351.5), Style{NumFmt: "yyyy-mm-dd hh:mm"}, "2026-11-25 12:00"},
		{n(1234.5), Style{NumFmt: "#,##0.00"}, "1,234.50"},
		{n(0.5), Style{NumFmt: "0.0%"}, "50.0%"},
		{n(7), Style{NumFmt: "000"}, "007"},
		{n(-0.0001), Style{NumFmt: "Number"}, "0.00"},
		{xlStrV("text"), Style{NumFmt: "Currency"}, "text"},
		{xlBoolV(true), Style{}, "TRUE"},
		{xlErrV(ErrDiv0), Style{NumFmt: "Percent"}, "#DIV/0!"},
		{Value{}, Style{NumFmt: "Percent"}, ""},
	} {
		if got := FormatValue(c.v, c.st); got != c.want {
			t.Errorf("FormatValue(%+v, %+v) = %q want %q", c.v, c.st, got, c.want)
		}
	}
}

func TestEngineStarterWorkbook(t *testing.T) {
	w := NewWorkbook()
	if len(w.Sheets) != 3 || w.Sheets[0].Name != "Sheet1" || w.Sheet() != w.Sheets[0] {
		t.Fatal("starter sheets")
	}
	s := w.Sheet()
	u := s.UsedRange()
	for r := u.Min.Row; r <= u.Max.Row; r++ {
		for c := u.Min.Col; c <= u.Max.Col; c++ {
			if v := s.Value(r, c); v.Kind == KindError {
				t.Errorf("%s has error %s (%s)", CellRef{r, c}.A1(), v.Err, s.Raw(r, c))
			}
		}
	}
	if got := s.Display(1, 6); got != "$55,600" {
		t.Errorf("G2 display %q", got)
	}
	if v := s.Value(9, 6); v.Kind != KindNumber || v.Num <= 0 {
		t.Errorf("grand total %+v", v)
	}
	if s.FreezeRows() != 1 {
		t.Error("header freeze")
	}
}

func TestEngineSortAndFilter(t *testing.T) {
	w := NewWorkbook()
	s := w.Sheets[1]
	rows := [][]string{
		{"Region", "Revenue", "Share"},
		{"South", "30", "=B2/100"},
		{"North", "10", "=B3/100"},
		{"North", "50", "=B4/100"},
		{"East", "20", "=B5/100"},
		{"", "99", ""},
	}
	for r, row := range rows {
		for c, v := range row {
			s.SetRaw(r, c, v)
		}
	}
	s.SetStyle(Range{CellRef{1, 0}, CellRef{1, 0}}, func(st *Style) { st.Bold = true }) // South is bold
	rng := Range{CellRef{0, 0}, CellRef{4, 2}}
	w.SortRange(s, rng, 0, true, true)
	want := []string{"East", "North", "North", "South"}
	for i, name := range want {
		if got := s.Raw(i+1, 0); got != name {
			t.Fatalf("row %d region %q want %q", i+1, got, name)
		}
	}
	// stable: the two North rows keep input order (10 then 50), formulas follow their rows
	if s.Raw(2, 1) != "10" || s.Raw(3, 1) != "50" {
		t.Errorf("stable order %q %q", s.Raw(2, 1), s.Raw(3, 1))
	}
	for r := 1; r <= 4; r++ {
		b := s.Value(r, 1).Num
		if got := s.Value(r, 2).Num; math.Abs(got-b/100) > 1e-9 {
			t.Errorf("row %d formula follows row: %v vs %v (%s)", r, got, b/100, s.Raw(r, 2))
		}
	}
	if !s.BaseStyle(4, 0).Bold || s.BaseStyle(1, 0).Bold {
		t.Error("style moves with row")
	}
	if s.Raw(0, 0) != "Region" || s.Raw(5, 1) != "99" {
		t.Error("header and outside rows untouched")
	}
	// multi-key: sort by Revenue desc
	s.Sort(rng, 1, false, true)
	if s.Raw(1, 1) != "50" || s.Raw(4, 1) != "10" {
		t.Errorf("desc sort %q..%q", s.Raw(1, 1), s.Raw(4, 1))
	}
	if !w.Undo() {
		t.Fatal("sort undo missing") // undoes the SortRange, not the bare Sort
	}
	if s.Raw(1, 0) != "South" {
		t.Errorf("undo sort: %q", s.Raw(1, 0))
	}

	// filter
	s.SetFilterHeader(0)
	s.SetFilter(0, map[string]bool{"North": true})
	hidden := 0
	for r := 0; r <= 5; r++ {
		if s.HiddenRow(r) {
			hidden++
		}
	}
	if hidden != 3 || s.HiddenRow(0) { // South, East, and the blank-region row
		t.Errorf("hidden rows = %d", hidden)
	}
	if got := s.FilterValues(0); len(got) != 4 {
		t.Errorf("filter values %v", got)
	}
	s.SetFilter(0, nil)
	if s.HiddenRow(1) || s.HasFilter() {
		t.Error("clear filter")
	}
	s.SetRowHidden(2, true)
	if !s.HiddenRow(2) {
		t.Error("manual hide")
	}
}

func TestEngineFindStatsAndMisc(t *testing.T) {
	w := NewWorkbook()
	s := w.Sheets[1]
	s.SetRaw(0, 0, "Apple")
	s.SetRaw(1, 0, "pineapple")
	s.SetRaw(2, 1, "=A1&\"!\"")
	s.SetRaw(3, 0, "10")
	s.SetRaw(4, 0, "30")
	if got := s.Find("apple", false); len(got) != 3 || got[0] != (CellRef{0, 0}) || got[2] != (CellRef{2, 1}) {
		t.Errorf("find %v", got)
	}
	if got := s.Find("apple", true); len(got) != 1 || got[0].Row != 1 {
		t.Errorf("find case %v", got)
	}
	count, nums, sum, avg, mn, mx := s.Stats(Range{CellRef{0, 0}, CellRef{9, 1}})
	if count != 5 || nums != 2 || sum != 40 || avg != 20 || mn != 10 || mx != 30 {
		t.Errorf("stats %d %d %v %v %v %v", count, nums, sum, avg, mn, mx)
	}
	if c, n, _, _, _, _ := s.Stats(Range{CellRef{20, 20}, CellRef{22, 22}}); c != 0 || n != 0 {
		t.Error("empty stats")
	}
	s.SetColWidth(2, 140)
	s.SetRowHeight(3, 40)
	if s.ColWidth(2) != 140 || s.ColWidth(3) != 84 || s.RowHeight(3) != 40 || s.RowHeight(4) != 22 {
		t.Error("sizes")
	}
	s.SetValidation(Range{CellRef{0, 4}, CellRef{5, 4}}, []string{"a", "b"})
	if l := s.ValidationList(2, 4); len(l) != 2 || s.ValidationList(2, 3) != nil {
		t.Error("validation")
	}
	s.SetComment(0, 0, "note")
	if s.Comment(0, 0) != "note" || s.Comment(1, 0) != "" {
		t.Error("comment")
	}
	if s.EffectiveAlign(3, 0) != "right" || s.EffectiveAlign(0, 0) != "left" {
		t.Error("align")
	}
	s.SetFreeze(2, 1)
	if s.FreezeRows() != 2 || s.FreezeCols() != 1 {
		t.Error("freeze")
	}
	w.AddSheet()
	if len(w.Sheets) != 4 || w.Sheets[3].Name != "Sheet4" {
		t.Errorf("add sheet %q", w.Sheets[len(w.Sheets)-1].Name)
	}
	w.RenameSheet(3, "Sheet1") // duplicate name rejected
	if w.Sheets[3].Name != "Sheet4" {
		t.Error("duplicate rename accepted")
	}
}

func TestEngineConditionalFormats(t *testing.T) {
	w := NewWorkbook()
	s := w.Sheets[1]
	for i, v := range []string{"5", "15", "25", "15", "abc"} {
		s.SetRaw(i, 0, v)
	}
	r := Range{CellRef{0, 0}, CellRef{4, 0}}
	s.AddCondFormat(r, CondRule{Kind: "gt", A: "10", Preset: "green"})
	if s.StyleAt(0, 0).BG != "" || s.StyleAt(1, 0).BG != "#C6EFCE" || s.StyleAt(4, 0).BG != "" {
		t.Error("gt rule")
	}
	s.ClearCondFormat(r)
	s.AddCondFormat(r, CondRule{Kind: "between", A: "10", B: "20", Preset: "yellow"})
	if s.StyleAt(1, 0).BG != "#FFEB9C" || s.StyleAt(2, 0).BG != "" {
		t.Error("between rule")
	}
	s.ClearCondFormat(r)
	s.AddCondFormat(r, CondRule{Kind: "dupes"})
	if s.StyleAt(1, 0).BG != "#FFC7CE" || s.StyleAt(3, 0).FG != "#9C0006" || s.StyleAt(0, 0).BG != "" {
		t.Error("dupes rule")
	}
	s.ClearCondFormat(r)
	s.AddCondFormat(r, CondRule{Kind: "top", A: "1"})
	if s.StyleAt(2, 0).BG == "" || s.StyleAt(1, 0).BG != "" {
		t.Error("top rule")
	}
	s.ClearCondFormat(r)
	s.AddCondFormat(r, CondRule{Kind: "text", A: "AB"})
	if s.StyleAt(4, 0).BG == "" || s.StyleAt(0, 0).BG != "" {
		t.Error("text rule")
	}
	s.ClearCondFormat(r)
	s.AddCondFormat(r, CondRule{Kind: "databar", Preset: "green"})
	f, col, ok := s.DataBar(2, 0)
	if !ok || f != 1 || col != "#63BE7B" {
		t.Errorf("databar max %v %v %v", f, col, ok)
	}
	if f, _, _ := s.DataBar(0, 0); math.Abs(f-0.2) > 1e-9 {
		t.Errorf("databar frac %v", f)
	}
	if _, _, ok := s.DataBar(4, 0); ok {
		t.Error("text cell has no bar")
	}
	s.SetRaw(2, 0, "50") // cache must follow edits
	if f, _, _ := s.DataBar(1, 0); math.Abs(f-0.3) > 1e-9 {
		t.Errorf("databar after edit %v", f)
	}
	s.ClearCondFormat(r)
	s.AddCondFormat(r, CondRule{Kind: "scale3"})
	lo, hi := s.StyleAt(0, 0).BG, s.StyleAt(2, 0).BG
	if lo == "" || hi == "" || lo == hi {
		t.Errorf("scale3 %q %q", lo, hi)
	}
	// base style survives merging
	s.ClearCondFormat(r)
	s.SetStyle(r, func(st *Style) { st.Bold = true })
	s.AddCondFormat(r, CondRule{Kind: "gt", A: "0"})
	if st := s.StyleAt(0, 0); !st.Bold || st.BG == "" {
		t.Error("merge with base style")
	}
}

func TestEngineUndoRedo(t *testing.T) {
	w := NewWorkbook()
	s := w.Sheets[1]
	w.Edit(s, 0, 0, "1")
	w.Edit(s, 0, 1, "=A1*2")
	w.Edit(s, 0, 0, "5")
	if s.Value(0, 1).Num != 10 || !w.CanUndo() || w.CanRedo() {
		t.Fatal("setup")
	}
	w.Undo()
	if s.Raw(0, 0) != "1" || s.Value(0, 1).Num != 2 {
		t.Fatalf("undo edit: %q %v", s.Raw(0, 0), s.Value(0, 1))
	}
	if !w.CanRedo() {
		t.Fatal("redo available")
	}
	w.Redo()
	if s.Value(0, 1).Num != 10 {
		t.Fatal("redo recalculated")
	}
	w.Undo()
	w.Undo()
	w.Undo()
	if s.Raw(0, 0) != "" || s.Raw(0, 1) != "" || w.CanUndo() || w.Undo() {
		t.Fatal("undo to empty")
	}
	w.Redo()
	w.Redo()
	if s.Raw(0, 1) != "=A1*2" || s.Value(0, 1).Num != 2 {
		t.Fatal("redo chain")
	}
	// a new edit clears redo
	w.Edit(s, 4, 4, "x")
	if w.CanRedo() {
		t.Fatal("redo cleared by new edit")
	}
	// unchanged edit records nothing
	n := len(w.hist.undo)
	w.Edit(s, 4, 4, "x")
	if len(w.hist.undo) != n {
		t.Fatal("no-op edit recorded")
	}
	// format
	w.Format(s, Range{CellRef{0, 0}, CellRef{1, 1}}, func(st *Style) { st.Bold, st.BG = true, "#FF0000" })
	if !s.BaseStyle(1, 1).Bold {
		t.Fatal("format")
	}
	w.Undo()
	if s.BaseStyle(1, 1).Bold || s.BaseStyle(0, 0).BG != "" {
		t.Fatal("undo format")
	}
	w.Redo()
	if s.BaseStyle(0, 0).BG != "#FF0000" {
		t.Fatal("redo format")
	}
	// typed literal's inferred format reverts with the edit
	w.Edit(s, 6, 0, "50%")
	if s.BaseStyle(6, 0).NumFmt != "Percent" {
		t.Fatal("percent inference")
	}
	w.Undo()
	if s.BaseStyle(6, 0).NumFmt != "" || s.Raw(6, 0) != "" {
		t.Fatal("undo inference")
	}
	// EditRange as one step
	w.EditRange(s, Range{CellRef{10, 0}, CellRef{11, 1}}, [][]string{{"a", "b"}, {"c", "=A11&B11"}})
	if s.Value(11, 1).String() != "ab" {
		t.Fatal("range edit")
	}
	w.Undo()
	if s.Raw(10, 0) != "" || s.Raw(11, 1) != "" {
		t.Fatal("range undo is one step")
	}
	// bounded stack
	for i := 0; i < 250; i++ {
		w.Edit(s, 20, 0, string(rune('a'+i%26))+string(rune('a'+i/26)))
	}
	if len(w.hist.undo) != historyLimit {
		t.Fatalf("stack %d", len(w.hist.undo))
	}
}

func TestEngineClipboardTSV(t *testing.T) {
	w := NewWorkbook()
	s := w.Sheets[1]
	w.EditRange(s, Range{CellRef{0, 0}, CellRef{1, 2}}, [][]string{{"a", "b\tc", "say \"hi\""}, {"1", "=A2+1", "multi\nline"}})
	tsv := s.CopyTSV(Range{CellRef{0, 0}, CellRef{1, 2}})
	if !strings.Contains(tsv, "\"b\tc\"") || !strings.Contains(tsv, "\"say \"\"hi\"\"\"") {
		t.Fatalf("tsv quoting %q", tsv)
	}
	got := ParseTSV(tsv)
	if len(got) != 2 || got[0][1] != "b\tc" || got[0][2] != `say "hi"` || got[1][2] != "multi\nline" {
		t.Fatalf("parse %q", got)
	}
	dst := w.PasteTSV(s, CellRef{5, 5}, tsv)
	if dst.String() != "F6:H7" {
		t.Fatalf("pasted range %s", dst)
	}
	if s.Raw(5, 6) != "b\tc" || s.Raw(6, 6) != "=A2+1" || s.Value(6, 5).Num != 1 {
		t.Fatal("paste contents")
	}
	if !w.Undo() || s.Raw(5, 5) != "" || s.Raw(6, 7) != "" {
		t.Fatal("paste undo is one step")
	}
	// external TSV with CRLF and no trailing newline
	r := w.PasteTSV(s, CellRef{0, 10}, "x\ty\r\n3\t4.5")
	if r.String() != "K1:L2" || s.Value(1, 11).Num != 4.5 {
		t.Fatalf("external tsv %s", r)
	}
	w.ClearRange(s, Range{CellRef{0, 10}, CellRef{1, 11}})
	if s.Raw(1, 11) != "" {
		t.Fatal("clear")
	}
	w.Undo()
	if s.Raw(1, 11) != "4.5" {
		t.Fatal("clear undo")
	}
}

func TestEngineRichPasteAndAutoFill(t *testing.T) {
	w := NewWorkbook()
	s := w.Sheets[1]
	w.EditRange(s, Range{CellRef{0, 0}, CellRef{1, 2}}, [][]string{{"1", "2", "3"}, {"=A1+1", "=B1+1", "=C1+1"}})
	w.Format(s, Range{CellRef{0, 0}, CellRef{0, 0}}, func(st *Style) { st.Bold = true })
	buf := s.CopyRich(Range{CellRef{0, 0}, CellRef{1, 2}})
	w.Paste(s, CellRef{4, 1}, buf)
	if s.Raw(5, 1) != "=B5+1" || s.Value(5, 3).Num != 4 || !s.BaseStyle(4, 1).Bold {
		t.Fatalf("paste shifts formulas: %q %v", s.Raw(5, 1), s.Value(5, 3))
	}
	// values only
	w.PasteSpecial(s, CellRef{8, 0}, buf, PasteValues, false, false)
	if s.Raw(9, 0) != "2" || s.BaseStyle(8, 0).Bold {
		t.Fatalf("values only: %q", s.Raw(9, 0))
	}
	// transpose 2x3 -> 3x2
	r := w.PasteSpecial(s, CellRef{12, 0}, buf, PasteValues, false, true)
	if r.Rows() != 3 || r.Cols() != 2 || s.Value(13, 0).Num != 2 || s.Value(14, 1).Num != 4 {
		t.Fatalf("transpose %s %v", r, s.Value(14, 1))
	}
	// skip blanks keeps destination
	s.SetRaw(20, 0, "keep")
	blank := &ClipBuf{Src: Range{CellRef{0, 0}, CellRef{0, 1}}, Sheet: s, Cells: [][]ClipCell{{{}, {Raw: "new"}}}}
	w.PasteSpecial(s, CellRef{20, 0}, blank, PasteAll, true, false)
	if s.Raw(20, 0) != "keep" || s.Raw(20, 1) != "new" {
		t.Fatal("skip blanks")
	}
	// formats only
	w.PasteSpecial(s, CellRef{30, 0}, buf, PasteFormats, false, false)
	if s.Raw(30, 0) != "" || !s.BaseStyle(30, 0).Bold {
		t.Fatal("formats only")
	}
	// cut moves
	cut := s.CopyRich(Range{CellRef{0, 0}, CellRef{0, 2}})
	cut.Cut = true
	w.Paste(s, CellRef{40, 0}, cut)
	if s.Raw(0, 0) != "" || s.Raw(40, 2) != "3" {
		t.Fatal("cut and move")
	}
	w.Undo()
	if s.Raw(0, 0) != "1" || s.Raw(40, 2) != "" {
		t.Fatal("cut undo")
	}

	// autofill series
	f := NewWorkbook().Sheets[1]
	fw := f.wb
	fw.EditRange(f, Range{CellRef{0, 0}, CellRef{1, 0}}, [][]string{{"1"}, {"2"}})
	fw.AutoFill(f, Range{CellRef{0, 0}, CellRef{1, 0}}, CellRef{4, 0})
	if f.Raw(2, 0) != "3" || f.Raw(4, 0) != "5" {
		t.Fatalf("numeric series %q %q", f.Raw(2, 0), f.Raw(4, 0))
	}
	fw.EditRange(f, Range{CellRef{0, 2}, CellRef{0, 2}}, [][]string{{"Jan"}})
	fw.AutoFill(f, Range{CellRef{0, 2}, CellRef{0, 2}}, CellRef{0, 5})
	if f.Raw(0, 3) != "Feb" || f.Raw(0, 5) != "Apr" {
		t.Fatalf("month series %q %q", f.Raw(0, 3), f.Raw(0, 5))
	}
	fw.EditRange(f, Range{CellRef{3, 6}, CellRef{3, 6}}, [][]string{{"2026-11-25"}})
	fw.AutoFill(f, Range{CellRef{3, 6}, CellRef{3, 6}}, CellRef{5, 6})
	if f.Display(5, 6) != "2026-11-27" {
		t.Fatalf("date series %q", f.Display(5, 6))
	}
	fw.EditRange(f, Range{CellRef{0, 8}, CellRef{1, 8}}, [][]string{{"5"}, {"=I1*2"}})
	fw.AutoFill(f, Range{CellRef{1, 8}, CellRef{1, 8}}, CellRef{3, 8})
	if f.Raw(3, 8) != "=I3*2" {
		t.Fatalf("formula fill %q", f.Raw(3, 8))
	}
	fw.EditRange(f, Range{CellRef{0, 10}, CellRef{0, 10}}, [][]string{{"Item 1"}})
	fw.AutoFill(f, Range{CellRef{0, 10}, CellRef{0, 10}}, CellRef{0, 12})
	if f.Raw(0, 12) != "Item 3" {
		t.Fatalf("text-number %q", f.Raw(0, 12))
	}
	fw.EditRange(f, Range{CellRef{5, 14}, CellRef{5, 14}}, [][]string{{"x"}})
	fw.AutoFill(f, Range{CellRef{5, 14}, CellRef{5, 14}}, CellRef{7, 14})
	if f.Raw(7, 14) != "x" {
		t.Fatal("copy fill")
	}
	fw.EditRange(f, Range{CellRef{10, 0}, CellRef{11, 0}}, [][]string{{"10"}, {"20"}})
	fw.AutoFill(f, Range{CellRef{10, 0}, CellRef{11, 0}}, CellRef{8, 0})
	if f.Raw(9, 0) != "0" || f.Raw(8, 0) != "-10" {
		t.Fatalf("upward fill %q %q", f.Raw(9, 0), f.Raw(8, 0))
	}
	fw.Undo()
	if f.Raw(9, 0) != "" {
		t.Fatal("autofill undo")
	}
}

func xlSameValues(t *testing.T, a, b *Sheet) {
	t.Helper()
	ua, ub := a.UsedRange(), b.UsedRange()
	if ua != ub {
		t.Fatalf("used range %s vs %s", ua, ub)
	}
	for r := 0; r <= ua.Max.Row; r++ {
		for c := 0; c <= ua.Max.Col; c++ {
			if a.Raw(r, c) != b.Raw(r, c) || a.Display(r, c) != b.Display(r, c) || a.StyleAt(r, c) != b.StyleAt(r, c) {
				t.Fatalf("%s differs: %q/%q vs %q/%q", CellRef{r, c}.A1(), a.Raw(r, c), a.Display(r, c), b.Raw(r, c), b.Display(r, c))
			}
		}
	}
}

func xlRichWorkbook() *Workbook {
	w := NewWorkbook()
	s := w.Sheets[0]
	s.AddCondFormat(Range{CellRef{1, 2}, CellRef{8, 2}}, CondRule{Kind: "gt", A: "10000", Preset: "green"})
	s.Merge(Range{CellRef{12, 0}, CellRef{12, 3}})
	s.SetRaw(12, 0, "Merged \"title\", with comma")
	s.SetRaw(13, 0, "'00123")
	s.SetRaw(13, 1, "2026-11-25")
	s.SetRaw(13, 2, "12.5%")
	s.SetRaw(13, 3, "TRUE")
	s.SetComment(1, 0, "note")
	s.SetValidation(Range{CellRef{1, 1}, CellRef{8, 1}}, []string{"Widget", "Gadget"})
	s.SetFreeze(1, 2)
	w.Sheets[1].SetRaw(0, 0, "=Sheet1!G10*2")
	w.Active = 1
	return w
}

func TestEngineJSONRoundtrip(t *testing.T) {
	w := xlRichWorkbook()
	b, err := w.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	w2, err := UnmarshalWorkbook(b)
	if err != nil {
		t.Fatal(err)
	}
	if w2.Active != 1 || len(w2.Sheets) != 3 || w2.Sheets[1].Name != "Sheet2" {
		t.Fatal("sheets")
	}
	for i := range w.Sheets {
		xlSameValues(t, w.Sheets[i], w2.Sheets[i])
	}
	s := w2.Sheets[0]
	if s.ColWidth(2) != 100 || s.FreezeRows() != 1 || s.FreezeCols() != 2 || s.Comment(1, 0) != "note" ||
		len(s.Merges()) != 1 || len(s.ValidationList(2, 1)) != 2 || s.StyleAt(1, 2).BG != "#C6EFCE" {
		t.Errorf("lost workbook features: %d %d %q %d %d %q", s.ColWidth(2), s.FreezeCols(), s.Comment(1, 0), len(s.Merges()), len(s.ValidationList(2, 1)), s.StyleAt(1, 2).BG)
	}
	if v := w2.Sheets[1].Value(0, 0); v.Num != w.Sheets[1].Value(0, 0).Num || v.Num == 0 {
		t.Errorf("cross-sheet formula after load: %+v", v)
	}
	if _, err := UnmarshalWorkbook([]byte(`{"Sheets":[]}`)); err == nil {
		t.Error("empty workbook accepted")
	}
	if _, err := UnmarshalWorkbook([]byte(`nope`)); err == nil {
		t.Error("garbage accepted")
	}
}

func TestEngineCSVRoundtrip(t *testing.T) {
	w := xlRichWorkbook()
	s := w.Sheets[0]
	var buf bytes.Buffer
	if err := s.ExportCSV(&buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"Merged ""title"", with comma"`) {
		t.Fatalf("csv quoting:\n%s", buf.String())
	}
	if !strings.HasPrefix(buf.String(), "Region,Product,Q1,Q2,Q3,Q4,Total,Growth\n") {
		t.Fatalf("csv header: %q", buf.String()[:60])
	}
	if err := w.ImportCSV(strings.NewReader(buf.String()), "Back"); err != nil {
		t.Fatal(err)
	}
	back := w.SheetByName("Back")
	if back == nil || w.Sheet() != back {
		t.Fatal("imported sheet missing or not active")
	}
	u := s.UsedRange()
	for r := 0; r <= u.Max.Row; r++ {
		for c := 0; c <= u.Max.Col; c++ {
			if a, b := s.Display(r, c), back.Display(r, c); a != b {
				t.Errorf("%s: %q vs %q", CellRef{r, c}.A1(), a, b)
			}
		}
	}
	if back.Raw(13, 0) != "'00123" || back.Value(13, 0).Str != "00123" {
		t.Errorf("leading zero text lost: %q", back.Raw(13, 0))
	}
	var f bytes.Buffer
	_ = s.ExportCSVFormulas(&f)
	if !strings.Contains(f.String(), "=SUM(C2:F2)") {
		t.Error("formula export")
	}
	if err := w.ImportCSV(strings.NewReader("a\tb\n1\t2\n"), "Tabbed"); err != nil {
		t.Fatal(err)
	}
	if v := w.SheetByName("Tabbed").Value(1, 1); v.Num != 2 {
		t.Errorf("tsv import %+v", v)
	}
	if err := w.ImportCSV(strings.NewReader("a,b\n1,2\n"), "Back"); err != nil || w.SheetByName("Back (2)") == nil {
		t.Error("duplicate import name should be uniquified")
	}
}

func TestEngineXLSXRoundtrip(t *testing.T) {
	w := xlRichWorkbook()
	w.Sheets[0].SetRaw(14, 0, "line1 & <b>")
	w.Sheets[0].SetRaw(14, 1, "=A15&\"!\"")
	w.Sheets[0].SetRaw(14, 2, "=1/0")
	w.Sheets[0].SetStyle(Range{CellRef{14, 0}, CellRef{14, 0}}, func(st *Style) {
		st.Bold, st.Italic, st.FG, st.BG, st.Align, st.Wrap, st.FontSize = true, true, "#112233", "#FFEEDD", "center", true, 14
	})
	path := filepath.Join(t.TempDir(), "book.xlsx")
	if err := w.SaveXLSX(path); err != nil {
		t.Fatal(err)
	}
	w2, err := LoadXLSX(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(w2.Sheets) != 3 || w2.Active != 1 {
		t.Fatalf("sheets %d active %d", len(w2.Sheets), w2.Active)
	}
	a, b := w.Sheets[0], w2.Sheets[0]
	u := a.UsedRange()
	for r := 0; r <= u.Max.Row; r++ {
		for c := 0; c <= u.Max.Col; c++ {
			if a.IsFormula(r, c) && a.Raw(r, c) != b.Raw(r, c) {
				t.Errorf("%s raw %q vs %q", CellRef{r, c}.A1(), a.Raw(r, c), b.Raw(r, c))
			}
			if a.Display(r, c) != b.Display(r, c) {
				t.Errorf("%s display %q vs %q", CellRef{r, c}.A1(), a.Display(r, c), b.Display(r, c))
			}
			sa, sb := a.BaseStyle(r, c), b.BaseStyle(r, c)
			if sa != sb {
				t.Errorf("%s style %+v vs %+v", CellRef{r, c}.A1(), sa, sb)
			}
		}
	}
	if b.ColWidth(2) != 100 || b.FreezeRows() != 1 || b.FreezeCols() != 2 || len(b.Merges()) != 1 {
		t.Errorf("layout: w=%d freeze=%d,%d merges=%d", b.ColWidth(2), b.FreezeRows(), b.FreezeCols(), len(b.Merges()))
	}
	if w2.Sheets[1].Value(0, 0).Num != w.Sheets[1].Value(0, 0).Num {
		t.Error("cross-sheet formula")
	}
	if _, err := LoadXLSX(filepath.Join(t.TempDir(), "missing.xlsx")); err == nil {
		t.Error("missing file accepted")
	}
	bad := filepath.Join(t.TempDir(), "bad.xlsx")
	_ = os.WriteFile(bad, []byte("not a zip"), savedFileMode)
	if _, err := LoadXLSX(bad); err == nil {
		t.Error("garbage accepted")
	}
}
