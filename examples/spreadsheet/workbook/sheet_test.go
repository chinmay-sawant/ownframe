package workbook

import (
	"errors"
	"testing"
)

func TestSheetEachAndUsedRange(t *testing.T) {
	w := New(1, "test")
	s := w.AddSheet("S")

	mustApply(t, w, s, Pos{2, 3}, "c")
	mustApply(t, w, s, Pos{0, 1}, "a")
	mustApply(t, w, s, Pos{1, 0}, "b")

	var order []Pos

	s.Each(func(p Pos, _ Cell) bool {
		order = append(order, p)

		return true
	})

	want := []Pos{{0, 1}, {1, 0}, {2, 3}}
	if len(order) != len(want) {
		t.Fatalf("Each visited %v", order)
	}

	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("Each order = %v, want %v", order, want)
		}
	}

	stopped := 0

	s.Each(func(Pos, Cell) bool {
		stopped++

		return false
	})

	if stopped != 1 {
		t.Fatalf("Each visited %d cells after false", stopped)
	}

	r, ok := s.UsedRange()
	if !ok || r.Min != (Pos{0, 0}) || r.Max != (Pos{2, 3}) {
		t.Fatalf("UsedRange = %+v", r)
	}

	if s.Count() != 3 {
		t.Fatalf("Count = %d", s.Count())
	}
}

func TestSheetAccessors(t *testing.T) {
	w := New(1, "test")

	if w.SheetByName("missing") != nil || w.Sheet(7) != nil {
		t.Fatal("lookups found a missing sheet")
	}

	a := w.AddSheet("A")
	b := w.AddSheet("B")

	if w.AddSheet("A") != a {
		t.Fatal("AddSheet duplicated a name")
	}

	if len(w.Sheets()) != 2 {
		t.Fatalf("sheets = %d", len(w.Sheets()))
	}

	if err := w.RemoveSheet(a.ID()); err != nil {
		t.Fatal(err)
	}

	if w.Sheet(a.ID()) != nil || len(w.Sheets()) != 1 {
		t.Fatal("remove did not drop the sheet")
	}

	if err := w.RemoveSheet(b.ID()); !errors.Is(err, ErrNoSheet) {
		t.Fatalf("removing the last sheet = %v", err)
	}
}
