package dirty

import (
	"math"
	"reflect"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// This file proves the diff fast paths without timing: per-index SameOp
// calls are counted, so a one-operation change on a large list must stay
// near one region, identical lists must call SameOp zero times, and every
// generated mismatch must match the pre-edit algorithm exactly.
//
// Image payloads cannot be built here: the extra payload setters are
// unexported outside the layout module, so image ops carry nil bytes and
// the shared-backing shortcut is covered by code review plus the
// sameImageIdentity counter staying zero on imageless workloads. Fonts
// are nil for the same reason; grids are attached through reflection
// (mirroring internal/replay/image_test.go), so they cover the
// pointer-identity fallback.

// diffGen is a deterministic xorshift64* source so golden cases replay.
type diffGen struct{ s uint64 }

func (g *diffGen) next() uint64 {
	g.s ^= g.s >> 12
	g.s ^= g.s << 25
	g.s ^= g.s >> 27
	return g.s * 2685821657736338717
}

func (g *diffGen) intn(n int) int {
	if n <= 0 {
		panic("dirty: diffGen.intn bound must be positive")
	}
	return int(g.next() % uint64(n)) //nolint:gosec // reason: the remainder is below the positive int bound n, so it always fits int.
}

// setGrid attaches a GridRun payload to op. The Grid type is unexported
// outside the layout module, so the test builds it by shape and writes
// the field the way internal/replay/image_test.go does.
func setGrid(t *testing.T, op *layout.DisplayOp, segs [][8]float64) {
	t.Helper()

	gridType, ok := reflect.TypeOf(layout.DisplayOp{}).FieldByName("Grid")
	if !ok {
		t.Fatal("DisplayOp has no Grid field")
	}
	run := reflect.New(gridType.Type.Elem())
	segsField := run.Elem().FieldByName("Segs")
	if !segsField.IsValid() {
		t.Fatal("GridRun has no Segs field")
	}
	slice := reflect.MakeSlice(segsField.Type(), len(segs), len(segs))
	names := []string{"X", "Y", "W", "H", "Width", "R", "G", "B"}
	for i, s := range segs {
		seg := slice.Index(i)
		for j, name := range names {
			f := seg.FieldByName(name)
			if !f.IsValid() || !f.CanSet() {
				t.Fatalf("GridSeg has no settable %s field", name)
			}
			f.SetFloat(s[j])
		}
	}
	segsField.Set(slice)

	opv := reflect.ValueOf(op).Elem()
	gf := opv.FieldByName("Grid")
	if !gf.IsValid() || !gf.CanSet() {
		t.Fatal("DisplayOp has no settable Grid field")
	}
	gf.Set(run)
}

// genDiffOp builds one random op covering every SameOp input except Font
// and image bytes, which external packages cannot construct.
func genDiffOp(t *testing.T, g *diffGen) layout.DisplayOp {
	t.Helper()

	words := []string{"hello", "world", "ownframe", "diff", "stamp", "region", "frame", "paint", "", "a"}
	kinds := []layout.DisplayKind{
		layout.DisplayOpFillRect,
		layout.DisplayOpStrokeRect,
		layout.DisplayOpLine,
		layout.DisplayOpText,
		layout.DisplayOpBullet,
		layout.DisplayOpGridRun,
		layout.DisplayOpLinkURI,
		layout.DisplayOpImage,
	}

	op := layout.DisplayOp{
		Kind:  kinds[g.intn(len(kinds))],
		X:     float64(g.intn(78) * 10),
		Y:     float64(g.intn(58) * 10),
		W:     float64(8 + g.intn(5)*8),
		H:     float64(8 + g.intn(5)*8),
		R:     float64(g.intn(4)) / 3,
		G:     float64(g.intn(4)) / 3,
		B:     float64(g.intn(4)) / 3,
		Alpha: 1,
		Width: float64(1 + g.intn(3)),
		Size:  float64(10 + g.intn(7)),
	}
	if g.intn(4) == 0 {
		op.Bold = true
	}
	if g.intn(6) == 0 {
		op.FakeOblique = true
	}
	if g.intn(8) == 0 {
		op.IsJPEG = true
	}
	if g.intn(8) == 0 {
		op.Fixed = true
	}
	if g.intn(8) == 0 {
		op.Pinned = true
	}
	if g.intn(8) == 0 {
		op.Positioned = true
	}
	if g.intn(5) == 0 {
		op.ZIndexSet = true
	}
	op.ZIndex = g.intn(4) - 1
	if g.intn(6) == 0 {
		op.StickyID = 1 + g.intn(2)
	}
	if g.intn(8) == 0 {
		op.RotateDeg = 45
	}
	if g.intn(8) == 0 {
		op.StrokeMask = 1
	}
	if g.intn(8) == 0 {
		op.LineInset = 1
	}
	if g.intn(4) == 0 {
		op.Radius = 4
		op.RadiusTopLeft = 4
	}
	if op.Kind == layout.DisplayOpText && g.intn(3) == 0 {
		op.LetterSpacing = 1
		op.InkDescent = 2
	}
	if op.Kind == layout.DisplayOpText || op.Kind == layout.DisplayOpBullet ||
		op.Kind == layout.DisplayOpLinkURI {
		op.Text = words[g.intn(len(words))]
	}
	if op.Kind == layout.DisplayOpGridRun {
		setGrid(t, &op, [][8]float64{{op.X, op.Y, op.W, op.H, op.Width, op.R, op.G, op.B}})
	}
	switch g.intn(8) {
	case 0:
		op.SetURI("https://example.com/" + words[g.intn(len(words))])
	case 1:
		op.SetBlendMode("multiply")
	case 2:
		op.SetXform(layout.Matrix2D{A: 1, D: 1, E: float64(g.intn(20)), F: float64(g.intn(20))})
	case 3:
		op.SetPaintOpacity(0.5)
	case 4:
		op.SetTextLanguage("en")
		op.SetFontFeatures("liga")
		op.SetTextTransform("uppercase")
	}
	return op
}

func genDisplay(t *testing.T, seed uint64, n int) *layout.Display {
	t.Helper()
	g := &diffGen{s: seed | 1}
	ops := make([]layout.DisplayOp, n)
	for i := range ops {
		ops[i] = genDiffOp(t, g)
	}
	return &layout.Display{Width: 800, Height: 600, Ops: ops}
}

// seedFor derives a deterministic golden seed from a non-negative size.
func seedFor(n int, mult uint64) uint64 {
	if n < 0 {
		panic("dirty: negative display size")
	}
	return uint64(n+1) * mult
}

// shallowCopy shares Grid pointers, like consecutive frames that reuse the
// same backing runs.
func shallowCopy(d *layout.Display) *layout.Display {
	return &layout.Display{Width: d.Width, Height: d.Height, Ops: append([]layout.DisplayOp(nil), d.Ops...)}
}

func resetDiffCounters() {
	sameOpCalls.Store(0)
	diffSkippedOps.Store(0)
	sameImageIdentity.Store(0)
}

func TestDiffRegionStamp_SingleChangeCost(t *testing.T) {
	const n = 4096
	prev := genDisplay(t, 7, n)
	next := shallowCopy(prev)
	mid := n / 2
	next.Ops[mid].R = 1 - next.Ops[mid].R

	resetDiffCounters()
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

	// One touched region re-compares at most its own ops; every other
	// region must skip SameOp instead of paying the full list.
	base := mid / diffRegionSize * diffRegionSize
	end := base + diffRegionSize
	if end > n {
		end = n
	}
	wantCalls := uint64(end - base)
	if got := sameOpCalls.Load(); got != wantCalls {
		t.Fatalf("expected %d SameOp calls for one touched region, got %d", wantCalls, got)
	}
	if got := diffSkippedOps.Load(); got != uint64(n)-wantCalls {
		t.Fatalf("expected %d skipped ops, got %d", uint64(n)-wantCalls, got)
	}
	if bound := uint64(n / 8); wantCalls >= bound {
		t.Fatalf("one change cost %d SameOp calls, want under %d (N=%d)", wantCalls, bound, n)
	}
}

func TestDiffRegionStamp_IdenticalShortCircuit(t *testing.T) {
	const n = 1000
	prev := genDisplay(t, 11, n)

	resetDiffCounters()
	set, full, changed := Diff(prev, shallowCopy(prev))
	if full || changed != 0 || set.N != 0 {
		t.Fatalf("identical lists must diff empty, got full=%v changed=%d rects=%d", full, changed, set.N)
	}
	if got := sameOpCalls.Load(); got != 0 {
		t.Fatalf("identical lists must skip SameOp, got %d calls", got)
	}
	if got := diffSkippedOps.Load(); got != n {
		t.Fatalf("identical lists must skip all %d ops, skipped %d", n, got)
	}

	// Same backing store short-circuits without even stamping.
	resetDiffCounters()
	set, full, changed = Diff(prev, prev)
	if full || changed != 0 || set.N != 0 {
		t.Fatalf("same display must diff empty, got full=%v changed=%d rects=%d", full, changed, set.N)
	}
	if got := sameOpCalls.Load(); got != 0 {
		t.Fatalf("same display must skip SameOp, got %d calls", got)
	}

	// Empty lists stay empty.
	resetDiffCounters()
	empty := &layout.Display{Width: 800, Height: 600}
	set, full, changed = Diff(empty, &layout.Display{Width: 800, Height: 600})
	if full || changed != 0 || set.N != 0 {
		t.Fatalf("empty lists must diff empty, got full=%v changed=%d rects=%d", full, changed, set.N)
	}
	if got := sameOpCalls.Load(); got != 0 {
		t.Fatalf("empty lists must skip SameOp, got %d calls", got)
	}
}

func TestDiffRegionStamp_GridPointerFallback(t *testing.T) {
	// Distinct Grid pointers with equal content: stamps match, the
	// pointer check refuses the skip, and the per-index scan decides.
	prev := genDisplay(t, 13, 5)
	for i := range prev.Ops {
		prev.Ops[i].Kind = layout.DisplayOpGridRun
		setGrid(t, &prev.Ops[i], [][8]float64{{1, 2, 3, 4, 1, 0, 0, 0}})
	}
	next := shallowCopy(prev)
	for i := range next.Ops {
		setGrid(t, &next.Ops[i], [][8]float64{{1, 2, 3, 4, 1, 0, 0, 0}})
	}

	resetDiffCounters()
	set, full, changed := Diff(prev, next)
	calls := sameOpCalls.Load()
	skipped := diffSkippedOps.Load()
	legacySet, legacyFull, legacyChanged := legacyDiff(prev, next)

	if changed != 5 || full != legacyFull || changed != legacyChanged || set.N != legacySet.N {
		t.Fatalf("grid-pointer fallback diverged: got full=%v changed=%d rects=%d, legacy full=%v changed=%d rects=%d",
			full, changed, set.N, legacyFull, legacyChanged, legacySet.N)
	}
	if calls != 5 {
		t.Fatalf("pointer mismatch must fall back to all 5 SameOp calls, got %d", calls)
	}
	if skipped != 0 {
		t.Fatalf("pointer mismatch must skip nothing, skipped %d", skipped)
	}
}

func TestSameOp_NilImageEquality(t *testing.T) {
	a := layout.DisplayOp{Kind: layout.DisplayOpImage, X: 10, Y: 10, W: 20, H: 20}
	b := layout.DisplayOp{Kind: layout.DisplayOpImage, X: 10, Y: 10, W: 20, H: 20}
	c := layout.DisplayOp{Kind: layout.DisplayOpText, X: 10, Y: 10, W: 20, H: 20, Text: "x"}

	resetDiffCounters()
	if !SameOp(&a, &b) {
		t.Fatalf("imageless image ops must compare equal")
	}
	if SameOp(&a, &c) {
		t.Fatalf("image op must differ from text op")
	}
	if got := sameImageIdentity.Load(); got != 0 {
		t.Fatalf("imageless workload must not hit the identity path, got %d", got)
	}
}

func TestArenaEpochReuse_NoAllocs(t *testing.T) {
	prevA := genDisplay(t, 17, 200)
	nextA := shallowCopy(prevA)
	nextA.Ops = append(nextA.Ops, layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 1, Y: 1, W: 8, H: 8, Alpha: 1})
	prevB := genDisplay(t, 19, 300)
	nextB := shallowCopy(prevB)
	nextB.Ops = append(nextB.Ops[:150], nextB.Ops[151:]...)

	// Warm both sizes so pooled slabs fit, then alternate: reuse must
	// serve both without allocating.
	Diff(prevA, nextA)
	Diff(prevB, nextB)

	allocs := testing.AllocsPerRun(20, func() {
		Diff(prevA, nextA)
		Diff(prevB, nextB)
	})
	if allocs != 0 {
		t.Fatalf("expected 0 allocs per run from epoch reuse, got %f", allocs)
	}
}

