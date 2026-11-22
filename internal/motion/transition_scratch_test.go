package motion

import (
	"fmt"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestCaptureTransitionsReleasesOversizedScratchMaps(t *testing.T) {
	t.Parallel()

	const n = maxCaptureScratchMap + 16
	prev := &Display{Boxes: make([]layout.Box, n)}
	next := &Display{Boxes: make([]layout.Box, n+1)}
	roots := make([]string, n)
	tr := []layout.Transition{{Property: "opacity", DurationMS: 200}}

	for i := range n {
		id := fmt.Sprintf("el-%d", i)
		roots[i] = id
		prev.Boxes[i] = layout.Box{ID: id, Transitions: tr}
		next.Boxes[i] = layout.Box{ID: id, Transitions: tr}
	}
	// Extra trailing box makes previous and next non-aligned so capOldBoxes
	// and capOrdinals are both populated past maxCaptureScratchMap.
	next.Boxes[n] = layout.Box{ID: "extra"}

	r := New()
	_ = r.CaptureTransitionsScoped(prev, next, roots)

	if r.capScope != nil {
		t.Fatal("oversized capScope was not released")
	}
	if r.capIDs != nil {
		t.Fatal("oversized capIDs was not released")
	}
	if r.capOrdinals != nil {
		t.Fatal("oversized capOrdinals was not released")
	}
	if r.capOldBoxes != nil {
		t.Fatal("oversized capOldBoxes was not released")
	}

	// A small capture keeps scratch maps allocated and cleared for reuse.
	smallPrev := &Display{Boxes: []layout.Box{{ID: "a", Transitions: tr}}}
	smallNext := &Display{Boxes: []layout.Box{{ID: "a", Transitions: tr}, {ID: "b"}}}
	_ = r.CaptureTransitionsScoped(smallPrev, smallNext, []string{"a"})

	if r.capScope == nil || len(r.capScope) != 0 {
		t.Fatalf("small capScope = %v, want non-nil empty map", r.capScope)
	}
	if r.capIDs == nil || len(r.capIDs) != 0 {
		t.Fatalf("small capIDs = %v, want non-nil empty map", r.capIDs)
	}
	if r.capOrdinals == nil || len(r.capOrdinals) != 0 {
		t.Fatalf("small capOrdinals = %v, want non-nil empty map", r.capOrdinals)
	}
	if r.capOldBoxes == nil || len(r.capOldBoxes) != 0 {
		t.Fatalf("small capOldBoxes = %v, want non-nil empty map", r.capOldBoxes)
	}
}
