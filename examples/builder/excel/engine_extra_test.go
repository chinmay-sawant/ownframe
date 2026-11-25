package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

func numSheet(vals ...float64) *Sheet {
	w := NewWorkbook()
	s := w.Sheets[1]
	for i, v := range vals {
		s.SetRaw(i, 0, xlValueRaw(xlNumV(v)))
	}
	return s
}

func TestEngineCondBottomAboveBelow(t *testing.T) {
	s := numSheet(5, 1, 9, 3, 7)
	s.SetRaw(5, 0, "text")
	rng := Range{CellRef{0, 0}, CellRef{6, 0}}
	hit := func(i int) bool { return s.StyleAt(i, 0).BG != "" }

	s.AddCondFormat(rng, CondRule{Kind: "bottom", A: "2", Preset: "yellow"})
	want := []bool{false, true, false, true, false, false, false} // 1 and 3
	for i, w := range want {
		if hit(i) != w {
			t.Errorf("bottom 2: row %d hit=%v want %v", i, hit(i), w)
		}
	}
	s.ClearCondFormat(rng)

	s.AddCondFormat(rng, CondRule{Kind: "above", Preset: "green"}) // average 5
	want = []bool{false, false, true, false, true, false, false}
	for i, w := range want {
		if hit(i) != w {
			t.Errorf("above: row %d hit=%v want %v", i, hit(i), w)
		}
	}
	s.ClearCondFormat(rng)

	s.AddCondFormat(rng, CondRule{Kind: "below", Preset: "red"})
	want = []bool{false, true, false, true, false, false, false}
	for i, w := range want {
		if hit(i) != w {
			t.Errorf("below: row %d hit=%v want %v", i, hit(i), w)
		}
	}
	// an empty range never matches
	s2 := numSheet()
	s2.AddCondFormat(rng, CondRule{Kind: "above"})
	if s2.StyleAt(0, 0).BG != "" {
		t.Error("above-average matched an empty cell")
	}
	// bottom N larger than the data matches everything numeric
	s3 := numSheet(4, 2)
	s3.AddCondFormat(rng, CondRule{Kind: "bottom", A: "10"})
	if s3.StyleAt(0, 0).BG == "" || s3.StyleAt(1, 0).BG == "" {
		t.Error("bottom 10 of two values should match both")
	}
}

func TestEngineStyleAtDataBarScale3(t *testing.T) {
	s := numSheet(0, 50, 100)
	rng := Range{CellRef{0, 0}, CellRef{2, 0}}
	s.AddCondFormat(rng, CondRule{Kind: "scale3"})
	if got := s.StyleAt(0, 0).BG; got != "#F8696B" {
		t.Errorf("scale3 low = %s", got)
	}
	if got := s.StyleAt(1, 0).BG; got != "#FFEB84" {
		t.Errorf("scale3 mid = %s", got)
	}
	if got := s.StyleAt(2, 0).BG; got != "#63BE7B" {
		t.Errorf("scale3 high = %s", got)
	}
	// halfway between low and mid
	s.SetRaw(3, 0, "25")
	s.ClearCondFormat(rng)
	s.AddCondFormat(Range{CellRef{0, 0}, CellRef{3, 0}}, CondRule{Kind: "scale3"})
	if got := s.StyleAt(3, 0).BG; got == "#F8696B" || got == "#FFEB84" || got == "" {
		t.Errorf("scale3 25 should interpolate, got %q", got)
	}
	s.ClearCondFormat(Range{CellRef{0, 0}, CellRef{3, 0}})

	s.AddCondFormat(rng, CondRule{Kind: "databar", Preset: "green"})
	for row, wantFrac := range map[int]float64{0: 0, 1: 0.5, 2: 1} {
		frac, color, ok := s.DataBar(row, 0)
		if !ok || frac != wantFrac || color != "#63BE7B" {
			t.Errorf("databar row %d = %v %q %v", row, frac, color, ok)
		}
	}
	if _, _, ok := s.DataBar(5, 0); ok {
		t.Error("databar outside the range")
	}
	s.SetCondOff(0, true)
	if _, _, ok := s.DataBar(1, 0); ok {
		t.Error("disabled databar still reports")
	}
	// Display keeps the number format under a rule
	s.SetStyle(rng, func(st *Style) { st.NumFmt, st.Decimals = "Number", 1 })
	if got := s.Display(1, 0); got != "50.0" {
		t.Errorf("Display under rules = %q", got)
	}
}