func TestDiffRegionStamp_GoldenMatchesLegacy(t *testing.T) {
	sizes := []int{0, 1, 2, 63, 64, 65, 127, 128, 200, 1000}
	for _, n := range sizes {
		t.Run("identical", func(t *testing.T) {
			prev := genDisplay(t, seedFor(n, 101), n)
			checkGolden(t, prev, shallowCopy(prev))
		})
		t.Run("single_first", func(t *testing.T) {
			prev, next := mutateOne(t, n, seedFor(n, 103), 0)
			checkGolden(t, prev, next)
		})
		t.Run("single_mid", func(t *testing.T) {
			prev, next := mutateOne(t, n, seedFor(n, 107), n/2)
			checkGolden(t, prev, next)
		})
		t.Run("single_last", func(t *testing.T) {
			prev, next := mutateOne(t, n, seedFor(n, 109), n-1)
			checkGolden(t, prev, next)
		})
		t.Run("scattered", func(t *testing.T) {
			prev := genDisplay(t, seedFor(n, 113), n)
			next := shallowCopy(prev)
			g := &diffGen{s: seedFor(n, 127) | 1}
			for k := 0; k < 5 && k < n; k++ {
				i := g.intn(n)
				next.Ops[i].R = 1 - next.Ops[i].R
			}
			checkGolden(t, prev, next)
		})
		t.Run("kind_flip", func(t *testing.T) {
			prev, next := mutateOne(t, n, seedFor(n, 131), n/3)
			if n > 0 {
				next.Ops[(n/3)%n].Kind = layout.DisplayOpText
				next.Ops[(n/3)%n].Text = "flipped"
			}
			checkGolden(t, prev, next)
		})
		t.Run("geometry_shift", func(t *testing.T) {
			prev, next := mutateOne(t, n, seedFor(n, 137), n/4)
			if n > 0 {
				next.Ops[(n/4)%n].X += 30
				next.Ops[(n/4)%n].W += 6
			}
			checkGolden(t, prev, next)
		})
		t.Run("payload", func(t *testing.T) {
			prev, next := mutateOne(t, n, seedFor(n, 139), n/5)
			if n > 0 {
				i := (n / 5) % n
				next.Ops[i].SetURI("https://example.com/changed")
				next.Ops[i].SetBlendMode("screen")
				next.Ops[i].SetPaintOpacity(0.25)
				next.Ops[i].SetXform(layout.Matrix2D{A: 2, D: 2})
			}
			checkGolden(t, prev, next)
		})
		t.Run("grid_content", func(t *testing.T) {
			prev, next := mutateOne(t, n, seedFor(n, 149), n/6)
			if n > 0 {
				i := (n / 6) % n
				next.Ops[i].Kind = layout.DisplayOpGridRun
				setGrid(t, &next.Ops[i], [][8]float64{{7, 7, 9, 9, 1, 0, 0, 0}})
			}
			checkGolden(t, prev, next)
		})
		t.Run("swap", func(t *testing.T) {
			prev := genDisplay(t, seedFor(n, 151), n)
			next := shallowCopy(prev)
			if n > 1 {
				next.Ops[0], next.Ops[n-1] = next.Ops[n-1], next.Ops[0]
			}
			checkGolden(t, prev, next)
		})
		t.Run("grow", func(t *testing.T) {
			prev := genDisplay(t, seedFor(n, 157), n)
			next := shallowCopy(prev)
			g := &diffGen{s: seedFor(n, 163) | 1}
			for k := 0; k < 3; k++ {
				next.Ops = append(next.Ops, genDiffOp(t, g))
			}
			checkGolden(t, prev, next)
		})
		t.Run("shrink", func(t *testing.T) {
			prev := genDisplay(t, seedFor(n, 167), n)
			next := shallowCopy(prev)
			if n > 3 {
				next.Ops = next.Ops[:n-3]
			} else {
				next.Ops = nil
			}
			checkGolden(t, prev, next)
		})
		t.Run("size_mismatch", func(t *testing.T) {
			prev := genDisplay(t, seedFor(n, 173), n)
			next := shallowCopy(prev)
			next.Width++
			checkGolden(t, prev, next)
		})
		t.Run("nan_geometry", func(t *testing.T) {
			prev := genDisplay(t, seedFor(n, 179), n)
			next := shallowCopy(prev)
			if n > 0 {
				// NaN forces the per-index fallback; results must still match.
				next.Ops[n/2].X = math.NaN()
			}
			checkGolden(t, prev, next)
		})
	}
}

