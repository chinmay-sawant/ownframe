package form

import "testing"

func TestWordAndLineBounds(t *testing.T) {
	t.Parallel()

	start, end := WordBounds("hello world", 1)
	if start != 0 || end != 5 {
		t.Fatalf("word [%d,%d)", start, end)
	}
	start, end = WordBounds("hello world", 10)
	if start != 6 || end != 11 {
		t.Fatalf("second [%d,%d)", start, end)
	}
	start, end = LineBounds("one\ntwo", 5)
	if start != 4 || end != 7 {
		t.Fatalf("line [%d,%d)", start, end)
	}
	if PrevWordStart("abcd ef", 2) != 0 {
		t.Fatal("prev")
	}
	if NextWordStart("abcd ef", 2) != 5 {
		t.Fatal("next")
	}
	if !IsWordRune('a') || IsWordRune(' ') {
		t.Fatal("rune")
	}
}
