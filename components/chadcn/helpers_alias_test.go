package chadcn

import (
	"slices"
	"testing"
)

// TestClassesDoesNotMutateVariadicBacking pins HYG-CLEAN-01: classes(slice...)
// must leave the caller's slice untouched.
func TestClassesDoesNotMutateVariadicBacking(t *testing.T) {
	t.Parallel()
	in := []string{"", "a", "", "b"}
	want := slices.Clone(in)
	if got := classes(in...); got != "a b" {
		t.Fatalf("classes = %q, want %q", got, "a b")
	}
	if !slices.Equal(in, want) {
		t.Fatalf("caller slice mutated: %q, want %q", in, want)
	}
}