// mutateOne returns a display pair differing in exactly the op at idx,
// or an identical pair when n is 0.
func mutateOne(t *testing.T, n int, seed uint64, idx int) (*layout.Display, *layout.Display) {
	t.Helper()
	prev := genDisplay(t, seed, n)
	next := shallowCopy(prev)
	if n > 0 {
		i := idx % n
		if i < 0 {
			i = 0
		}
		next.Ops[i].G = 1 - next.Ops[i].G
	}
	return prev, next
}

// checkGolden asserts the new diff matches the pre-edit algorithm on
// every observable: full flag, changed count, added/removed tallies,
// rect list, and bounding box.
func checkGolden(t *testing.T, prev, next *layout.Display) {
	t.Helper()
	set, full, changed := Diff(prev, next)
	legacySet, legacyFull, legacyChanged := legacyDiff(prev, next)

	if full != legacyFull || changed != legacyChanged ||
		set.Added != legacySet.Added || set.Removed != legacySet.Removed ||
		set.N != legacySet.N || set.Bounds() != legacySet.Bounds() {
		t.Fatalf("golden diverged: got full=%v changed=%d added=%d removed=%d rects=%d bounds=%v, legacy full=%v changed=%d added=%d removed=%d rects=%d bounds=%v",
			full, changed, set.Added, set.Removed, set.N, set.Bounds(),
			legacyFull, legacyChanged, legacySet.Added, legacySet.Removed, legacySet.N, legacySet.Bounds())
	}
	for i := 0; i < set.N; i++ {
		if set.Rects[i] != legacySet.Rects[i] {
			t.Fatalf("rect %d diverged: got %v, legacy %v", i, set.Rects[i], legacySet.Rects[i])
		}
	}
}

