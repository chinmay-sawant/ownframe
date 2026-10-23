package form

import "testing"

func TestInsertBackspaceDelete(t *testing.T) {
	t.Parallel()

	c := Control{Value: "abcd", MaxLength: 8}
	next, pos, _, ok := InsertValue(c, 2, 2, "X")
	if !ok || next.Value != "abXcd" || pos != 3 {
		t.Fatalf("insert %q %d %v", next.Value, pos, ok)
	}
	next, pos, _, ok = BackspaceValue(next, 3, 3)
	if !ok || next.Value != "abcd" || pos != 2 {
		t.Fatalf("backspace %q %d %v", next.Value, pos, ok)
	}
	next, pos, _, ok = DeleteWordValue(next, 2, 2)
	if !ok || next.Value != "cd" || pos != 0 {
		t.Fatalf("delete %q %d %v", next.Value, pos, ok)
	}
	c = Control{MaxLength: 2}
	next, pos, _, ok = InsertValue(c, 0, 0, "abcdef")
	if !ok || next.Value != "ab" || pos != 2 {
		t.Fatalf("clamp insert %q %d %v", next.Value, pos, ok)
	}
}
