package element

import "testing"

func TestRegistryRetainsInstance(t *testing.T) {
	r := NewRegistry()
	a := r.GetOrCreate("btn", "button")
	b := r.GetOrCreate("btn", "button")
	if a != b {
		t.Fatal("instance not retained")
	}
	if r.Len() != 1 {
		t.Fatalf("len=%d", r.Len())
	}
}

func TestDispositionBeforePaint(t *testing.T) {
	r := NewRegistry()
	n := r.GetOrCreate("row", "div")
	q := NewDisposalQueue()
	alive := map[string]bool{}
	removed := DisposeStale(n, alive, q)
	if len(removed) != 1 {
		t.Fatalf("removed=%d", len(removed))
	}
	if q.Pending() != 1 {
		t.Fatalf("pending=%d", q.Pending())
	}
	q.Flush()
	if !n.Disposed {
		t.Fatal("not disposed")
	}
}