func TestEngineCondRulesManage(t *testing.T) {
	s := numSheet(1, 2, 3)
	r := Range{CellRef{0, 0}, CellRef{2, 0}}
	s.AddCondFormat(r, CondRule{Kind: "gt", A: "1"})
	s.AddCondFormat(r, CondRule{Kind: "lt", A: "3"})
	s.MoveCondUp(1)
	if got := s.CondRules(); got[0].Rule.Kind != "lt" || got[1].Rule.Kind != "gt" {
		t.Errorf("order = %+v", got)
	}
	s.DeleteCond(0)
	if got := s.CondRules(); len(got) != 1 || got[0].Rule.Kind != "gt" {
		t.Errorf("after delete = %+v", got)
	}
}

func TestEngineStructuralShiftsCondAndValidation(t *testing.T) {
	w := NewWorkbook()
	s := w.Sheets[1]
	s.AddCondFormat(Range{CellRef{4, 0}, CellRef{8, 0}}, CondRule{Kind: "gt", A: "1"})
	s.AddCondFormat(Range{CellRef{2, 1}, CellRef{3, 1}}, CondRule{Kind: "lt", A: "1"})
	s.SetValidation(Range{CellRef{4, 2}, CellRef{6, 2}}, []string{"a", "b"})
	s.SetValidationBounds(Range{CellRef{0, 3}, CellRef{0, 5}}, "whole", 1, 5)
	s.SetRowHidden(6, true)

	w.InsertRows(s, 2, 2)
	rules := s.CondRules()
	if rules[0].Range.String() != "A7:A11" || rules[1].Range.String() != "B5:B6" {
		t.Errorf("insert rows cond ranges: %v %v", rules[0].Range, rules[1].Range)
	}
	if ranges, _ := s.ValidationRules(); ranges[0].String() != "C7:C9" {
		t.Errorf("insert rows validation range: %v", ranges)
	}
	if !s.HiddenRow(8) || s.HiddenRow(6) {
		t.Error("hidden row did not shift")
	}

	w.DeleteRows(s, 4, 2) // rows 5-6 (0-based 4,5)
	if got := s.CondRules()[0].Range.String(); got != "A5:A9" {
		t.Errorf("delete rows cond range = %s", got)
	}
	w.DeleteRows(s, 4, 20) // rows 5 down: both rules and the list validation vanish
	if len(s.CondRules()) != 0 {
		t.Errorf("rules should vanish with their rows: %v", s.CondRules())
	}

	w.InsertCols(s, 0, 1)
	if ranges, _ := s.ValidationRules(); len(ranges) != 1 || ranges[0].String() != "E1:G1" {
		t.Errorf("insert cols validation: %v", ranges)
	}
	w.Undo()
	if ranges, _ := s.ValidationRules(); ranges[0].String() != "D1:F1" {
		t.Errorf("undo insert cols validation: %v", ranges)
	}
	w.Undo() // the delete of all rows
	// the B rule died with rows 5-6 in the first delete; the A rule is back
	if rules := s.CondRules(); len(rules) != 1 || rules[0].Range.String() != "A5:A9" {
		t.Errorf("undo should restore the A rule, got %v", rules)
	}
}

