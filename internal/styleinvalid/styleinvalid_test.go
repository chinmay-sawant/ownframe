package styleinvalid

import "testing"

func TestTrackerTouchAndFull(t *testing.T) {
	tr := New(1)
	tr.Touch(Change{BoxID: "a", Property: "color"})
	if tr.Full() {
		t.Fatal("one change overflowed bound 1?")
	}
	tr.Touch(Change{BoxID: "b", Property: "color"})
	tr.Touch(Change{BoxID: "b", Property: "color"})
	if !tr.Full() {
		t.Fatal("expected full after 2 changes")
	}
	if len(tr.Changes()) != 2 {
		t.Fatalf("changes = %d", len(tr.Changes()))
	}
	tr.Reset()
	if len(tr.Changes()) != 0 {
		t.Fatal("reset failed")
	}
}

func TestRanges(t *testing.T) {
	idx := map[string]int{"a": 0, "b": 2}
	changes := []Change{{BoxID: "a"}, {BoxID: "b"}, {BoxID: "missing"}}
	got := RangesFromChanges(changes, idx, 4)
	if len(got) != 2 || got[0].Start != 0 || got[1].Start != 2 {
		t.Fatalf("got %+v", got)
	}
}

func TestPromoteAndRunCache(t *testing.T) {
	if !PromoteLayer("fixed") || PromoteLayer("static") {
		t.Fatal("promotion wrong")
	}
	c := NewRunCache()
	k := TextRunKey{Face: "f", Content: "hi"}
	if !c.Touch(k) || c.Touch(k) {
		t.Fatal("run cache broken")
	}
	if c.Count() != 1 {
		t.Fatal("count wrong")
	}
}
