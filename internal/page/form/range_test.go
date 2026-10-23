package form

import "testing"

func TestRangeAndClamp(t *testing.T) {
	t.Parallel()

	start, end := RangeOf(5, 1, 4)
	if start != 1 || end != 4 {
		t.Fatalf("range [%d,%d)", start, end)
	}
	if ClampPos(-1, 3) != 0 || ClampPos(9, 3) != 3 || ClampPos(2, 3) != 2 {
		t.Fatal("clamp")
	}
	if RuneLen("hél") != 3 {
		t.Fatal("runes")
	}
	if RuneAtByte("héllo", 0) != 0 || RuneAtByte("héllo", 2) != 1 {
		t.Fatal("byte")
	}
	if ClampField("abcdef", 3) != "abc" || ClampField("ab", 3) != "ab" {
		t.Fatal("field")
	}
	if FieldCap(Control{}) != DefaultMaxFieldLen {
		t.Fatal("default cap")
	}
	if FieldCap(Control{MaxLength: 8}) != 8 {
		t.Fatal("author cap")
	}
}
