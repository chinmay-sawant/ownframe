package state

import "testing"

func TestPathRoundTripEmptySegments(t *testing.T) {
	bad := []string{"a.", "a..b", ".a"}
	for _, path := range bad {
		data := map[string]any{"a": map[string]any{"b": 1}}
		if _, ok := setPath(data, path, 2); ok {
			t.Fatalf("%q set ok=true", path)
		}
		if got := valueAtPath(data, path); got != nil {
			t.Fatalf("%q read=%v want nil", path, got)
		}
		if _, has := data[""]; has {
			t.Fatalf("%q created empty key", path)
		}
		if _, ok := splitPath(path); ok {
			t.Fatalf("%q split ok=true", path)
		}
	}
	nd, ok := setPath(map[string]any{"a": 1}, "", map[string]any{"b": 2})
	if !ok {
		t.Fatal("empty set ok=false")
	}
	if m, good := nd.(map[string]any); !good || m["b"] != 2 {
		t.Fatalf("empty set=%v", nd)
	}
	if got := valueAtPath(map[string]any{"x": 1}, ""); got == nil {
		t.Fatal("empty read=nil")
	}
}
