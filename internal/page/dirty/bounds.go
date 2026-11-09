package dirty

import (
	"image"
	"math"
	"sync/atomic"

	"github.com/chinmay-sawant/blinkless/layout"
)

// halfDivisor splits a stroke width into its two-sided growth.
const halfDivisor = 2

// maxDirtyScan caps the ops CountOps inspects, so one huge frame cannot
// make the dirty count cost more than the replay it measures.
const maxDirtyScan = 4096

// OpBounds is one operation's box in CSS pixels. A text or bullet op
// carries its baseline in Y, so the top comes from the face ascent. A
// transformed op bounds where its ink lands, not its layout box.
func OpBounds(op *layout.DisplayOp, d *layout.Display) image.Rectangle {
	if op.XformSet && !op.Transform().IsIdentity() {
		return opBoundsXform(op, d)
	}

	scale := d.PixelPerPoint
	if scale <= 0 {
		scale = 1
	}

	x0, y0 := op.X, op.Y
	x1, y1 := op.X+op.W, op.Y+op.H

	if op.Kind == layout.DisplayOpText || op.Kind == layout.DisplayOpBullet {
		y0 = op.Y - ascentOf(op)
		y1 = op.Y + op.InkDescent
	}

	return image.Rect(
		int(x0*scale)-1, int(y0*scale)-1,
		int(x1*scale)+1, int(y1*scale)+1,
	)
}

// opBoundsXform maps an op's local painted box through its baked 2D
// transform, so a partial repaint covers where the ink actually lands.
func opBoundsXform(op *layout.DisplayOp, d *layout.Display) image.Rectangle {
	scale := d.PixelPerPoint
	if scale <= 0 {
		scale = 1
	}

	x0, y0, x1, y1 := localPaintBox(op)
	m := op.Transform()

	loX, loY := math.Inf(1), math.Inf(1)
	hiX, hiY := math.Inf(-1), math.Inf(-1)

	for _, p := range [4][2]float64{{x0, y0}, {x1, y0}, {x1, y1}, {x0, y1}} {
		x, y := m.Apply(p[0], p[1])
		loX, loY = math.Min(loX, x), math.Min(loY, y)
		hiX, hiY = math.Max(hiX, x), math.Max(hiY, y)
	}

	return image.Rect(
		int(loX*scale)-1, int(loY*scale)-1,
		int(hiX*scale)+1, int(hiY*scale)+1,
	)
}

// localPaintBox is the op's painted box in points before its transform,
// matching the geometry the replay bounds use.
func localPaintBox(op *layout.DisplayOp) (x0, y0, x1, y1 float64) {
	x0, y0 = op.X, op.Y
	x1, y1 = op.X+op.W, op.Y+op.H

	if op.Kind == layout.DisplayOpText || op.Kind == layout.DisplayOpBullet {
		y0 = op.Y - ascentOf(op)
		y1 = op.Y + op.InkDescent
	}

	if op.Kind == layout.DisplayOpStrokeRect || op.Kind == layout.DisplayOpLine {
		grow := math.Max(op.Width, 1) / halfDivisor
		x0, y0, x1, y1 = x0-grow, y0-grow, x1+grow, y1+grow
	}

	return x0, y0, x1, y1
}

// ascentOf returns the face ascent in points for one text op.
func ascentOf(op *layout.DisplayOp) float64 {
	if op.Font == nil || op.Font.UnitsPerEm() <= 0 {
		return op.Size
	}

	return float64(op.Font.Ascent()) / float64(op.Font.UnitsPerEm()) * op.Size
}

// CountOps counts the ops whose bounds meet r, stopping at maxDirtyScan
// so the count stays cheaper than a full replay.
func CountOps(d *layout.Display, r image.Rectangle) int {
	if d == nil {
		return 0
	}

	n := 0

	for i := range d.Ops {
		if i >= maxDirtyScan {
			break
		}

		if OpBounds(&d.Ops[i], d).Intersect(r).Empty() {
			continue
		}

		n++
	}

	return n
}