func TestEngineCutPasteRewritesDependents(t *testing.T) {
	w := NewWorkbook()
	s := w.Sheets[1]
	s.SetRaw(0, 0, "10")
	s.SetRaw(1, 0, "=A1*2")  // A2 depends on A1
	s.SetRaw(0, 2, "=A1+A2") // C1 depends on both
	w.Sheets[2].SetRaw(0, 0, "=Sheet2!A1+1")
	w.ClearHistory()

	buf := s.CopyRich(Range{CellRef{0, 0}, CellRef{1, 0}}) // A1:A2
	buf.Cut = true
	dst := w.PasteSpecial(s, CellRef{0, 4}, buf, PasteAll, false, false) // to E1:E2
	if dst.String() != "E1:E2" {
		t.Fatalf("dst = %s", dst)
	}
	if s.Raw(0, 0) != "" || s.Raw(1, 0) != "" {
		t.Error("source not cleared")
	}
	if got := s.Raw(1, 4); got != "=E1*2" {
		t.Errorf("moved formula = %q, want =E1*2 (refs into the block follow it)", got)
	}
	if got := s.Raw(0, 2); got != "=E1+E2" {
		t.Errorf("dependent formula = %q, want =E1+E2", got)
	}
	if got := w.Sheets[2].Raw(0, 0); got != "=Sheet2!E1+1" {
		t.Errorf("cross-sheet dependent = %q", got)
	}
	if s.Value(0, 2).Num != 30 {
		t.Errorf("C1 = %v", s.Value(0, 2))
	}
	if !w.Undo() {
		t.Fatal("cut paste is not undoable")
	}
	if s.Raw(0, 0) != "10" || s.Raw(0, 2) != "=A1+A2" || s.Raw(0, 4) != "" {
		t.Errorf("undo: A1=%q C1=%q E1=%q", s.Raw(0, 0), s.Raw(0, 2), s.Raw(0, 4))
	}
	w.Redo()
	if s.Raw(0, 2) != "=E1+E2" {
		t.Errorf("redo C1 = %q", s.Raw(0, 2))
	}

	// a plain copy still shifts relative references
	s.SetRaw(5, 0, "=E1+1")
	cp := s.CopyRich(Range{CellRef{5, 0}, CellRef{5, 0}})
	w.PasteSpecial(s, CellRef{6, 0}, cp, PasteAll, false, false)
	if got := s.Raw(6, 0); got != "=E2+1" {
		t.Errorf("copy paste = %q", got)
	}
}

func TestEngineSheetOpsAreUndoable(t *testing.T) {
	w := NewWorkbook()
	w.ClearHistory()
	n := len(w.Sheets)
	w.AddSheet()
	if len(w.Sheets) != n+1 || !w.CanUndo() {
		t.Fatal("add sheet")
	}
	w.Undo()
	if len(w.Sheets) != n {
		t.Errorf("undo add: %d sheets", len(w.Sheets))
	}
	w.Redo()
	if len(w.Sheets) != n+1 {
		t.Errorf("redo add: %d sheets", len(w.Sheets))
	}

	w.Sheets[0].SetRaw(20, 0, "=Sheet2!A1")
	w.RenameSheet(1, "Data")
	if w.Sheets[1].Name != "Data" || w.Sheets[0].Raw(20, 0) != "=Data!A1" {
		t.Fatalf("rename: %q %q", w.Sheets[1].Name, w.Sheets[0].Raw(20, 0))
	}
	w.Undo()
	if w.Sheets[1].Name != "Sheet2" || w.Sheets[0].Raw(20, 0) != "=Sheet2!A1" {
		t.Errorf("undo rename: %q %q", w.Sheets[1].Name, w.Sheets[0].Raw(20, 0))
	}

	w.Sheets[1].SetRaw(0, 0, "7")
	w.DeleteSheet(1)
	if w.SheetByName("Sheet2") != nil || w.Sheets[0].Value(20, 0).Kind != KindError {
		t.Fatalf("delete sheet: %v", w.Sheets[0].Value(20, 0))
	}
	w.Undo()
	if w.SheetByName("Sheet2") == nil || w.Sheets[0].Value(20, 0).Num != 7 {
		t.Errorf("undo delete: formula = %v", w.Sheets[0].Value(20, 0))
	}
	w.RenameSheet(0, "Sheet2") // clash: no change, no history entry
	before := len(w.hist.undo)
	w.RenameSheet(0, "Sheet2")
	if len(w.hist.undo) != before {
		t.Error("rejected rename recorded history")
	}
}

