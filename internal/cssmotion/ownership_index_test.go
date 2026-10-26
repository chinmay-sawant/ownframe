package cssmotion

import (
	"slices"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// TestIndexSelectMatchesDirect pins the shared-index contract: every
// selection through a per-display Index returns exactly what the direct scan
// returns, across owner, subtree, fallback, anonymous, and out-of-range
// boxes, with and without an owner map.
func TestIndexSelectMatchesDirect(t *testing.T) {
	kinds := []Kind{KindFill, KindInk, KindStroke, KindSubtree}
	ownersVariants := map[string][]int{
		"owned":      {0, 0, 0, 0, 1},
		"unowned":    nil,
		"mismatched": {0, 0},
	}
	for name, owners := range ownersVariants {
		d := selectDisplay(owners)
		idx := NewIndex(d)
		for bi := range d.Boxes {
			for _, boxIndex := range []int{bi, -1, 99} {
				for _, kind := range kinds {
					want := Select(d, d.Boxes[bi], boxIndex, kind)
					got := idx.Select(d.Boxes[bi], boxIndex, kind)
					if !slices.Equal(got, want) {
						t.Fatalf("%s box %d index %d kind %d = %v, want %v",
							name, bi, boxIndex, kind, got, want)
					}
				}
			}
		}
	}
}

// TestIndexSelectGeometricMatchesDirect covers boxes that predate operation
// ids: the geometric fallback must agree with the direct scan too.
func TestIndexSelectGeometricMatchesDirect(t *testing.T) {
	d := selectDisplay([]int{0, 0, 0, 0, 1})
	geo := layout.Box{ID: "geo", X: 1, Y: 1, W: 8, H: 8}
	idx := NewIndex(d)
	for _, kind := range []Kind{KindFill, KindInk, KindStroke, KindSubtree} {
		want := Select(d, geo, 0, kind)
		if got := idx.Select(geo, 0, kind); !slices.Equal(got, want) {
			t.Fatalf("geometric kind %d = %v, want %v", kind, got, want)
		}
	}
}

// TestIndexSelectNilDisplayMatchesDirect checks that an empty index selects
// nothing, like Select on nil.
func TestIndexSelectNilDisplayMatchesDirect(t *testing.T) {
	idx := NewIndex(nil)
	box := layout.Box{ID: "x", OperationIDs: []uint64{1}}
	for _, kind := range []Kind{KindFill, KindInk, KindStroke, KindSubtree} {
		if got := idx.Select(box, 0, kind); got != nil {
			t.Fatalf("nil index kind %d = %v, want nil", kind, got)
		}
	}
	if got := idx.ImageIndex(box, 0); got != nil {
		t.Fatalf("nil image index = %v, want nil", got)
	}
}

// imageDisplay is one SVG root with a background fill before its raster plus
// an unrelated trailing image, so the tests can check the last-image rule.
func imageDisplay(owners []int) *layout.Display {
	return &layout.Display{
		PointsPerPixel: 1,
		Ops: []layout.DisplayOp{
			{ID: 1, Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 8, H: 8},
			{ID: 2, Kind: layout.DisplayOpImage, X: 0, Y: 0, W: 8, H: 8},
			{ID: 3, Kind: layout.DisplayOpImage, X: 20, Y: 20, W: 8, H: 8},
		},
		OpOwners: owners,
		Boxes: []layout.Box{
			{ID: "icon", Tag: "svg", X: 0, Y: 0, W: 8, H: 8,
				OperationIDs: []uint64{1, 2}, ParentIndex: -1},
		},
	}
}

// TestIndexImageMatchesTiers checks the raster lookup across the owner,
// id-set, geometric, and empty tiers.
func TestIndexImageMatchesTiers(t *testing.T) {
	d := imageDisplay([]int{0, 0, 1})
	idx := NewIndex(d)
	if got := idx.ImageIndex(d.Boxes[0], 0); !slices.Equal(got, []int{1}) {
		t.Fatalf("owner image = %v, want the box raster", got)
	}

	d = imageDisplay(nil)
	idx = NewIndex(d)
	if got := idx.ImageIndex(d.Boxes[0], 0); !slices.Equal(got, []int{1}) {
		t.Fatalf("id-set image = %v, want the box raster", got)
	}

	geo := layout.Box{ID: "geo", X: 0, Y: 0, W: 8, H: 8}
	if got := idx.ImageIndex(geo, -1); !slices.Equal(got, []int{1}) {
		t.Fatalf("geometric image = %v, want the last image in the area", got)
	}

	bare := &layout.Display{PointsPerPixel: 1,
		Ops:      []layout.DisplayOp{{ID: 1, Kind: layout.DisplayOpFillRect}},
		OpOwners: []int{0},
		Boxes:    []layout.Box{{ID: "bare", OperationIDs: []uint64{1}}},
	}
	if got := NewIndex(bare).ImageIndex(bare.Boxes[0], 0); len(got) != 0 {
		t.Fatalf("imageless image = %v, want empty", got)
	}
}
