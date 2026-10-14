package state

import "testing"

func TestStateTransitionsIdentity(t *testing.T) {
	s := NewStore(map[string]any{"count": 0, "user": map[string]any{"name": "a"}})
	s.SetData(map[string]any{"count": 5, "user": map[string]any{"name": "a"}})

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
}
