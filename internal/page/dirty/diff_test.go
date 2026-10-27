package dirty

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestDiffEqualLength(t *testing.T) {
	prev := &layout.Display{
		Width:  400,
		Height: 300,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, R: 10, G: 20, B: 30, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 50, Y: 50, W: 30, H: 30, R: 40, G: 50, B: 60, Alpha: 1},
		},
	}
	next := &layout.Display{
		Width:  400,
		Height: 300,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, R: 10, G: 20, B: 30, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 50, Y: 50, W: 30, H: 30, R: 99, G: 50, B: 60, Alpha: 1}, // changed color
		},
	}

	set, full, changed := Diff(prev, next)
	if full {
		t.Fatalf("expected partial diff, got full")
	}
	if changed != 1 {
		t.Fatalf("expected 1 changed op, got %d", changed)
	}
	if set.N != 1 {
		t.Fatalf("expected 1 rect, got %d", set.N)
	}
}

func TestDiffByPosition_AddAndRemove(t *testing.T) {
	prev := &layout.Display{
		Width:  400,
		Height: 300,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, R: 10, G: 20, B: 30, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 100, Y: 100, W: 40, H: 40, R: 10, G: 20, B: 30, Alpha: 1},
		},
	}
	// next has 3 ops: first is unchanged, second is deleted, two new ops added
	next := &layout.Display{
		Width:  400,
		Height: 300,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, R: 10, G: 20, B: 30, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 200, Y: 200, W: 30, H: 30, R: 70, G: 80, B: 90, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 250, Y: 200, W: 30, H: 30, R: 70, G: 80, B: 90, Alpha: 1},
		},
	}

	set, full, changed := Diff(prev, next)
	if full {
		t.Fatalf("expected partial diff, got full")
	}
	if changed != 3 { // 1 removed + 2 added
		t.Fatalf("expected 3 changed ops, got %d", changed)
	}
	if set.Removed != 1 {
		t.Fatalf("expected 1 removed, got %d", set.Removed)
	}
	if set.Added != 2 {
		t.Fatalf("expected 2 added, got %d", set.Added)
	}
}

func TestDiffByPosition_FIFOOrderDuplicateKeys(t *testing.T) {
	// Two ops in prev with identical key (same kind, geom), but different color
	prev := &layout.Display{
		Width:  400,
		Height: 300,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, R: 1, G: 0, B: 0, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, R: 2, G: 0, B: 0, Alpha: 1},
		},
	}
	// Three ops in next with identical key: first matches prev[0], second matches prev[1], third is added
	next := &layout.Display{
		Width:  400,
		Height: 300,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, R: 1, G: 0, B: 0, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, R: 2, G: 0, B: 0, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, R: 3, G: 0, B: 0, Alpha: 1},
		},
	}

	set, full, changed := Diff(prev, next)
	if full {
		t.Fatalf("expected partial diff, got full")
	}
	// Only the third op is new, first two match exactly in FIFO order
	if changed != 1 {
		t.Fatalf("expected 1 changed op (the added one), got %d", changed)
	}
	if set.Added != 1 {
		t.Fatalf("expected 1 added, got %d", set.Added)
	}
	if set.Removed != 0 {
		t.Fatalf("expected 0 removed, got %d", set.Removed)
	}
}

func TestDiffByPosition_EmptyDisplays(t *testing.T) {
	empty := &layout.Display{Width: 400, Height: 300}
	nonEmpty := &layout.Display{
		Width:  400,
		Height: 300,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, R: 1, G: 0, B: 0, Alpha: 1},
		},
	}

	// Prev non-empty, Next empty
	set, _, changed := Diff(nonEmpty, empty)
	if changed != 1 || set.Removed != 1 {
		t.Fatalf("expected 1 removed, got changed=%d removed=%d", changed, set.Removed)
	}

	// Prev empty, Next non-empty
	set2, _, changed2 := Diff(empty, nonEmpty)
	if changed2 != 1 || set2.Added != 1 {
		t.Fatalf("expected 1 added, got changed=%d added=%d", changed2, set2.Added)
	}
}

func TestDiffArena_Allocs(t *testing.T) {
	const n = 200
	prevOps := make([]layout.DisplayOp, n)
	nextOps := make([]layout.DisplayOp, n+1)

	for i := 0; i < n; i++ {
		prevOps[i] = layout.DisplayOp{
			Kind:  layout.DisplayOpFillRect,
			X:     float64(i % 10 * 10),
			Y:     float64(i / 10 * 10),
			W:     8,
			H:     8,
			R:     float64(i % 256),
			Alpha: 1,
		}
		nextOps[i] = prevOps[i]
	}
	nextOps[n] = layout.DisplayOp{
		Kind:  layout.DisplayOpFillRect,
		X:     150,
		Y:     150,
		W:     8,
		H:     8,
		R:     1,
		Alpha: 1,
	}

	prev := &layout.Display{Width: 600, Height: 600, Ops: prevOps}
	next := &layout.Display{Width: 600, Height: 600, Ops: nextOps}

	// Warm pool
	Diff(prev, next)

	allocs := testing.AllocsPerRun(10, func() {
		Diff(prev, next)
	})

	if allocs > 0 {
		t.Fatalf("expected 0 allocs per run from pooled arena, got %f", allocs)
	}
}
