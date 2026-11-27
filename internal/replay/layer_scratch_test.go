package replay

import (
	"slices"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

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

func TestActiveLayerBoxIndices(t *testing.T) {
	var r Renderer
	d := &layout.Display{}
	other := &layout.Display{}
	r.layers.frame.display = d
	r.layers.frame.active = []*Layer{
		{ElementID: "", BoxIndex: 7},
		{ElementID: "card", BoxIndex: 2},
		{ElementID: "", BoxIndex: 7},
		{ElementID: "", BoxIndex: 5},
	}

	got := r.ActiveLayerBoxIndices(d)
	want := []int{2, 5, 7}
	if !slices.Equal(got, want) {
		t.Fatalf("boxes = %v, want %v", got, want)
	}
	if allocs := testing.AllocsPerRun(100, func() { r.ActiveLayerBoxIndices(d) }); allocs != 0 {
		t.Fatalf("ActiveLayerBoxIndices allocs/op = %v, want 0", allocs)
	}
	if got := r.ActiveLayerBoxIndices(other); len(got) != 0 {
		t.Fatalf("mismatched display boxes = %v, want empty", got)
	}
	if got := r.ActiveLayerBoxIndices(nil); len(got) != 0 {
		t.Fatalf("nil display boxes = %v, want empty", got)
	}
}
