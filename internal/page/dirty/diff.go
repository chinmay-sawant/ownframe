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
	// seen marks next-slice matches by checkout epoch: seen[j] equals the
	// arena epoch only when op j paired this diff, so reuse needs no
	// O(N) clear.
	seen  []uint32
	epoch uint32
	// keys caches keyOf for each next-slice op, computed once per diff,
	// so the match walk compares the stored key instead of rebuilding it
	// per chain visit.
	keys []opKey
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

// diffRegionSize is how many operations one content stamp covers. A commit
// that touches one operation re-compares at most this many per-index pairs,
// while untouched regions skip SameOp entirely.
const diffRegionSize = 64

const (
	stampOffsetBasis uint64 = 14695981039346656037
	stampPrime       uint64 = 1099511628211
)

//nolint:gochecknoglobals // reason: process-wide diff comparator counters; atomics keep concurrent diffs race-safe and tests reset them per case.
var (
	// sameOpCalls counts SameOp invocations across every diff path.
	sameOpCalls atomic.Uint64
	// diffSkippedOps counts operations that skipped SameOp by region stamp.
	diffSkippedOps atomic.Uint64
	// sameImageIdentity counts SameOp image matches decided by shared
	// backing array instead of a byte scan.
	sameImageIdentity atomic.Uint64
	// regionHashedOps counts operations folded into region stamps. Stamps
	// are computed fresh per diff: retained operations are patched in
	// place between frames, so no hash survives past the diff that
	// computed it.
	regionHashedOps atomic.Uint64
	// boundsMemoHits counts BoundsMemo.Get calls served from the diff's
	// stored bounds instead of a fresh OpBounds computation.
	boundsMemoHits atomic.Uint64
	// boundsMemoMiss counts BoundsMemo.Get calls that fell back to
	// OpBounds because the diff never computed that index.
	boundsMemoMiss atomic.Uint64
)

// Counters is the diff observability snapshot: SameOp calls, region
// stamp skips, stamped operations, and memo hits and misses. Tests use
// it to prove a one-operation change stays near one region without
// timing the diff.
type Counters struct {
	SameOp    uint64
	Skipped   uint64
	HashedOps uint64
	MemoHits  uint64
	MemoMiss  uint64
}

// DiffCounters reports the diff counters accumulated so far.
func DiffCounters() Counters {
	return Counters{
		SameOp:    sameOpCalls.Load(),
		Skipped:   diffSkippedOps.Load(),
		HashedOps: regionHashedOps.Load(),
		MemoHits:  boundsMemoHits.Load(),
		MemoMiss:  boundsMemoMiss.Load(),
	}
}

// ResetDiffCounters zeroes the diff counters. Tests reset them per case;
// production code never calls it.
func ResetDiffCounters() {
	sameOpCalls.Store(0)
	diffSkippedOps.Store(0)
	sameImageIdentity.Store(0)
	regionHashedOps.Store(0)
	boundsMemoHits.Store(0)
	boundsMemoMiss.Store(0)
}

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
		a.seen = make([]uint32, numOps)
	} else {
		a.seen = a.seen[:numOps]
	}
	// The epoch marks this checkout: seen entries stamped with older
	// epochs never match, so reuse needs no O(N) clear. The wrap resets
	// once per 2^32 checkouts, which a frame loop never approaches.
	a.epoch++
	if a.epoch == 0 {
		clear(a.seen)
		a.epoch = 1
	}

	if cap(a.keys) < numOps {
		a.keys = make([]opKey, numOps)
	} else {
		a.keys = a.keys[:numOps]
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
	set, full, changed := diffInner(prev, next, nil)
	return set, full, changed
}

// DiffWithBounds diffs like Diff and also returns the next-side rounded
// op bounds the diff already computed, so the perf accounting pass
// reuses them instead of recomputing every bound a second time. The
// caller frees the memo after use; a nil memo means the diff took the
// size-mismatch early path and computed no bounds.
func DiffWithBounds(prev, next *layout.Display) (Set, bool, int, *BoundsMemo) {
	if prev.Width != next.Width || prev.Height != next.Height {
		return Set{}, true, len(next.Ops), nil
	}

	memo := rentBoundsMemo(len(next.Ops))
	set, full, changed := diffInner(prev, next, memo)
	return set, full, changed, memo
}