func TestEngineRowAutoGrowsForWrappedText(t *testing.T) {
	w := NewWorkbook()
	s := w.Sheets[1]
	long := "alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu nu xi omicron"
	s.SetRaw(2, 0, long)
	if s.RowHeight(2) != defaultRowHeight {
		t.Fatalf("unwrapped row = %d", s.RowHeight(2))
	}
	s.SetStyle(Range{CellRef{2, 0}, CellRef{2, 0}}, func(st *Style) { st.Wrap = true })
	grown := s.RowHeight(2)
	if grown <= defaultRowHeight {
		t.Fatalf("wrapped row = %d, want > %d", grown, defaultRowHeight)
	}
	s.SetColWidth(0, 400) // a wide column needs fewer lines
	if s.RowHeight(2) >= grown {
		t.Errorf("wider column should shrink the row: %d vs %d", s.RowHeight(2), grown)
	}
	s.SetColWidth(0, 84)
	s.SetRowHeight(2, 30) // a manual height wins
	if s.RowHeight(2) != 30 {
		t.Errorf("manual height = %d", s.RowHeight(2))
	}
	s.SetRaw(3, 0, "short")
	s.SetStyle(Range{CellRef{3, 0}, CellRef{3, 0}}, func(st *Style) { st.Wrap = true })
	if s.RowHeight(3) != defaultRowHeight {
		t.Errorf("one-line wrapped row = %d", s.RowHeight(3))
	}
	s.SetRaw(4, 0, "one\ntwo\nthree")
	s.SetStyle(Range{CellRef{4, 0}, CellRef{4, 0}}, func(st *Style) { st.Wrap = true })
	if s.RowHeight(4) < 3*wrapLineHeight {
		t.Errorf("three-line cell row = %d", s.RowHeight(4))
	}
}

