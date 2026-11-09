package page

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
)

// This file proves the incremental dirty path without timing anything:
// a one-operation change on a large op list stays near one region (via
// the SameOp-call and hash counters), the memoized diff agrees with the
// plain diff on scattered, multi-region, swap, and grow cases, and the
// single-diff recovery reports exactly one scan's work. All tests are
// race-clean: shared dirty state goes through atomics or the kid-index
// mutex.

// incGen is a deterministic xorshift64* source so fixtures replay.
type incGen struct{ s uint64 }

func (g *incGen) next() uint64 {
	g.s ^= g.s >> 12
	g.s ^= g.s << 25
	g.s ^= g.s >> 27
	return g.s * 2685821657736338717
}

// incDisplay builds a display of n fill ops spread across the frame, so
// scattered changes stay partial and a whole-frame change goes full.
func incDisplay(seed uint64, n int) *Display {
	g := &incGen{s: seed | 1}
	ops := make([]layout.DisplayOp, n)
	for i := range ops {
		ops[i] = layout.DisplayOp{
			Kind:  layout.DisplayOpFillRect,
			X:     float64(g.next() % 780),
			Y:     float64(g.next() % 580),
			W:     8,
			H:     8,
			R:     float64(g.next() % 4),
			G:     float64(g.next() % 4),
			B:     float64(g.next() % 4),
			Alpha: 1,
		}
	}
	return &Display{Width: 800, Height: 600, Ops: ops}
}

func incCopy(d *Display) *Display {
	return &Display{Width: d.Width, Height: d.Height, Ops: append([]layout.DisplayOp(nil), d.Ops...)}
}

// incCount converts a non-negative test count to uint64 for counter compares.
func incCount(v int) uint64 {
	if v < 0 {
		panic("page: negative test count")
	}
	return uint64(v)
}

// incSeed derives a deterministic fixture seed from a non-negative size.
func incSeed(n int, mult uint64) uint64 {
	return (incCount(n) + 1) * mult
}

// intn returns a pseudo-random index in [0, n). n must be positive.
func (g *incGen) intn(n int) int {
	if n <= 0 {
		panic("page: incGen.intn bound must be positive")
	}
	return int(g.next() % uint64(n)) //nolint:gosec // reason: the remainder is below the positive int bound n, so it always fits int.
}

func TestDirtyIncrementalSingleChangeCounters(t *testing.T) {
	const n = 2048
	prev := incDisplay(7, n)
	next := incCopy(prev)
	mid := n / 2
	next.Ops[mid].R = 7 - next.Ops[mid].R

	dirty.ResetDiffCounters()
	set, full, changed, memo := dirty.DiffWithBounds(prev, next)
	if memo == nil {
		t.Fatal("memoized diff must return a memo on the equal-length path")
	}
	defer memo.Free()

	if full {
		t.Fatalf("expected partial diff, got full")
	}
	if changed != 1 {
		t.Fatalf("expected 1 changed op, got %d", changed)
	}
	if set.N != 1 {
		t.Fatalf("expected 1 rect, got %d", set.N)
	}

	c := dirty.DiffCounters()
	if c.SameOp+c.Skipped != n {
		t.Fatalf("every op must scan or skip exactly once, got calls=%d skipped=%d n=%d",
			c.SameOp, c.Skipped, n)
	}
	if bound := uint64(n / 8); c.SameOp >= bound {
		t.Fatalf("one change cost %d SameOp calls, want under %d (n=%d)", c.SameOp, bound, n)
	}
	// Stamps are computed fresh per diff: both sides hash every op once.
	// No cross-frame hash cache exists, so this documents the hashing
	// volume the in-place patch paths require for correctness.
	if c.HashedOps != 2*n {
		t.Fatalf("expected %d hashed ops for a fresh two-sided stamp, got %d", 2*n, c.HashedOps)
	}

	// The memo must carry exactly the changed op's bound.
	if memo.Len() != n {
		t.Fatalf("expected memo length %d, got %d", n, memo.Len())
	}
	got, ok := memo.Get(mid)
	if !ok {
		t.Fatalf("memo must hold the changed op's bound")
	}
	if want := dirty.OpBounds(&next.Ops[mid], next); got != want {
		t.Fatalf("memo bound %v must equal OpBounds %v", got, want)
	}
}

