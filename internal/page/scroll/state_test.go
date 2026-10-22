package scroll

import "testing"

func TestStateToByTake(t *testing.T) {
	t.Parallel()

	var s State
	s.To(10, 20)
	req, ok := s.Take()
	if !ok || req.X != 10 || req.Y != 20 || !req.Absolute {
		t.Fatalf("to %+v %v", req, ok)
	}
	if _, ok := s.Take(); ok {
		t.Fatal("cleared")
	}

	s.By(1, 2)
	s.By(3, 4)
	req, ok = s.Take()
	if !ok || req.X != 4 || req.Y != 6 || req.Absolute {
		t.Fatalf("by %+v %v", req, ok)
	}

	s.To(5, 6)
	s.By(1, 1)
	req, ok = s.Take()
	if !ok || req.X != 6 || req.Y != 7 || !req.Absolute {
		t.Fatalf("replace %+v %v", req, ok)
	}
}

func TestWindowingApplyAndStep(t *testing.T) {
	t.Parallel()

	var s State
	if !s.Step(100) {
		t.Fatal("nil fn")
	}
	s.Apply(100)

	var gotY, gotH int
	s.SetFn(func(offsetY, viewH int) bool {
		gotY, gotH = offsetY, viewH
		return false
	})
	s.SetOffset(3, 9)
	if s.Step(40) {
		t.Fatal("step")
	}
	if gotY != 9 || gotH != 40 {
		t.Fatalf("step %d %d", gotY, gotH)
	}
	s.SetOn(true)
	if !s.On() {
		t.Fatal("on")
	}
	s.Apply(50)
	if gotY != 9 || gotH != 50 {
		t.Fatalf("apply %d %d", gotY, gotH)
	}
	x, y := s.Offset()
	if x != 3 || y != 9 {
		t.Fatalf("offset %d %d", x, y)
	}
}