// diffInner is the shared equal-length and by-position diff. A non-nil
// memo collects every next-side bound the scan computes; a nil memo
// skips the stores, so Diff pays nothing for bounds it never reuses.
func diffInner(prev, next *layout.Display, memo *BoundsMemo) (Set, bool, int) {
	if prev.Width != next.Width || prev.Height != next.Height {
		return Set{}, true, len(next.Ops)
	}

	if len(prev.Ops) != len(next.Ops) {
		return diffByPosition(prev, next, memo)
	}

	// Same generation: one backing store cannot differ from itself.
	if prev == next {
		return newSet(next.Width, next.Height), false, 0
	}

	d := newSet(next.Width, next.Height)
	changed := 0

	n := len(prev.Ops)
	for base := 0; base < n; base += diffRegionSize {
		end := base + diffRegionSize
		if end > n {
			end = n
		}
		// Stamps are computed fresh per diff: internal/frame patches
		// retained operations in place, so no cross-frame cache stays
		// valid. A matching stamp plus matching Font/Grid identities
		// implies SameOp on every index, and the region skips the
		// per-index scan.
		if ps, ok := stampRegionOps(prev.Ops[base:end]); ok {
			if ns, ok := stampRegionOps(next.Ops[base:end]); ok && ps == ns &&
				regionPtrsEqual(prev.Ops[base:end], next.Ops[base:end]) {
				diffSkippedOps.Add(uint64(end - base)) //nolint:gosec // reason: end-base is a positive region length that fits uint64.
				continue
			}
		}

		for i := base; i < end; i++ {
			if SameOp(&prev.Ops[i], &next.Ops[i]) {
				continue
			}

			changed++
			pb := OpBounds(&prev.Ops[i], prev)
			nb := OpBounds(&next.Ops[i], next)
			if memo != nil {
				memo.Set(i, nb)
			}
			d.add(pb.Union(nb))
			if d.full {
				return d, true, changed
			}
		}
	}

	return d, false, changed
}