func TestDirtyIncrementalGoldenMatchesPlain(t *testing.T) {
	sizes := []int{63, 64, 65, 200}
	for _, n := range sizes {
		cases := map[string]func() (*Display, *Display){
			"identical": func() (*Display, *Display) {
				prev := incDisplay(incSeed(n, 101), n)
				return prev, incCopy(prev)
			},
			"single_mid": func() (*Display, *Display) {
				prev := incDisplay(incSeed(n, 107), n)
				next := incCopy(prev)
				if n > 0 {
					next.Ops[n/2].G = 7 - next.Ops[n/2].G
				}
				return prev, next
			},
			"scattered": func() (*Display, *Display) {
				prev := incDisplay(incSeed(n, 113), n)
				next := incCopy(prev)
				g := &incGen{s: incSeed(n, 127) | 1}
				for k := 0; k < 5 && k < n; k++ {
					i := g.intn(imax(n, 1))
					next.Ops[i].R = 7 - next.Ops[i].R
				}
				return prev, next
			},
			"multi_region": func() (*Display, *Display) {
				prev := incDisplay(incSeed(n, 131), n)
				next := incCopy(prev)
				// Touch one op per 64-op region so every region rescans.
				for i := 0; i < n; i += 64 {
					next.Ops[i].B = 7 - next.Ops[i].B
				}
				return prev, next
			},
			"swap": func() (*Display, *Display) {
				prev := incDisplay(incSeed(n, 151), n)
				next := incCopy(prev)
				if n > 1 {
					next.Ops[0], next.Ops[n-1] = next.Ops[n-1], next.Ops[0]
				}
				return prev, next
			},
			"grow": func() (*Display, *Display) {
				prev := incDisplay(incSeed(n, 157), n)
				next := incCopy(prev)
				g := &incGen{s: incSeed(n, 163) | 1}
				for k := 0; k < 3; k++ {
					next.Ops = append(next.Ops, layout.DisplayOp{
						Kind: layout.DisplayOpFillRect,
						X:    float64(g.next() % 780), Y: float64(g.next() % 580),
						W: 8, H: 8, R: 1, Alpha: 1,
					})
				}
				return prev, next
			},
			"shrink": func() (*Display, *Display) {
				prev := incDisplay(incSeed(n, 167), n)
				next := incCopy(prev)
				if n > 3 {
					next.Ops = next.Ops[:n-3]
				} else {
					next.Ops = nil
				}
				return prev, next
			},
		}
		for name, make := range cases {
			t.Run(name, func(t *testing.T) {
				prev, next := make()
				plain, plainFull, plainChanged := dirty.Diff(prev, next)
				memoSet, memoFull, memoChanged, memo := dirty.DiffWithBounds(prev, next)
				if memo != nil {
					defer memo.Free()
				}
				if memoFull != plainFull || memoChanged != plainChanged ||
					memoSet.Added != plain.Added || memoSet.Removed != plain.Removed ||
					memoSet.N != plain.N || memoSet.Bounds() != plain.Bounds() {
					t.Fatalf("n=%d memoized diff diverged: got full=%v changed=%d added=%d removed=%d rects=%d bounds=%v, plain full=%v changed=%d added=%d removed=%d rects=%d bounds=%v",
						n, memoFull, memoChanged, memoSet.Added, memoSet.Removed, memoSet.N, memoSet.Bounds(),
						plainFull, plainChanged, plain.Added, plain.Removed, plain.N, plain.Bounds())
				}
				for i := 0; i < memoSet.N; i++ {
					if memoSet.Rects[i] != plain.Rects[i] {
						t.Fatalf("n=%d rect %d diverged: got %v, plain %v", n, i, memoSet.Rects[i], plain.Rects[i])
					}
				}
			})
		}
	}
}

