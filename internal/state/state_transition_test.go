package state

import "testing"

func TestStateTransitionsWithoutRender(t *testing.T) {
	s := NewStore(map[string]any{"count": 0, "user": map[string]any{"name": "a"}})

	// SetData resets history.
	s.SetData(map[string]any{"count": 5, "user": map[string]any{"name": "a"}})
	if got := s.Data().(map[string]any)["count"]; got != 5 {
		t.Fatalf("setdata = %v", got)
	}
	if s.CanUndo() {
		t.Fatal("history not cleared")
	}

	// PatchChange and origin.
	s.SetOrigin(OriginUserInput)
	c := s.PatchChange("count", 7)
	if c.Origin != OriginUserInput || c.Old != 5 || c.New != 7 {
		t.Fatalf("change = %+v", c)
	}
	if got := s.ChangedProperties(); len(got) != 1 || got[0] != "count" {
		t.Fatalf("changed = %v", got)
	}
	if cp := c.Properties; len(cp) != 1 || cp[0] != "count" {
		t.Fatalf("change properties = %v", cp)
	}

	// Nested patch.
	nc := s.PatchChange("user.name", "b")
	if nc.Old != "a" || nc.New != "b" {
		t.Fatalf("nested = %+v", nc)
	}
	s.Undo()
	if got := s.Data().(map[string]any)["user"].(map[string]any)["name"]; got != "a" {
		t.Fatalf("undo = %v", got)
	}
	s.Redo()
	if got := s.Data().(map[string]any)["user"].(map[string]any)["name"]; got != "b" {
		t.Fatalf("redo = %v", got)
	}

	// PatchProjected.
	s.PatchProjected(map[string]any{"count": 9})
	if got := s.Data().(map[string]any)["count"]; got != 9 {
		t.Fatalf("projected = %v", got)
	}

	// Patch with Patch struct.
	s.Patch(Patch{Origin: OriginBackground, Changed: []string{"count"}, Values: map[string]any{"count": 11}})
	if s.Origin() != OriginBackground {
		t.Fatalf("origin = %s", s.Origin())
	}
	if got := s.Data().(map[string]any)["count"]; got != 11 {
		t.Fatalf("patch = %v", got)
	}

	// Whole patch.
	s.Patch(Patch{Values: map[string]any{"": map[string]any{"count": 42}}})
	if got := s.Data().(map[string]any)["count"]; got != 42 {
		t.Fatalf("whole = %v", got)
	}

}