// diffByPosition pairs operations by kind and geometry when the two lists
// differ in length. An operation with no partner dirties its bounds, and
// changed counts every unpaired or repainted operation. A non-nil memo
// collects every next-side bound the scan computes, keyed by absolute
// index into next.Ops.
func diffByPosition(prev, next *layout.Display, memo *BoundsMemo) (Set, bool, int) {
	d := newSet(next.Width, next.Height)
	changed := 0

	prefix := 0
	for prefix < len(prev.Ops) && prefix < len(next.Ops) && SameOp(&prev.Ops[prefix], &next.Ops[prefix]) {
		prefix++
	}

	suffix := 0
	for suffix < len(prev.Ops)-prefix && suffix < len(next.Ops)-prefix &&
		SameOp(&prev.Ops[len(prev.Ops)-1-suffix], &next.Ops[len(next.Ops)-1-suffix]) {
		suffix++
	}

	prevSlice := prev.Ops[prefix : len(prev.Ops)-suffix]
	nextSlice := next.Ops[prefix : len(next.Ops)-suffix]

	if len(prevSlice) == 0 && len(nextSlice) == 0 {
		return d, false, 0
	}

	if len(nextSlice) == 0 {
		for i := range prevSlice {
			changed++
			d.Removed++
			d.add(OpBounds(&prevSlice[i], prev))
			if d.full {
				return d, true, changed
			}
		}
		return d, false, changed
	}

	if len(prevSlice) == 0 {
		for j := range nextSlice {
			changed++
			d.Added++
			nb := OpBounds(&nextSlice[j], next)
			if memo != nil {
				memo.Set(prefix+j, nb)
			}
			d.add(nb)
			if d.full {
				return d, true, changed
			}
		}
		return d, false, changed
	}

	numOps := len(nextSlice)
	numBuckets := nextPowerOf2(numOps * diffBucketGrowth)
	mask := uint64(numBuckets - 1) //nolint:gosec // reason: numBuckets is a nextPowerOf2 result always >= 16, so numBuckets-1 is non-negative and fits uint64.

	arena := getDiffArena(numBuckets, numOps)
	defer putDiffArena(arena)
	epoch := arena.epoch

	// Cache each next key once so the match walk below compares the stored
	// key instead of rebuilding it per chain visit.
	for i := range nextSlice {
		arena.keys[i] = keyOf(&nextSlice[i])
	}

	// Build bucket chains in reverse order so arena.head[b] yields ops in
	// FIFO index order.
	for i := numOps - 1; i >= 0; i-- {
		b := hashKey(arena.keys[i]) & mask
		arena.next[i] = arena.head[b]
		arena.head[b] = i
	}

	for i := range prevSlice {
		k := keyOf(&prevSlice[i])
		b := hashKey(k) & mask

		found := -1
		prevIdx := -1
		for curr := arena.head[b]; curr != -1; {
			if arena.keys[curr] == k {
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
			d.add(OpBounds(&prevSlice[i], prev))
			if d.full {
				return d, true, changed
			}

			continue
		}

		j := found
		arena.seen[j] = epoch

		if SameOp(&prevSlice[i], &nextSlice[j]) {
			continue
		}

		changed++
		pb := OpBounds(&prevSlice[i], prev)
		nb := OpBounds(&nextSlice[j], next)
		if memo != nil {
			memo.Set(prefix+j, nb)
		}
		d.add(pb.Union(nb))
		if d.full {
			return d, true, changed
		}
	}

	for j := range nextSlice {
		if arena.seen[j] == epoch {
			continue
		}

		changed++
		d.Added++
		nb := OpBounds(&nextSlice[j], next)
		if memo != nil {
			memo.Set(prefix+j, nb)
		}
		d.add(nb)
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

// stampRegionOps hashes every SameOp input in ops except the Font and Grid
// identities, which regionPtrsEqual verifies before a skip. ok is false
// when a NaN float would make == disagree with the hash, forcing the
// per-index path so results stay identical to a full scan.
func stampRegionOps(ops []layout.DisplayOp) (uint64, bool) {
	h := stampOffsetBasis
	hashed := 0
	for i := range ops {
		var ok bool
		if h, ok = mixOpStamp(h, &ops[i]); !ok {
			return h, false
		}
		hashed++
	}
	regionHashedOps.Add(uint64(hashed)) //nolint:gosec // reason: hashed counts stamped ops and is non-negative.
	return h, true
}

// regionPtrsEqual reports whether the Font and Grid identities match per
// index. The content stamp skips those two pointers, so a region skips
// SameOp only when both agree.
func regionPtrsEqual(prev, next []layout.DisplayOp) bool {
	for i := range prev {
		if prev[i].Font != next[i].Font || prev[i].Grid != next[i].Grid {
			return false
		}
	}
	return true
}

// mixOpStamp folds every SameOp input except the Font and Grid identities.
func mixOpStamp(h uint64, o *layout.DisplayOp) (uint64, bool) {
	h = mixStamp(h, uint64(o.Kind))
	h = mixStampStr(h, o.Text)
	h = mixStampBool(h, o.Bold)
	h = mixStampBool(h, o.FakeOblique)
	h = mixStampBool(h, o.IsJPEG)
	h = mixStampBool(h, o.IsBackground)
	h = mixStampBool(h, o.Fixed)
	h = mixStampBool(h, o.Pinned)
	h = mixStampBool(h, o.Positioned)
	h = mixStampBool(h, o.ZIndexSet)
	h = mixStampBool(h, o.XformSet)
	h = mixStampInt(h, o.ZIndex)
	h = mixStampInt(h, o.StickyID)
	h = mixStamp(h, uint64(o.StrokeMask))
	h = mixStamp(h, uint64(o.LineInset))
	h = mixStampStr(h, o.LinkURI())
	h = mixStampStr(h, o.ImageAlt())
	h = mixStampStr(h, o.BlendModeName())
	h = mixStampStr(h, o.FontFeatures())
	h = mixStampStr(h, o.TextLanguage())
	h = mixStampStr(h, o.TextAutospace())
	h = mixStampStr(h, o.TextTransformValue())
	h = mixStampBool(h, o.NoFakeBoldValue())
	h = mixStampBool(h, o.Outline())
	h = mixStamp(h, uint64(o.GroupBoundary()))

	nums := numOf(o)
	var ok bool
	if h, ok = mixStampFloats(h, nums[:]); !ok {
		return h, false
	}

	m := o.Transform()
	mats := [8]float64{m.A, m.B, m.C, m.D, m.E, m.F, o.Opacity(), float64(o.RotateDeg)}
	if h, ok = mixStampFloats(h, mats[:]); !ok {
		return h, false
	}

	ab, aw, ah := o.ImageBytes()
	h = mixStampInt(h, aw)
	h = mixStampInt(h, ah)
	h = mixStampInt(h, len(ab))
	h = mixStampBytes(h, ab)

	return h, true
}

// mixStamp folds one value into an FNV-1a chain.
func mixStamp(h, v uint64) uint64 {
	return (h ^ v) * stampPrime
}

// mixStampInt folds a paint int. Negative values wrap, which stays
// deterministic, and that is all a content stamp needs.
//
//nolint:gosec // reason: int-to-uint64 wrap is deterministic; the stamp only needs equal inputs to hash equal.
func mixStampInt(h uint64, v int) uint64 {
	return (h ^ uint64(v)) * stampPrime
}

// mixStampBool folds one flag.
func mixStampBool(h uint64, v bool) uint64 {
	if v {
		h ^= 1
	}
	return h * stampPrime
}

// mixStampStr folds string bytes without allocating.
func mixStampStr(h uint64, s string) uint64 {
	for i := 0; i < len(s); i++ {
		h = (h ^ uint64(s[i])) * stampPrime
	}
	return h
}

// mixStampBytes folds a payload without retaining it.
func mixStampBytes(h uint64, b []byte) uint64 {
	for i := 0; i < len(b); i++ {
		h = (h ^ uint64(b[i])) * stampPrime
	}
	return h
}

// mixStampFloats folds paint floats, normalizing signed zero the way
// floatBits does so the stamp agrees with ==. A NaN reports false so the
// caller falls back to the per-index scan.
func mixStampFloats(h uint64, vs []float64) (uint64, bool) {
	for _, v := range vs {
		if math.IsNaN(v) {
			return h, false
		}
		h = (h ^ floatBits(v)) * stampPrime
	}
	return h, true
}
