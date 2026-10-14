package state

import "testing"

func TestStateTransitionsTrackerAndListeners(t *testing.T) {
	s := NewStore(map[string]any{"count": 0, "user": map[string]any{"name": "a"}})
	s.SetData(map[string]any{"count": 5, "user": map[string]any{"name": "a"}})

	// Tracker.
	tr := NewTracker()
	tr.Record(Patch{Origin: OriginBackground, Changed: []string{"count"}, Values: map[string]any{"count": 1}})
	gotPatch, ok := tr.Last()
	if !ok || gotPatch.IsWhole() || !gotPatch.Has("count") {
		t.Fatalf("tracker = %+v, %v", gotPatch, ok)
	}
	if len(tr.All()) != 1 {
		t.Fatal("tracker all")
	}
	tr.Reset()
	if len(tr.All()) != 0 {
		t.Fatal("tracker reset")
	}

	// Listeners.
	gotListener := ""
	s.AddListener(func(c Change) {
		gotListener = c.Path
	})
	s.PatchChange("count", 99)
	if gotListener != "count" {
		t.Fatalf("listener = %q", gotListener)
	}
	if !s.RemoveListener(s.Listeners()[len(s.Listeners())-1]) {
		t.Fatal("remove listener failed")
	}

	// Undo/redo empty safety.
	empty := NewStore(nil)
	empty.Undo()
	empty.Redo()
	if empty.Data() != nil {
		t.Fatal("empty data changed")
	}

	// Stable.
	if ids := StableIDs(2, nil); ids[0] != "0" || ids[1] != "1" {
		t.Fatalf("ids = %v", ids)
	}
}
