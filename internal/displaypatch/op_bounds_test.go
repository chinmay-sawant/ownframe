package displaypatch

import (
	"image"
	"testing"
)

func TestOpDominatesAndBounds(t *testing.T) {
	outer := Op{Bounds: image.Rect(0, 0, 10, 10)}
	inner := Op{Bounds: image.Rect(2, 2, 8, 8)}
	if !outer.Dominates(inner) {
		t.Fatal("domination failed")
	}
	if got := inner.PrecomputeBounds(3, 4); got != image.Rect(5, 6, 11, 12) {
		t.Fatalf("bounds %v", got)
	}
}

func TestDiffCountsOps(t *testing.T) {
	old := []Op{{ID: 1, Hash: 1, Bounds: image.Rect(0, 0, 1, 1)}, {ID: 2, Hash: 2, Bounds: image.Rect(1, 1, 2, 2)}}
	next := []Op{{ID: 1, Hash: 9, Bounds: image.Rect(0, 0, 1, 1)}, {ID: 3, Hash: 3, Bounds: image.Rect(2, 2, 3, 3)}}
	res := Diff(old, next)
	if len(res.Changed) != 1 || len(res.Added) != 1 || len(res.Removed) != 1 || res.Unchanged != 0 {
		t.Fatalf("bad diff: %+v", res)
	}
}
