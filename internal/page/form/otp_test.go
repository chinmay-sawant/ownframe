package form

import "testing"

func TestScanReadsNumericSlot(t *testing.T) {
	t.Parallel()

	spans := Scan(`<input id="code-0" type="text" inputmode="numeric" maxlength="1" value="7">`)
	if len(spans) != 1 {
		t.Fatalf("spans %d", len(spans))
	}

	got := spans[0].Control
	if got.InputMode != "numeric" || got.MaxLength != 1 || got.Value != "7" {
		t.Fatalf("%+v", got)
	}
}

func TestInsertSlotRejectsNonDigitAndAdvances(t *testing.T) {
	t.Parallel()

	slot := Control{ID: "code-0", Tag: "input", Type: "text", MaxLength: 1, InputMode: "numeric"}
	next, caret, _, ok := InsertSlot(slot, 0, 0, "a")
	if ok || next.Value != "" || next.SlotStep != 0 {
		t.Fatalf("non-digit %+v %d %v", next, caret, ok)
	}

	next, caret, _, ok = InsertSlot(slot, 0, 0, "4")
	if !ok || next.Value != "4" || caret != 1 || next.SlotStep != slotNext {
		t.Fatalf("digit %+v %d %v", next, caret, ok)
	}

	plain := Control{ID: "name", Tag: "input", Type: "text", Value: "a"}
	next, _, _, ok = InsertSlot(plain, 1, 1, "b")
	if !ok || next.Value != "ab" || next.SlotStep != 0 {
		t.Fatalf("plain %+v %v", next, ok)
	}
}

func TestBackspaceSlotMovesWhenEmpty(t *testing.T) {
	t.Parallel()

	byID := map[string]Control{
		"code-0": {ID: "code-0", Tag: "input", Type: "text", MaxLength: 1, InputMode: "numeric", Value: "4"},
		"code-1": {ID: "code-1", Tag: "input", Type: "text", MaxLength: 1, InputMode: "numeric"},
	}
	empty := byID["code-1"]
	next, _, _, ok := BackspaceSlot(empty, 0, 0)
	if ok || next.SlotStep != slotPrev {
		t.Fatalf("empty %+v %v", next, ok)
	}
	if got := SlotNeighbor(byID, "code-1", next.SlotStep); got != "code-0" {
		t.Fatalf("previous %q", got)
	}

	filled := byID["code-0"]
	next, caret, _, ok := BackspaceSlot(filled, 1, 1)
	if !ok || next.Value != "" || caret != 0 || next.SlotStep != 0 {
		t.Fatalf("delete %+v %d %v", next, caret, ok)
	}
	if got := SlotNeighbor(byID, "code-1", slotNext); got != "" {
		t.Fatalf("past end %q", got)
	}
}
