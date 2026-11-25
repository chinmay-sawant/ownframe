package main

import (
	"fmt"
	"testing"
)

func TestEngineDepGraphChainAndDiamond(t *testing.T) {
	w := NewWorkbook()
	s := w.Sheets[1]
	s.SetRaw(0, 0, "1")
	// deep chain: A2 = A1+1, A3 = A2+1, ... A300
	for r := 1; r < 300; r++ {
		s.SetRaw(r, 0, fmt.Sprintf("=A%d+1", r))
	}
	if v := s.Value(299, 0); v.Num != 300 {
		t.Fatalf("chain end = %+v", v)
	}
	s.SetRaw(0, 0, "10")
	if v := s.Value(299, 0); v.Num != 309 {
		t.Fatalf("chain after edit = %+v", v)
	}
	// diamond: B1 -> C1, D1 -> E1
	s.SetRaw(0, 1, "2")
	s.SetRaw(0, 2, "=B1*2")
	s.SetRaw(0, 3, "=B1*3")
	s.SetRaw(0, 4, "=C1+D1")
	if v := s.Value(0, 4); v.Num != 10 {
		t.Fatalf("diamond = %+v", v)
	}
	s.SetRaw(0, 1, "4")
	if v := s.Value(0, 4); v.Num != 20 {
		t.Fatalf("diamond after edit = %+v", v)
	}
	// replacing a formula drops its old edges
	s.SetRaw(0, 4, "=7")
	s.SetRaw(0, 1, "100")
	if v := s.Value(0, 4); v.Num != 7 {
		t.Fatalf("stale edge recomputed cell: %+v", v)
	}
	// formulas depending on a cell that is cleared read 0
	s.SetRaw(5, 5, "=B1+1")
	s.SetRaw(0, 1, "")
	if v := s.Value(5, 5); v.Num != 1 {
		t.Fatalf("dependent of cleared cell = %+v", v)
	}
}

func TestEngineDepGraphCycles(t *testing.T) {
	w := NewWorkbook()
	s := w.Sheets[1]
	s.SetRaw(0, 0, "=A1")
	if v := s.Value(0, 0); v.Err != ErrCirc {
		t.Fatalf("self reference = %+v", v)
	}
	s.SetRaw(0, 0, "=B1")
	s.SetRaw(0, 1, "=C1")
	s.SetRaw(0, 2, "=A1")
	for c := 0; c < 3; c++ {
		if v := s.Value(0, c); v.Err != ErrCirc {
			t.Fatalf("cycle cell %d = %+v", c, v)
		}
	}
	// a downstream reader of the cycle is in error too, not hung
	s.SetRaw(3, 0, "=A1+1")
	if v := s.Value(3, 0); v.Kind != KindError {
		t.Fatalf("downstream of cycle = %+v", v)
	}
	// breaking the cycle heals everything
	s.SetRaw(0, 2, "5")
	if v := s.Value(0, 0); v.Num != 5 {
		t.Fatalf("A1 after break = %+v", v)
	}
	if v := s.Value(3, 0); v.Num != 6 {
		t.Fatalf("downstream after break = %+v", v)
	}
	// range containing the formula cell itself
	s.SetRaw(10, 0, "=SUM(A10:A12)")
	if v := s.Value(10, 0); v.Err != ErrCirc {
		t.Fatalf("range self-cycle = %+v", v)
	}
	// cross-sheet cycle
	a, b := w.Sheets[0], w.Sheets[2]
	a.SetRaw(30, 0, "=Summary!A1")
	b.SetRaw(0, 0, "=Sheet1!A31")
	if v := a.Value(30, 0); v.Err != ErrCirc {
		t.Fatalf("cross-sheet cycle = %+v", v)
	}
}

