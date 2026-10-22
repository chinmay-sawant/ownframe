package form

import "testing"

func TestMoveCaret(t *testing.T) {
	t.Parallel()

	pos, extend, ok := MoveCaret("arrowleft", "abcd ef", 2, 2, 2)
	if !ok || extend || pos != 1 {
		t.Fatalf("left %d %v %v", pos, extend, ok)
	}
	pos, _, ok = MoveCaret("end", "abcd ef", 2, 2, 2)
	if !ok || pos != 7 {
		t.Fatalf("end %d", pos)
	}
	pos, _, ok = MoveCaret("home", "abcd ef", 2, 2, 2)
	if !ok || pos != 0 {
		t.Fatalf("home %d", pos)
	}
	pos, _, ok = MoveCaret("ctrl+arrowright", "abcd ef", 2, 2, 2)
	if !ok || pos != 5 {
		t.Fatalf("word %d", pos)
	}
	pos, extend, ok = MoveCaret("shift+arrowright", "abcd ef", 5, 5, 5)
	if !ok || !extend || pos != 6 {
		t.Fatalf("shift %d %v", pos, extend)
	}
	_, _, ok = MoveCaret("enter", "ab", 0, 0, 0)
	if ok {
		t.Fatal("enter")
	}
}
