package viewtransition

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestViewTransitionPairsBlinklessSnapshots(t *testing.T) {
	old, next := pairedDisplays(t)
	if len(old.OpOwners) != len(old.Ops) || len(next.OpOwners) != len(next.Ops) {
		t.Fatal("snapshots lack operation owners")
	}
	if next.Boxes[boxNamed(next, "c")].ParentIndex != boxNamed(next, "p") {
		t.Fatal("child parent index is wrong")
	}
	tr, err := New(old, next, Scope{DocumentRoot, DocumentRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	g := tr.Groups()
	if len(g) != 2 || g[0].Name != "gallery" || g[1].Name != "item" || g[1].Parent != "gallery" {
		t.Fatalf("groups=%v", g)
	}
	if g[1].Mode != "nearest" {
		t.Fatalf("item mode=%q", g[1].Mode)
	}
	for _, op := range tr.groups[0].newOps {
		if next.OpOwners[op] == boxNamed(next, "c") {
			t.Fatalf("gallery capture duplicates child op %d", op)
		}
	}
	f := tr.At(0.5)
	if f.Groups[0].Bounds.W != 100 || len(f.Base.Ops) >= len(next.Ops) {
		t.Fatalf("frame bounds/base = %+v/%d", f.Groups[0].Bounds, len(f.Base.Ops))
	}
}

func pairedDisplays(t *testing.T) (*layout.Display, *layout.Display) {
	t.Helper()
	old := `<style>.p{display:block;width:80px;height:60px;background:red;view-transition-name:gallery;view-transition-group:contain}.c{display:block;width:30px;height:20px;background:blue;view-transition-name:item;view-transition-group:nearest;view-transition-class:old}</style><main id="p" class="p"><div id="c" class="c">old</div></main><p id="base">base</p>`
	next := `<style>.p{display:block;width:120px;height:80px;background:green;view-transition-name:gallery;view-transition-group:contain}.c{display:block;width:40px;height:24px;background:yellow;view-transition-name:item;view-transition-group:nearest;view-transition-class:card promo}</style><main id="p" class="p"><div id="c" class="c">new</div></main><p id="base">base</p>`
	return snapshot(t, old), snapshot(t, next)
}