func TestEngineShiftFormula(t *testing.T) {
	cases := []struct {
		in     string
		dc, dr int
		want   string
	}{
		{"=A1+B2", 1, 1, "=B2+C3"},
		{"=$A$1+B2", 1, 1, "=$A$1+C3"},
		{"=$A1+B$2", 2, 3, "=$A4+D$2"},
		{"=SUM(A1:B5)", 0, 1, "=SUM(A2:B6)"},
		{"=SUM($A$1:B5)", 1, 1, "=SUM($A$1:C6)"},
		{"=Sheet2!A1+'My Sheet'!B2", 1, 0, "=Sheet2!B1+'My Sheet'!C2"},
		{"=\"A1\"&A1", 0, 1, "=\"A1\"&A2"},
		{"=A1", -1, 0, "=#REF!"},
		{"=SUM(LOG10(A1),2)", 1, 0, "=SUM(LOG10(B1),2)"},
		{"plain A1", 1, 1, "plain A1"},
	}
	for _, c := range cases {
		if got := ShiftFormula(c.in, c.dc, c.dr); got != c.want {
			t.Errorf("ShiftFormula(%q,%d,%d) = %q want %q", c.in, c.dc, c.dr, got, c.want)
		}
	}
	if CycleRefAbs("A1") != "$A$1" || CycleRefAbs("$A$1") != "A$1" || CycleRefAbs("A$1") != "$A1" || CycleRefAbs("$A1") != "A1" {
		t.Error("F4 cycle order wrong")
	}
}

func TestEngineFormulaRefs(t *testing.T) {
	got := FormulaRefs("=A1+SUM(B2:B4)*$C$5+Sheet2!D1")
	want := []string{"A1", "B2:B4", "C5"}
	if len(got) != len(want) {
		t.Fatalf("refs %v", got)
	}
	for i := range want {
		if got[i].String() != want[i] {
			t.Errorf("ref %d = %s want %s", i, got[i], want[i])
		}
	}
	toks := FormulaRefTokens("=A1+B2:B4")
	if len(toks) != 2 || toks[1].Start != 4 || toks[1].End != 9 {
		t.Errorf("tokens %+v", toks)
	}
	if len(FormulaRefs("=SUM(")) != 0 || len(FormulaRefs("x")) != 0 {
		t.Error("incomplete or non-formula should be empty")
	}
}

func TestEngineInsertDeleteRowsShiftFormulas(t *testing.T) {
	w := NewWorkbook()
	s := w.Sheets[1]
	for r := 0; r < 5; r++ {
		s.SetRaw(r, 0, fmt.Sprintf("%d", r+1))
	}
	s.SetRaw(6, 0, "=SUM(A1:A5)")
	s.SetRaw(6, 1, "=A5*2")
	s.SetRaw(6, 2, "=A2")
	w.Sheets[0].SetRaw(0, 10, "=Sheet2!A5")
	w.InsertRows(s, 4, 1) // above row 5
	if got := s.Raw(7, 1); got != "=A6*2" {
		t.Fatalf("shifted formula %q", got)
	}
	if got := s.Raw(7, 0); got != "=SUM(A1:A6)" {
		t.Fatalf("range expanded %q", got)
	}
	if got := w.Sheets[0].Raw(0, 10); got != "=Sheet2!A6" {
		t.Fatalf("cross-sheet shifted %q", got)
	}
	if v := s.Value(7, 0); v.Num != 15 {
		t.Fatalf("sum after insert %+v", v)
	}
	if !w.Undo() || s.Raw(6, 0) != "=SUM(A1:A5)" || s.Raw(7, 0) != "" {
		t.Fatal("undo of insert")
	}
	w.DeleteRows(s, 1, 1) // delete row 2: A2 is referenced by C7
	if v := s.Value(5, 2); v.Err != ErrRef {
		t.Fatalf("deleted target = %+v (raw %q)", v, s.Raw(5, 2))
	}
	if got := s.Raw(5, 0); got != "=SUM(A1:A4)" {
		t.Fatalf("range shrunk %q", got)
	}
	if v := s.Value(5, 0); v.Num != 13 {
		t.Fatalf("sum after delete %+v", v)
	}
	w.InsertCols(s, 0, 1)
	if got := s.Raw(5, 1); got != "=SUM(B1:B4)" {
		t.Fatalf("col insert %q", got)
	}
}