// BoundsMemo carries the next-side rounded op bounds one display diff
// already computed, so the perf accounting pass reuses them instead of
// recomputing every bound a second time. A zero rectangle means unknown:
// OpBounds never returns an empty rectangle for finite inputs, so an
// empty slot unambiguously falls back to a fresh computation. The memo
// is single-diff scoped: retained operations are patched in place
// between frames, so no entry survives past the diff that filled it.
type BoundsMemo struct {
	rects []image.Rectangle
}

// boundsMemoCache is a single reusable memo slot. One DiffWithBounds takes
// it, one Free returns it. Plain atomics keep reuse allocation-free in
// steady state, the same reason the diff arena avoids sync.Pool. A
// concurrent diff that misses the slot allocates a private memo; the
// result stays correct either way.
//
//nolint:gochecknoglobals // reason: process-wide single reusable bounds memo across frames, not shared mutable state beyond the atomic slot.
var boundsMemoCache atomic.Pointer[BoundsMemo]

// maxMemoRects caps the pooled memo slab, matching the arena cap policy:
// larger frames allocate a private memo instead of pinning the slot.
const maxMemoRects = 65536

// memoReuseMultiple bounds how oversized a reused memo slab may be: a
// cached slab more than this multiple of the request is dropped instead
// of pinning a giant backing array across small steady-state frames.
const memoReuseMultiple = 4

// minMemoRetainCap floors the shrink check: slabs at or under this many
// rects are always reused when they fit, so tiny frames never churn.
const minMemoRetainCap = 4096

// rentBoundsMemo takes a memo for n next-side operations, reusing the
// cached slab when it fits. The slab is cleared, so every slot reads
// unknown until the diff fills it.
func rentBoundsMemo(n int) *BoundsMemo {
	// Oversized frames bypass the cache with a private memo, leaving the
	// hot slab cached: taking it would only fail the fit check and drop
	// it, evicting the slab steady-state frames reuse.
	if n > maxMemoRects {
		return &BoundsMemo{rects: make([]image.Rectangle, n)}
	}
	if m := boundsMemoCache.Swap(nil); m != nil {
		if c := cap(m.rects); c >= n && (c <= minMemoRetainCap || c <= memoReuseMultiple*n) {
			m.rects = m.rects[:n]
			clear(m.rects)
			return m
		}
		// Too small or dwarfing the frame: drop it so neither a stale
		// slab nor a giant one pins the slot. The fresh memo this frame
		// caches on Free sets the new size class.
	}
	return &BoundsMemo{rects: make([]image.Rectangle, n)}
}

// Free returns the memo slab for reuse. A nil memo is a no-op, so full
// diffs that never fill one free unconditionally. The caller must not
// use the memo after Free.
func (m *BoundsMemo) Free() {
	if m == nil {
		return
	}
	if cap(m.rects) > maxMemoRects {
		return
	}
	if old := boundsMemoCache.Swap(m); old != nil {
		// Slot was refilled while this memo was in flight; drop the
		// evicted one so at most one stays cached.
		_ = old
	}
}

// Len reports how many next-side operations the memo covers.
func (m *BoundsMemo) Len() int {
	if m == nil {
		return 0
	}
	return len(m.rects)
}

// Set records the next-side bound of operation i.
func (m *BoundsMemo) Set(i int, r image.Rectangle) {
	if m == nil || i < 0 || i >= len(m.rects) {
		return
	}
	m.rects[i] = r
}

// Get returns the stored bound of operation i. ok is false when the diff
// never computed that index, and the caller falls back to OpBounds.
func (m *BoundsMemo) Get(i int) (r image.Rectangle, ok bool) {
	if m == nil || i < 0 || i >= len(m.rects) {
		return image.Rectangle{}, false
	}
	r = m.rects[i]
	if r.Empty() {
		boundsMemoMiss.Add(1)
		return image.Rectangle{}, false
	}
	boundsMemoHits.Add(1)
	return r, true
}
