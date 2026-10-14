package runtimebehavior

import (
	"bytes"
	"testing"
)

func TestVisualCompareHooksEdge(t *testing.T) {
	hookCalls := 0
	v := NewVisualCompare()
	v.SetMismatchHook(func() { hookCalls++ })
	v.Hook = func(full, partial []byte) bool {
		return bytes.Equal(full, partial) || len(full) == len(partial)
	}
	v.SetImages([]byte("ab"), []byte("cd"))
	if !v.Compare() {
		t.Fatal("hook should force equal")
	}
	if hookCalls != 0 {
		t.Fatal("hook called on match")
	}
	v.SetImages([]byte("a"), []byte("bb"))
	if v.Compare() {
		t.Fatal("hook should force mismatch")
	}
	if hookCalls != 1 {
		t.Fatal("mismatch hook not called once")
	}

	v3 := NewVisualComparer()
	called := false
	v3.CompareHook = func(full, partial []byte) bool {
		called = true
		return true
	}
	v3.SetImages([]byte("x"), []byte("y"))
	if !v3.Compare() || !called {
		t.Fatal("compare hook broken")
	}
}