func TestDirtyIncrementalRecordPartialMemoParity(t *testing.T) {
	const n = 200
	prev := incDisplay(23, n)
	next := incCopy(prev)
	next.Ops[10].R = 7 - next.Ops[10].R
	next.Ops[150].G = 7 - next.Ops[150].G

	set, full, changed, memo := dirty.DiffWithBounds(prev, next)
	if memo == nil {
		t.Fatal("memoized diff must return a memo here")
	}
	defer memo.Free()
	if full {
		t.Fatal("fixture must stay partial")
	}
	rects := set.Rects[:set.N]

	dirty.ResetDiffCounters()
	a := &Page{perf: true}
	a.recordPartial(changed, rects, next)
	b := &Page{perf: true}
	b.recordPartialBounds(changed, rects, memo, next)

	if a.stats != b.stats {
		t.Fatalf("memo counts %+v must equal fresh counts %+v", b.stats, a.stats)
	}
	if c := dirty.DiffCounters(); c.MemoHits == 0 {
		t.Fatal("memo path must serve at least one bound from the diff")
	}
}

func TestDirtyIncrementalSingleDiffRecovery(t *testing.T) {
	const n = 256
	painted := incDisplay(31, n)
	prev := incCopy(painted)
	display := incCopy(painted)
	for i := range display.Ops {
		display.Ops[i].R = 7 - display.Ops[i].R
	}

	// Reference: one diff of painted against display.
	_, refFull, refChanged := dirty.Diff(painted, display)
	if !refFull {
		t.Fatal("fixture must drive the recovery diff to full")
	}

	p := &Page{perf: true}
	p.painted = painted
	p.generation = 2
	p.fullGen = 1
	p.dirty.Full = true

	dirty.ResetDiffCounters()
	p.dirtyFromDisplay(prev, display)

	if !p.lastPatch.Full {
		t.Fatal("failed recovery must stay full")
	}
	if p.lastPatch.Changed != refChanged {
		t.Fatalf("single diff must report %d changed, got %d", refChanged, p.lastPatch.Changed)
	}
	if p.lastPatch.Added != 0 || p.lastPatch.Removed != 0 || p.lastPatch.Regions != 1 {
		t.Fatalf("unexpected recovery report %+v", p.lastPatch)
	}
	if !p.dirty.Full {
		t.Fatal("failed recovery must keep the full mark")
	}
	if p.fullGen != 3 {
		t.Fatalf("failed recovery must bind the mark to generation 3, got %d", p.fullGen)
	}
	// One diff scans each changed op once. The legacy double diff (one
	// against painted plus one against prev) would cost nearly twice.
	if c := dirty.DiffCounters(); c.SameOp != incCount(refChanged) {
		t.Fatalf("single diff must cost %d SameOp calls, got %d", refChanged, c.SameOp)
	}
	if p.stats.changedOps != refChanged || p.stats.dirtyRegions != 1 {
		t.Fatalf("unexpected recovery stats %+v", p.stats)
	}
}

func incBox(id string, parent int, x, y, w, h float64) Box {
	return Box{ID: id, ParentIndex: parent, X: x, Y: y, W: w, H: h}
}

func TestDirtyIncrementalSubtreeMark(t *testing.T) {
	// badge overflows card; tip is a foreign box inside card's rect.
	boxes := []Box{
		incBox("card", -1, 0, 0, 100, 100),
		incBox("title", 0, 10, 10, 60, 10),
		incBox("badge", 0, 90, 90, 40, 40),
		incBox("tip", -1, 5, 5, 10, 10),
	}

	p := &Page{}
	p.setBoxes(boxes)
	p.markElement("card")
	rects, ok := p.TakeDirtyRects()
	if !ok || len(rects) != 1 {
		t.Fatalf("expected one dirty rect, got %v ok=%v", rects, ok)
	}

	legacy := dirty.GrowBox(dirty.BoxRect(boxes[0]), boxes)
	if legacy != dirty.BoxRect(boxes[0]) {
		t.Fatalf("legacy GrowBox must be the id rect, got %v", legacy)
	}
	want := image.Rect(0, 0, 130, 130)
	if rects[0] != want {
		t.Fatalf("subtree mark must cover the overflowing child: got %v want %v", rects[0], want)
	}
	if union := rects[0].Union(legacy); union != rects[0] {
		t.Fatalf("subtree mark %v must cover legacy %v", rects[0], legacy)
	}
	if prev := p.dirty.Prev("card"); prev != want {
		t.Fatalf("noted prev rect must be the subtree union, got %v", prev)
	}
}

