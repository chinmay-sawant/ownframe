package viewtransition

import "testing"

func TestViewTransitionBaseRemovesCapturedOperationIDs(t *testing.T) {
	old, next := pairedDisplays(t)
	tr, err := New(old, next, Scope{DocumentRoot, DocumentRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	f := tr.At(0.5)
	kept := make(map[uint64]bool)
	for _, op := range f.Base.Ops {
		kept[op.ID] = true
	}
	for _, box := range f.Base.Boxes {
		for _, id := range box.OperationIDs {
			if !kept[id] {
				t.Fatalf("base box retains captured operation id %d", id)
			}
		}
	}
}
