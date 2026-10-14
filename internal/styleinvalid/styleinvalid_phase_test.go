package styleinvalid

import "testing"

func TestSkipUnmodifiedSelectors(t *testing.T) {
	sels := []string{"a", "b", "c"}
	got := SkipUnmodifiedSelectors(sels, []string{"b"})
	if len(got) != 1 || got[0] != "b" {
		t.Fatalf("got %+v", got)
	}
	tr := New(3)
	tr.Touch(Change{BoxID: "a"})
	got2 := SkipUnmodifiedSelectors(sels, tr.Changes())
	if len(got2) != 1 || got2[0] != "a" {
		t.Fatalf("got2 %+v", got2)
	}
}

func TestStylePaths(t *testing.T) {
	changes := []Change{{BoxID: "a", Property: "color"}, {BoxID: "a", Property: "color"}}
	paths := StylePathsFromChanges(changes)
	if len(paths) != 1 || paths[0] != "a/color" {
		t.Fatalf("paths %+v", paths)
	}
	tr := New(3)
	tr.Touch(Change{BoxID: "b", Property: "width"})
	if !tr.StyleDirty("b/width") {
		t.Fatal("dirty path missing")
	}
	if len(tr.StylePaths()) != 1 {
		t.Fatal("tracker paths wrong")
	}
	built := InvalidateStylePaths([]StylePath{"x/y"})
	if len(built) != 1 || built[0].StylePath() != "x/y" {
		t.Fatalf("built %+v", built)
	}
}

func TestTextRunStability(t *testing.T) {
	k := TextRunKey{Face: "f", Content: "hi"}
	if !k.Stable() {
		t.Fatal("key not stable")
	}
	c := NewRunCache()
	if !c.Touch(k) {
		t.Fatal("first touch")
	}
	if c.Touch(k) {
		t.Fatal("second touch new?")
	}
	if !c.Seen(k) || !c.Stable(k) {
		t.Fatal("stability broken")
	}
	if c.StableRuns() != 1 || c.Touches() != 2 {
		t.Fatalf("counts %d %d", c.StableRuns(), c.Touches())
	}
}
