package page

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func TestPageMotionIndexCaching(t *testing.T) {
	p := &Page{}

	// 1. Nil display or empty keyframes returns shared EmptyIndex with 0 allocs.
	emptyDisplay := &Display{
		Ops: []layout.DisplayOp{{ID: 1, Kind: layout.DisplayOpFillRect}},
	}
	idx := p.updateMotionIndex(nil, emptyDisplay)
	if idx != cssmotion.EmptyIndex() {
		t.Fatalf("expected emptyIndex, got %v", idx)
	}

	// 2. Display with keyframes builds a new index.
	keyframes := []layout.KeyframesRule{
		{
			Name: "fade",
			Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: "opacity", Value: "0"}}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: "opacity", Value: "1"}}},
			},
		},
	}
	d1 := &Display{
		Ops:       []layout.DisplayOp{{ID: 1, Kind: layout.DisplayOpFillRect}},
		OpOwners:  []int{0},
		Boxes:     []layout.Box{{ID: "b1", OperationIDs: []uint64{1}}},
		Keyframes: keyframes,
	}

	idx1 := p.updateMotionIndex(nil, d1)
	if idx1 == nil || idx1 == cssmotion.EmptyIndex() {
		t.Fatalf("expected non-empty index, got %v", idx1)
	}
	if p.motionIndex != idx1 {
		t.Fatalf("expected p.motionIndex to be cached, got %v", p.motionIndex)
	}

	// 3. Same keyframes (pointer equality) reuses the cached index.
	d2 := &Display{
		Ops:       []layout.DisplayOp{{ID: 1, Kind: layout.DisplayOpFillRect}},
		OpOwners:  []int{0},
		Boxes:     []layout.Box{{ID: "b1", OperationIDs: []uint64{1}}},
		Keyframes: keyframes, // identical slice
	}

	idx2 := p.updateMotionIndex(d1, d2)
	if idx2 != idx1 {
		t.Fatalf("expected reused index %p, got %p", idx1, idx2)
	}

	// 4. Zero allocs on repeated commits with same keyframes.
	allocs := testing.AllocsPerRun(100, func() {
		idx := p.updateMotionIndex(d1, d2)
		if idx == nil {
			t.Fatal("unexpected nil index")
		}
	})
	if allocs != 0 {
		t.Fatalf("allocs = %v, want 0", allocs)
	}

	// 5. Zero allocs when both have empty keyframes.
	p.motionIndex = nil
	allocsEmpty := testing.AllocsPerRun(100, func() {
		idx := p.updateMotionIndex(emptyDisplay, emptyDisplay)
		if idx != cssmotion.EmptyIndex() {
			t.Fatal("expected emptyIndex")
		}
	})
	if allocsEmpty != 0 {
		t.Fatalf("allocsEmpty = %v, want 0", allocsEmpty)
	}
}
