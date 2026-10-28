package patch

import (
	"image"
	"sync"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
)

// opCap is the pooled buffer capacity for one display update.
const opCap = 64

// OpBuffer is a pooled holder for a display operation slice.
type OpBuffer struct {
	Ops []layout.DisplayOp
}

// Damage is one display comparison. N is the count of partial rects.
// A full report ignores the rects; the page marks the whole frame.
type Damage struct {
	Patch DisplayPatch
	Rects [dirty.MaxRects]image.Rectangle
	N     int
}

// EmptyOps is the buffer the page pool hands out before the first rent.
func EmptyOps() []layout.DisplayOp {
	return make([]layout.DisplayOp, 0, opCap)
}

// NewOpBuffer returns a new OpBuffer initialized with capacity opCap.
func NewOpBuffer() *OpBuffer {
	return &OpBuffer{Ops: EmptyOps()}
}

// Rent takes an operation buffer of at least n slots from pool.
func Rent(pool *sync.Pool, n int) *OpBuffer {
	buf, ok := pool.Get().(*OpBuffer)
	if !ok || buf == nil {
		return &OpBuffer{Ops: make([]layout.DisplayOp, 0, n)}
	}

	if cap(buf.Ops) < n {
		buf.Ops = make([]layout.DisplayOp, 0, n)
	}
	buf.Ops = buf.Ops[:0]

	return buf
}

// Free returns buf to pool. The caller must not use buf after Free.
func Free(pool *sync.Pool, buf *OpBuffer) {
	if buf == nil {
		return
	}
	buf.Ops = buf.Ops[:0]
	pool.Put(buf)
}

// Snapshot copies d's operations into a rented buffer.
// The returned display shares every slice except Ops, which is the copy.
func Snapshot(pool *sync.Pool, d *layout.Display) (layout.Display, *OpBuffer) {
	buf := Rent(pool, len(d.Ops))
	buf.Ops = append(buf.Ops, d.Ops...)
	before := *d
	before.Ops = buf.Ops

	return before, buf
}

// Compare reports how next differs from before.
// A full difference uses one region. Partial rects are Damage.Rects[:N].
func Compare(before, next *layout.Display) Damage {
	set, full, changed := dirty.Diff(before, next)
	regions := set.N
	if full {
		regions = 1
	}

	return Damage{
		Patch: DisplayPatch{
			Changed: changed,
			Added:   set.Added,
			Removed: set.Removed,
			Regions: regions,
			Full:    full,
		},
		Rects: set.Rects,
		N:     set.N,
	}
}

// Update snapshots d, runs fn on it, and compares the two lists.
// The caller frees the buffer after it has recorded the damage.
func Update(pool *sync.Pool, d *layout.Display, fn func(*layout.Display)) (Damage, *OpBuffer) {
	before, buf := Snapshot(pool, d)
	fn(d)

	return Compare(&before, d), buf
}

// OpDamage applies fn to operation i and returns the union of the old and
// new bounds. ok is false when d or fn is nil, or i is out of range.
func OpDamage(d *layout.Display, i int, fn func(*layout.DisplayOp)) (image.Rectangle, bool) {
	if d == nil || fn == nil || i < 0 || i >= len(d.Ops) {
		return image.Rectangle{}, false
	}

	old := dirty.OpBounds(&d.Ops[i], d)
	fn(&d.Ops[i])
	nw := dirty.OpBounds(&d.Ops[i], d)

	return old.Union(nw), true
}
