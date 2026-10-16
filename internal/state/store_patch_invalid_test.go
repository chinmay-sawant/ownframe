package state

import "testing"

func TestTryPatchChangeInvalidSkips(t *testing.T) {
	s := NewStore(map[string]any{"count": 0})
	fired := 0
	s.AddListener(func(c Change) { fired++ })
	c, ok := s.TryPatchChange("count.x", 1)
	if ok {
		t.Fatalf("ok=true for %+v", c)
	}
	if fired != 0 {
		t.Fatalf("events=%d want 0", fired)
	}
	if len(s.Changes()) != 0 {
		t.Fatalf("changes=%d want 0", len(s.Changes()))
	}
	if s.CanUndo() {
		t.Fatal("undo recorded on failure")
	}
	if got := s.Data().(map[string]any)["count"]; got != 0 {
		t.Fatalf("count=%v want 0", got)
	}
	s2 := NewStore(map[string]any{"count": 0})
	n := 0
	s2.AddListener(func(c Change) { n++ })
	s2.PatchChange("count.x", 1)
	if n != 0 || len(s2.Changes()) != 0 || s2.CanUndo() {
		t.Fatalf("shim side effects n=%d changes=%d", n, len(s2.Changes()))
	}
}
