package replay

import "testing"

// TestActiveLayerIDsReusesScratch pins PERF-MOT-02 for the replay side.
func TestActiveLayerIDsReusesScratch(t *testing.T) {
	var r Renderer
	r.layers.frame.active = []*Layer{{ElementID: "a"}, {ElementID: ""}, {ElementID: "b"}}
	if got := r.ActiveLayerIDs(); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("ids = %v, want [a b]", got)
	}
	if allocs := testing.AllocsPerRun(100, func() { r.ActiveLayerIDs() }); allocs != 0 {
		t.Fatalf("ActiveLayerIDs allocs/op = %v, want 0", allocs)
	}
}
