package state

import "testing"

func TestStorePatchChangeRecordsChange(t *testing.T) {
	s := NewStore(map[string]any{"count": 0})
	c := s.PatchChange("count", 1)
	if c.Path != "count" || c.Old != 0 || c.New != 1 {
		t.Fatalf("change = %+v", c)
	}
	if got := s.Data().(map[string]any)["count"]; got != 1 {
		t.Fatalf("data count = %v", got)
	}
	if len(s.Changes()) != 1 || !s.CanUndo() {
		t.Fatal("change not recorded")
	}
}

func TestStoreUndoRedo(t *testing.T) {
	s := NewStore(map[string]any{"count": 0})
	s.PatchChange("count", 1)
	s.Undo()
	if got := s.Data().(map[string]any)["count"]; got != 0 {
		t.Fatalf("after undo count = %v", got)
	}
	if !s.CanRedo() {
		t.Fatal("redo unavailable")
	}
	s.Redo()
	if got := s.Data().(map[string]any)["count"]; got != 1 {
		t.Fatalf("after redo count = %v", got)
	}
}

func TestStorePatchChangeNested(t *testing.T) {
	s := NewStore(map[string]any{"user": map[string]any{"name": "a"}})
	c := s.PatchChange("user.name", "b")
	if c.Old != "a" || c.New != "b" {
		t.Fatalf("change = %+v", c)
	}
	if got := s.Data().(map[string]any)["user"].(map[string]any)["name"]; got != "b" {
		t.Fatalf("nested data = %v", got)
	}
	s.Undo()
	if got := s.Data().(map[string]any)["user"].(map[string]any)["name"]; got != "a" {
		t.Fatalf("undo nested = %v", got)
	}
}

func TestStoreUndoRedoEmpty(t *testing.T) {
	s := NewStore(map[string]any{"count": 0})
	s.Undo()
	s.Redo()
	if got := s.Data().(map[string]any)["count"]; got != 0 {
		t.Fatalf("count = %v", got)
	}
}
