package ui

import (
	"sort"
	"strings"
	"testing"
)

// TestGridFrameElementBudget guards per-frame layout cost: a full grid view must
// stay near 500 elements, since every frame relayouts the whole page.
func TestGridFrameElementBudget(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.must(r.a.LoadSyntheticRows(r.ctx, 100000))
	n := map[string]int{}
	bs := r.a.Page().Boxes()
	for _, b := range bs {
		id := b.ID
		if i := strings.IndexAny(id, "0123456789"); i > 0 {
			id = id[:i]
		}
		if id == "" {
			id = "(" + b.Tag + ")"
		}
		n[id]++
	}
	type kv struct {
		k string
		v int
	}
	var l []kv
	for k, v := range n {
		l = append(l, kv{k, v})
	}
	sort.Slice(l, func(i, j int) bool { return l[i].v > l[j].v })
	if len(bs) > 600 {
		t.Fatalf("a frame has %d elements, budget 600", len(bs))
	}
	t.Logf("total %d", len(bs))
	for _, e := range l[:12] {
		t.Logf("%s %d", e.k, e.v)
	}
}
