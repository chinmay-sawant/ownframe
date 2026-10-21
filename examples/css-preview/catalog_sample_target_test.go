package main

import "testing"

// TestCatalogSamplesBindToTheirOwnTarget guards against reusing display
// indices from before the click redraw. A stale index points at the Start
// button's paint, so the sample moves the button instead of the preview.
func TestCatalogSamplesBindToTheirOwnTarget(t *testing.T) {
	p, d, v := fixturePage(t)
	for _, row := range v.Catalog {
		startCatalog(t, p, d, row.Name)
		s := d.samples[row.SampleKey]
		if s == nil || len(s.plan.pairs) == 0 {
			t.Fatalf("%s did not produce a motion plan", row.Name)
		}
		want := map[uint64]bool{}
		for _, box := range p.Boxes() {
			if box.ID == row.TargetID {
				for _, id := range box.OperationIDs {
					want[id] = true
				}
			}
		}
		for _, pair := range s.plan.pairs {
			if !want[pair.op.ID] {
				t.Fatalf("%s animated operation %d outside %s", row.Name, pair.op.ID, row.TargetID)
			}
		}
	}
}