func TestDirtyIncrementalSubtreeMarkContained(t *testing.T) {
	boxes := []Box{
		incBox("card", -1, 0, 0, 100, 100),
		incBox("title", 0, 10, 10, 60, 10),
	}

	p := &Page{}
	p.setBoxes(boxes)
	p.markElement("card")
	rects, ok := p.TakeDirtyRects()
	if !ok || len(rects) != 1 {
		t.Fatalf("expected one dirty rect, got %v ok=%v", rects, ok)
	}

	legacy := dirty.GrowBox(dirty.BoxRect(boxes[0]), boxes)
	if rects[0] != legacy {
		t.Fatalf("contained subtree must equal legacy %v, got %v", legacy, rects[0])
	}
}

func TestDirtyIncrementalGrowBoxIdentity(t *testing.T) {
	boxes := []Box{
		incBox("a", -1, 0, 0, 100, 100),
		incBox("b", 0, 90, 90, 40, 40),
		incBox("c", -1, 5, 5, 10, 10),
	}
	for _, r := range []image.Rectangle{
		image.Rect(0, 0, 100, 100),
		image.Rect(0, 0, 1000, 1000),
		image.Rect(95, 95, 96, 96),
	} {
		if got := dirty.GrowBox(r, boxes); got != r {
			t.Fatalf("GrowBox must return a non-empty r unchanged, got %v want %v", got, r)
		}
	}
	if got := dirty.GrowBox(image.Rectangle{}, boxes); got != (image.Rectangle{}) {
		t.Fatalf("GrowBox must return an empty r unchanged, got %v", got)
	}
}

func refContent(t *testing.T, w, h int, boxes []Box) image.Rectangle {
	t.Helper()
	r := image.Rect(0, 0, w, h)
	for _, b := range boxes {
		r = r.Union(dirty.BoxRect(b))
	}
	return r
}

func TestDirtyIncrementalContentRect(t *testing.T) {
	p := &Page{}
	p.width, p.height = 800, 600

	steps := []struct {
		name  string
		boxes []Box
		exact bool
	}{
		{"initial", []Box{incBox("a", -1, 0, 0, 100, 100), incBox("b", -1, 200, 200, 50, 50)}, true},
		{"grow", []Box{incBox("a", -1, 0, 0, 100, 100), incBox("b", -1, 200, 200, 50, 50), incBox("c", -1, 900, 700, 50, 50)}, true},
		{"grow_edge", []Box{incBox("a", -1, 0, 0, 100, 100), incBox("b", -1, 200, 200, 50, 50), incBox("c", -1, 900, 700, 60, 60)}, true},
		{"shrink_edge", []Box{incBox("a", -1, 0, 0, 100, 100), incBox("b", -1, 200, 200, 50, 50), incBox("c", -1, 900, 700, 50, 50)}, true},
		{"remove_inner", []Box{incBox("c", -1, 900, 700, 50, 50)}, true},
		{"remove_edge", []Box{incBox("a", -1, 0, 0, 100, 100)}, true},
		{"move_out", []Box{incBox("a", -1, 1000, 0, 10, 10)}, true},
		{"reorder_same", []Box{incBox("a", -1, 1000, 0, 10, 10)}, true},
		// A duplicate id recomputes exactly: the second copy's rect
		// differs without covering the first, so the union rebuilds.
		{"duplicate_add", []Box{incBox("a", -1, 1000, 0, 10, 10), incBox("a", -1, 900, 700, 5, 5)}, true},
		// Dropping one copy of a duplicate id keeps the survivor
		// covered: the lookup still finds the id, so the union
		// over-covers rather than recomputes. Over-cover repaints
		// more, never less.
		{"duplicate_drop", []Box{incBox("a", -1, 1000, 0, 10, 10)}, false},
	}

	for _, s := range steps {
		p.setBoxes(s.boxes)
		if !p.contentOK {
			t.Fatalf("%s: content union must stay current", s.name)
		}
		ref := refContent(t, 800, 600, s.boxes)
		got := p.contentRect()
		if s.exact {
			if got != ref {
				t.Fatalf("%s: content rect must equal the full union: got %v want %v", s.name, got, ref)
			}
			continue
		}
		if union := got.Union(ref); union != got {
			t.Fatalf("%s: content rect %v must cover the full union %v", s.name, got, ref)
		}
	}
}

func imax(a, b int) int {
	if a > b {
		return a
	}
	return b
}
