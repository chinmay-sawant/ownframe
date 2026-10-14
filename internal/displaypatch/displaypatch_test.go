package displaypatch

import (
	"image"
	"testing"
)

func TestDiff(t *testing.T) {
	old := []Op{{ID: 1, Hash: 1}, {ID: 2, Hash: 2}}
	new := []Op{{ID: 1, Hash: 9}, {ID: 3, Hash: 3}}
	res := Diff(old, new)
	if len(res.Changed) != 1 || len(res.Added) != 1 || len(res.Removed) != 1 {
		t.Fatalf("bad diff: %+v", res)
	}
}

func TestPool(t *testing.T) {
	p := NewPool()
	p.Release(5)
	if p.Acquire() != 5 || p.Size() != 0 {
		t.Fatal("pool broken")
	}
}

func TestTileDamage(t *testing.T) {
	td := TileDamage{}
	td.Merge(image.Rect(0, 0, 1, 1))
	td.Merge(image.Rectangle{})
	if len(td.Rects) != 1 {
		t.Fatal("tile damage broken")
	}
}

func TestRasterSignature(t *testing.T) {
	a := RasterSignature{Content: "x", AnimID: 1}
	b := a
	if !a.Equal(b) {
		t.Fatal("sig equal failed")
	}
}
