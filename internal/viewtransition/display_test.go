package viewtransition

import "testing"

func TestViewTransitionFrameBuildsRenderableDisplay(t *testing.T) {
	old, next := pairedDisplays(t)
	tr, err := New(old, next, Scope{DocumentRoot, DocumentRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	f := tr.At(.5)
	d := f.Display()
	want := len(f.Base.Ops)
	for _, g := range f.Groups {
		want += len(g.Old) + len(g.New)
	}
	if len(d.Ops) != want || len(d.Order) != want || len(d.OpOwners) != want {
		t.Fatalf("composited display sizes ops/order/owners=%d/%d/%d, want %d", len(d.Ops), len(d.Order), len(d.OpOwners), want)
	}
	seen := make(map[uint64]bool, len(d.Ops))
	for i, op := range d.Ops {
		if op.ID == 0 || seen[op.ID] {
			t.Fatalf("operation %d has non-unique id %d", i, op.ID)
		}
		seen[op.ID] = true
		owner := d.OpOwners[i]
		if owner >= 0 && owner < len(d.Boxes) && !hasID(d.Boxes[owner].OperationIDs, op.ID) {
			t.Fatalf("owner box %d does not list operation %d", owner, op.ID)
		}
	}
}
