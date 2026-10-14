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

	// Identity binding.
	s.PatchChangeFor("item-1", "count", 7)
	change, ok := s.LastChange()
	if !ok || change.Identity != "item-1" || change.ElementID != "item-1" || change.StableID != "item-1" {
		t.Fatalf("identity change = %+v, %v", change, ok)
	}
	if !s.CanUndoFor("item-1") {
		t.Fatal("can undo identity missing")
	}
	if !s.UndoFor("item-1") {
		t.Fatal("undo identity failed")
	}
	if !s.CanRedoFor("item-1") {
		t.Fatal("can redo identity missing")
	}
	if !s.RedoFor("item-1") {
		t.Fatal("redo identity failed")
	}

	// Focus/scroll preservation.
	s.SetFocusState(FocusState{ID: "f", Caret: 3, SelectionStart: 1, SelectionEnd: 2})
	s.SetFocusState(FocusState{ID: "f"})
	if fs := s.FocusStateValue(); fs.Caret != 3 || fs.SelectionStart != 1 || fs.SelectionEnd != 2 {
		t.Fatalf("focus not preserved: %+v", fs)
	}
	s.SetScrollState(ScrollState{X: 10, Y: 20})
	s.SetScrollState(ScrollState{})
	if sc := s.ScrollStateValue(); sc.X != 10 || sc.Y != 20 {
		t.Fatalf("scroll not preserved: %+v", sc)
	}
	if u := s.UIStateValue(); u.Focus.ID != "f" || u.Scroll.X != 10 {
		t.Fatalf("ui state = %+v", u)
	}

	// Interceptors.
	called := false
	s.AddRouteInterceptor(func(path string) bool {
		called = path == "/x"
		return called
	})
	if !s.InterceptRoute("/x") || !called {
		t.Fatal("interceptor missed")
	}
	if s.InterceptRoute("/y") {
		t.Fatal("interceptor false positive")
	}
	s.RemoveRouteInterceptor(s.RouteInterceptors()[0])
	if len(s.RouteInterceptors()) != 0 {
		t.Fatal("interceptor not removed")
	}

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