// legacyDiff is the pre-edit equal-length scan: every index compares,
// with no stamps or early-outs.
func legacyDiff(prev, next *layout.Display) (Set, bool, int) {
	if prev.Width != next.Width || prev.Height != next.Height {
		return Set{}, true, len(next.Ops)
	}

	if len(prev.Ops) != len(next.Ops) {
		return legacyDiffByPosition(prev, next)
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

// legacyDiffByPosition is the pre-edit mismatch path: fresh chains per
// diff, an O(N) bool clear, and keys rebuilt per chain visit.
func legacyDiffByPosition(prev, next *layout.Display) (Set, bool, int) {
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
			d.add(OpBounds(&nextSlice[j], next))
			if d.full {
				return d, true, changed
			}
		}
		return d, false, changed
	}

	numOps := len(nextSlice)
	numBuckets := nextPowerOf2(numOps * diffBucketGrowth)
	mask := uint64(numBuckets - 1) //nolint:gosec // reason: numBuckets is a nextPowerOf2 result always >= 16, so numBuckets-1 is non-negative and fits uint64.
	head := make([]int, numBuckets)
	for i := range head {
		head[i] = -1
	}
	chain := make([]int, numOps)
	seen := make([]bool, numOps)

	for i := numOps - 1; i >= 0; i-- {
		k := keyOf(&nextSlice[i])
		b := hashKey(k) & mask
		chain[i] = head[b]
		head[b] = i
	}

	for i := range prevSlice {
		k := keyOf(&prevSlice[i])
		b := hashKey(k) & mask

		found := -1
		prevIdx := -1
		for curr := head[b]; curr != -1; {
			if keyOf(&nextSlice[curr]) == k {
				found = curr
				if prevIdx == -1 {
					head[b] = chain[curr]
				} else {
					chain[prevIdx] = chain[curr]
				}
				break
			}
			prevIdx = curr
			curr = chain[curr]
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
		seen[j] = true

		if SameOp(&prevSlice[i], &nextSlice[j]) {
			continue
		}

		changed++
		d.add(OpBounds(&prevSlice[i], prev).Union(OpBounds(&nextSlice[j], next)))
		if d.full {
			return d, true, changed
		}
	}

	for j := range nextSlice {
		if seen[j] {
			continue
		}

		changed++
		d.Added++
		d.add(OpBounds(&nextSlice[j], next))
		if d.full {
			return d, true, changed
		}
	}

	return d, false, changed
}