func TestEngineXLSXRoundtripRichStyles(t *testing.T) {
	w := NewWorkbook()
	s := w.Sheets[0]
	s.SetRaw(15, 0, "2026-11-25")
	s.SetRaw(15, 1, "12.5%")
	s.SetRaw(15, 2, "$1,250.75")
	s.SetRaw(15, 3, "$5")
	s.SetRaw(15, 4, "0.5")
	s.SetStyle(Range{CellRef{15, 4}, CellRef{15, 4}}, func(st *Style) { st.NumFmt, st.Decimals = "Percent", 2 })
	s.SetRaw(15, 5, "46351.75")
	s.SetStyle(Range{CellRef{15, 5}, CellRef{15, 5}}, func(st *Style) { st.NumFmt = "Time" })
	s.SetRaw(15, 6, "46351")
	s.SetStyle(Range{CellRef{15, 6}, CellRef{15, 6}}, func(st *Style) { st.NumFmt = "LongDate" })
	s.SetRaw(16, 0, "wrapped")
	s.SetStyle(Range{CellRef{16, 0}, CellRef{16, 0}}, func(st *Style) {
		st.FontFamily, st.VAlign, st.Indent, st.Wrap = "mono", "middle", 3, true
	})
	w.ApplyBorder(s, Range{CellRef{17, 0}, CellRef{18, 1}}, BorderThick)
	w.ApplyBorder(s, Range{CellRef{19, 0}, CellRef{19, 1}}, BorderTotal)
	s.AddCondFormat(Range{CellRef{1, 2}, CellRef{8, 2}}, CondRule{Kind: "between", A: "9000", B: "12000", Preset: "yellow"})
	s.AddCondFormat(Range{CellRef{1, 3}, CellRef{8, 3}}, CondRule{Kind: "bottom", A: "3", Preset: "green"})
	s.AddCondFormat(Range{CellRef{1, 4}, CellRef{8, 4}}, CondRule{Kind: "databar", Preset: "green"})
	s.AddCondFormat(Range{CellRef{1, 5}, CellRef{8, 5}}, CondRule{Kind: "scale3"})
	s.AddCondFormat(Range{CellRef{1, 0}, CellRef{8, 0}}, CondRule{Kind: "text", A: "orth", Preset: "red"})
	s.AddCondFormat(Range{CellRef{1, 1}, CellRef{8, 1}}, CondRule{Kind: "dupes", Preset: "red"})
	s.AddCondFormat(Range{CellRef{1, 6}, CellRef{8, 6}}, CondRule{Kind: "above", Preset: "green"})
	s.AddCondFormat(Range{CellRef{1, 7}, CellRef{8, 7}}, CondRule{Kind: "top", A: "2", Preset: "red"})
	s.SetValidation(Range{CellRef{1, 8}, CellRef{5, 8}}, []string{"Yes", "No"})
	s.SetValidationBounds(Range{CellRef{1, 9}, CellRef{5, 9}}, "decimal", 0.5, 99.5)
	s.SetValidationBounds(Range{CellRef{1, 10}, CellRef{5, 10}}, "length", 1, 4)
	s.SetComment(3, 3, "check <this> & that")
	s.SetRowHidden(4, true)
	s.SetColHidden(9, true)
	s.TabColor = "#FF8800"

	path := filepath.Join(t.TempDir(), "rich.xlsx")
	if err := w.SaveXLSX(path); err != nil {
		t.Fatal(err)
	}
	w2, err := LoadXLSX(path)
	if err != nil {
		t.Fatal(err)
	}
	b := w2.Sheets[0]
	for ref, want := range map[string]string{"A16": "2026-11-25", "B16": "12.5%", "C16": "$1,250.75", "D16": "$5", "E16": "50.00%", "F16": "18:00:00", "G16": "Wed, Nov 25, 2026"} {
		r, _ := ParseA1(ref)
		if got := b.Display(r.Row, r.Col); got != want {
			t.Errorf("%s display %q want %q", ref, got, want)
		}
		sa, sb := s.BaseStyle(r.Row, r.Col), b.BaseStyle(r.Row, r.Col)
		// 2 places on Currency is the format default, which reads back as 0
		sa.Decimals, sb.Decimals = effectiveDecimals(sa), effectiveDecimals(sb)
		if sa != sb {
			t.Errorf("%s style %+v vs %+v", ref, sa, sb)
		}
	}
	if got := b.BaseStyle(16, 0); got.FontFamily != "mono" || got.VAlign != "middle" || got.Indent != 3 || !got.Wrap {
		t.Errorf("A17 style = %+v", got)
	}
	if e := b.BaseStyle(17, 0).Edges; e != (Edges{Left: EdgeThick, Top: EdgeThick}) {
		t.Errorf("A18 edges = %+v", e)
	}
	if e := b.BaseStyle(19, 1).Edges; e != (Edges{Top: EdgeThin, Bottom: EdgeDouble}) {
		t.Errorf("B20 edges = %+v", e)
	}
	got, want := b.CondRules(), s.CondRules()
	if len(got) != len(want) {
		t.Fatalf("cond rules %d vs %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Range != want[i].Range || got[i].Rule != want[i].Rule {
			t.Errorf("rule %d: %+v vs %+v", i, got[i], want[i])
		}
	}
	gr, gk := b.ValidationRules()
	wr, wk := s.ValidationRules()
	if !reflect.DeepEqual(gr, wr) || !reflect.DeepEqual(gk, wk) {
		t.Errorf("validation %v %v vs %v %v", gr, gk, wr, wk)
	}
	if !reflect.DeepEqual(b.ValidationList(1, 8), []string{"Yes", "No"}) {
		t.Errorf("list = %v", b.ValidationList(1, 8))
	}
	if ok, _ := b.CheckValidation(1, 9, "100"); ok {
		t.Error("decimal bound lost")
	}
	if b.Comment(3, 3) != "check <this> & that" {
		t.Errorf("comment = %q", b.Comment(3, 3))
	}
	if !b.HiddenRow(4) || !b.ColHidden(9) || b.TabColor != "#FF8800" {
		t.Errorf("hidden/tab: row=%v col=%v tab=%q", b.HiddenRow(4), b.ColHidden(9), b.TabColor)
	}
}

