package snip

import "testing"

func TestNormalize(t *testing.T) {
	want := Rect{X: 1, Y: 1, W: 2, H: 3}
	if got := Normalize(Point{3, 1}, Point{1, 4}); got != want {
		t.Fatalf("Normalize reversed = %+v, want %+v", got, want)
	}
	if got := Normalize(Point{1, 1}, Point{3, 4}); got != want {
		t.Fatalf("Normalize ordered = %+v, want %+v", got, want)
	}
	if got := Normalize(Point{2, 2}, Point{2, 2}); got != (Rect{X: 2, Y: 2}) {
		t.Fatalf("Normalize same = %+v, want empty at origin", got)
	}
}

func TestClamp(t *testing.T) {
	b := Rect{W: 10, H: 10}
	if got := Clamp(Rect{2, 2, 3, 3}, b); got != (Rect{2, 2, 3, 3}) {
		t.Fatalf("Clamp inside = %+v", got)
	}
	if got := Clamp(Rect{-2, -2, 5, 5}, b); got != (Rect{0, 0, 3, 3}) {
		t.Fatalf("Clamp partial = %+v", got)
	}
	if got := Clamp(Rect{8, 8, 5, 5}, b); got != (Rect{8, 8, 2, 2}) {
		t.Fatalf("Clamp overflow = %+v", got)
	}
	if got := Clamp(Rect{20, 20, 5, 5}, b); got.W != 0 || got.H != 0 {
		t.Fatalf("Clamp outside = %+v, want zero size", got)
	}
}

func TestSelectionRect(t *testing.T) {
	s := Selection{Start: Point{5, 5}, End: Point{2, 8}, Active: true}
	if got, want := s.Rect(), (Rect{2, 5, 3, 3}); got != want {
		t.Fatalf("Selection.Rect = %+v, want %+v", got, want)
	}
	if (Selection{}).Rect() != (Rect{}) {
		t.Fatal("zero Selection must give zero Rect")
	}
}
