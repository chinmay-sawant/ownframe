package dirty

import (
	"math"
	"sync/atomic"

	"github.com/chinmay-sawant/blinkless/layout"
)

// diffArena holds reusable flat index structures for diffByPosition.
type diffArena struct {
	head []int
	next []int
	seen []bool
}

// diffArenaCache is a single reusable arena slot. One Swap takes it, one
// Swap returns it. A sync.Pool was here before, but under -race the pool
// bookkeeping allocates once per Get/Put pair, which broke the zero-alloc
// guarantee TestDiffArena_Allocs asserts. Plain atomics reuse the same
// addresses and stay allocation-free in steady state, race or not.
// Concurrent diffs may each fall back to a fresh arena; correctness never
// depends on a cache hit.
//
//nolint:gochecknoglobals // reason: process-wide single reusable diff arena across frames, not shared mutable state beyond the atomic slot.
var diffArenaCache atomic.Pointer[diffArena]

const maxPooledArenaCap = 65536

const (
	minDiffBuckets   = 16
	diffBucketGrowth = 2
)

func getDiffArena(numBuckets, numOps int) *diffArena {
	a := diffArenaCache.Swap(nil)
	if a == nil {
		a = &diffArena{}
	}
	if cap(a.head) < numBuckets {
		a.head = make([]int, numBuckets)
	} else {
		a.head = a.head[:numBuckets]
	}
	for i := range a.head {
		a.head[i] = -1
	}

	if cap(a.next) < numOps {
		a.next = make([]int, numOps)
	} else {
		a.next = a.next[:numOps]
	}

	if cap(a.seen) < numOps {
		a.seen = make([]bool, numOps)
	} else {
		a.seen = a.seen[:numOps]
		clear(a.seen)
	}
	return a
}

func putDiffArena(a *diffArena) {
	if cap(a.head) > maxPooledArenaCap || cap(a.next) > maxPooledArenaCap {
		return
	}
	if old := diffArenaCache.Swap(a); old != nil {
		// Slot was refilled by a concurrent diff while this arena was
		// in flight; drop the evicted one so at most one stays cached.
		_ = old
	}
}

func nextPowerOf2(n int) int {
	size := minDiffBuckets
	for size < n {
		size *= diffBucketGrowth
	}
	return size
}

func hashKey(k opKey) uint64 {
	const offsetBasis = 14695981039346656037
	const prime = 1099511628211
	h := (offsetBasis ^ uint64(k.kind)) * prime
	h = (h ^ floatBits(k.x)) * prime
	h = (h ^ floatBits(k.y)) * prime
	h = (h ^ floatBits(k.w)) * prime
	h = (h ^ floatBits(k.h)) * prime
	return h
}

func floatBits(v float64) uint64 {
	if v == 0 {
		return 0
	}
	return math.Float64bits(v)
}

// Diff returns the regions covering the operations that differ between
// prev and next. full reports a difference too scattered or too large to
// merge, and changed counts the operations that differ. A change stays
// partial while its bounding box is under a third of the frame and its
// merged rectangle list fits MaxRects.
func Diff(prev, next *layout.Display) (Set, bool, int) {
	if prev.Width != next.Width || prev.Height != next.Height {
		return Set{}, true, len(next.Ops)
	}

	if len(prev.Ops) != len(next.Ops) {
		return diffByPosition(prev, next)
	}

	d := newSet(next.Width, next.Height)
	changed := 0

	for i := range prev.Ops {
		if SameOp(&prev.Ops[i], &next.Ops[i]) {
			continue
		}

		changed++
		d.add(OpBounds(&prev.Ops[i], prev).Union(OpBounds(&next.Ops[i], next)))
		if d.full {
			return d, true, changed
		}
	}

	return d, false, changed
}

// diffByPosition pairs operations by kind and geometry when the two lists
// differ in length. An operation with no partner dirties its bounds, and
// changed counts every unpaired or repainted operation.
func diffByPosition(prev, next *layout.Display) (Set, bool, int) {
	d := newSet(next.Width, next.Height)
	changed := 0

	if len(next.Ops) == 0 {
		for i := range prev.Ops {
			changed++
			d.Removed++
			d.add(OpBounds(&prev.Ops[i], prev))
			if d.full {
				return d, true, changed
			}
		}
		return d, false, changed
	}

	if len(prev.Ops) == 0 {
		for j := range next.Ops {
			changed++
			d.Added++
			d.add(OpBounds(&next.Ops[j], next))
			if d.full {
				return d, true, changed
			}
		}
		return d, false, changed
	}

	numOps := len(next.Ops)
	numBuckets := nextPowerOf2(numOps * diffBucketGrowth)
	mask := uint64(numBuckets - 1) //nolint:gosec // reason: numBuckets is a nextPowerOf2 result always >= 16, so numBuckets-1 is non-negative and fits uint64.

	arena := getDiffArena(numBuckets, numOps)
	defer putDiffArena(arena)

	// Build bucket chains in reverse order so arena.head[b] yields ops in
	// FIFO index order.
	for i := numOps - 1; i >= 0; i-- {
		k := keyOf(&next.Ops[i])
		b := hashKey(k) & mask
		arena.next[i] = arena.head[b]
		arena.head[b] = i
	}

	for i := range prev.Ops {
		k := keyOf(&prev.Ops[i])
		b := hashKey(k) & mask

		found := -1
		prevIdx := -1
		for curr := arena.head[b]; curr != -1; {
			if keyOf(&next.Ops[curr]) == k {
				found = curr
				if prevIdx == -1 {
					arena.head[b] = arena.next[curr]
				} else {
					arena.next[prevIdx] = arena.next[curr]
				}
				break
			}
			prevIdx = curr
			curr = arena.next[curr]
		}

		if found == -1 {
			changed++
			d.Removed++
			d.add(OpBounds(&prev.Ops[i], prev))
			if d.full {
				return d, true, changed
			}

			continue
		}

		j := found
		arena.seen[j] = true

		if SameOp(&prev.Ops[i], &next.Ops[j]) {
			continue
		}

		changed++
		d.add(OpBounds(&prev.Ops[i], prev).Union(OpBounds(&next.Ops[j], next)))
		if d.full {
			return d, true, changed
		}
	}

	for j := range next.Ops {
		if arena.seen[j] {
			continue
		}

		changed++
		d.Added++
		d.add(OpBounds(&next.Ops[j], next))
		if d.full {
			return d, true, changed
		}
	}

	return d, false, changed
}

// opKey identifies an operation by its paint kind and geometry.
type opKey struct {
	kind       layout.DisplayKind
	x, y, w, h float64
}

func keyOf(op *layout.DisplayOp) opKey {
	return opKey{kind: op.Kind, x: op.X, y: op.Y, w: op.W, h: op.H}
}
