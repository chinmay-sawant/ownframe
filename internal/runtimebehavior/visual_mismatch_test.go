package runtimebehavior

import (
	"testing"
)

func TestVisualCompareMismatchAlias(t *testing.T) {
	mismatchCalls := 0
	v := NewVisualComparer()
	v.SetOnMismatch(func() { mismatchCalls++ })
	v.SetImages([]byte("a"), []byte("b"))
	if v.Compare() {
		t.Fatal("expected mismatch")
	}
	if mismatchCalls != 1 {
		t.Fatal("on mismatch hook count wrong")
	}
	v.SetImages([]byte("a"), []byte("a"))
	if !v.Compare() {
		t.Fatal("expected match")
	}
	if mismatchCalls != 1 {
		t.Fatal("mismatch hook called on match")
	}
}
