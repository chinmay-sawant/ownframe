package state

import "testing"

func TestTrackerRecordAndLast(t *testing.T) {
	tr := NewTracker()
	p := Patch{Origin: OriginUserInput, Changed: []string{"count"}, Values: map[string]any{"count": 1}}
	tr.Record(p)
	got, ok := tr.Last()
	if !ok || got.Origin != OriginUserInput || !got.Has("count") {
		t.Fatalf("last patch = %+v, %v", got, ok)
	}
	if got.IsWhole() {
		t.Fatal("scalar patch reported whole")
	}
}

func TestPatchWhole(t *testing.T) {
	p := Patch{}
	if !p.IsWhole() {
		t.Fatal("empty patch not whole")
	}
}

func TestStableIDsFallback(t *testing.T) {
	got := StableIDs(3, nil)
	if len(got) != 3 || got[0] != "0" || got[2] != "2" {
		t.Fatalf("got %v", got)
	}
	got = StableIDs(2, func(i int) string { return "k" })
	if got[0] != "k" || got[1] != "k" {
		t.Fatalf("got %v", got)
	}
}

func TestOriginString(t *testing.T) {
	if OriginBackground.String() != "background" {
		t.Fatal("bad string")
	}
}