func TestEngineJSONRoundtripNewFields(t *testing.T) {
	w := NewWorkbook()
	s := w.Sheets[0]
	s.SetStyle(Range{CellRef{1, 1}, CellRef{1, 1}}, func(st *Style) { st.FontFamily, st.VAlign, st.Indent = "serif", "top", 2 })
	w.ApplyBorder(s, Range{CellRef{3, 3}, CellRef{3, 3}}, BorderAll)
	s.SetValidationBounds(Range{CellRef{5, 5}, CellRef{6, 5}}, "whole", 1, 9)
	s.AddCondFormat(Range{CellRef{1, 2}, CellRef{8, 2}}, CondRule{Kind: "below", Preset: "red"})
	s.SetCondOff(0, true)
	b, err := w.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	w2, err := UnmarshalWorkbook(b)
	if err != nil {
		t.Fatal(err)
	}
	o := w2.Sheets[0]
	if st := o.BaseStyle(1, 1); st.FontFamily != "serif" || st.VAlign != "top" || st.Indent != 2 {
		t.Errorf("style = %+v", st)
	}
	if e := o.BaseStyle(3, 3).Edges; e != (Edges{EdgeThin, EdgeThin, EdgeThin, EdgeThin}) {
		t.Errorf("edges = %+v", e)
	}
	if ok, _ := o.CheckValidation(5, 5, "10"); ok {
		t.Error("whole-number bound lost in JSON")
	}
	if rules := o.CondRules(); len(rules) != 1 || !rules[0].Off || rules[0].Rule.Kind != "below" {
		t.Errorf("rules = %+v", rules)
	}
}

func TestEngineNamedFormatsSwitchOnStyle(t *testing.T) {
	for _, c := range []struct {
		fmt  string
		want string
	}{
		{"Date", "2026-11-25"}, {"LongDate", "Wed, Nov 25, 2026"}, {"Time", "00:00:00"},
	} {
		if got := FormatValue(xlNumV(46351), Style{NumFmt: c.fmt}); got != c.want {
			t.Errorf("%s = %q want %q", c.fmt, got, c.want)
		}
	}
	if code, _ := xlStyleFromCode("m/d/yyyy"); code != "Date" {
		t.Errorf("builtin short date = %q", code)
	}
}

func TestStarterSummarySheet(t *testing.T) {
	w := NewWorkbook()
	s := w.SheetByName("Summary")
	if s == nil {
		t.Fatal("starter workbook has no Summary sheet")
	}
	total := w.Sheets[0].Value(9, 6).Num // Sheet1 grand total (G10)
	if got := s.Value(2, 1).Num; got != total {
		t.Errorf("Summary total %v, Sheet1 total %v", got, total)
	}
	if s.Display(5, 1) == "" || s.Value(5, 1).Kind != KindNumber {
		t.Errorf("growth cell = %q", s.Display(5, 1))
	}
	if got := s.Value(6, 1).Str; got != "On track" && got != "Review" {
		t.Errorf("status = %q", got)
	}
	var sum float64
	for r := 9; r <= 12; r++ {
		sum += s.Value(r, 1).Num
	}
	if sum != total {
		t.Errorf("quarter table sums to %v, want %v", sum, total)
	}
	w.Sheets[0].SetRaw(1, 2, "99999")
	if s.Value(9, 1).Num != w.Sheets[0].Value(9, 2).Num {
		t.Errorf("Summary did not recalc after a Sheet1 edit")
	}
	if _, _, ok := s.DataBar(10, 1); !ok {
		t.Errorf("quarter revenue has no data bar")
	}
}
