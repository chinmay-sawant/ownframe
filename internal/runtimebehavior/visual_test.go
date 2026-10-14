package runtimebehavior

import "testing"

func TestVisualCompareHook(t *testing.T) {
	called := false
	v := NewVisualCompare()
	v.SetMismatchHook(func() { called = true })
	v.SetImages([]byte("a"), []byte("b"))
	if v.Compare() {
		t.Fatal("expected mismatch")
	}
	if !called {
		t.Fatal("hook not called")
	}
	if v.IsEqual() {
		t.Fatal("equal flag wrong")
	}
	if !VisualEqual([]byte("x"), []byte("x")) {
		t.Fatal("VisualEqual broken")
	}
}
