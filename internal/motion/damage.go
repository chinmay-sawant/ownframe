package motion

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
	"image"
)

// Damage is the set of rectangles one Step changed, mirroring the page's
// dirty set. It keeps at most MaxRects of them: past the cap it collapses to
// one bounding box and Overflow reports that. A caller reads Rects[:N].
type Damage struct {
	Rects    [dirty.MaxRects]image.Rectangle
	N        int
	Overflow bool

	union image.Rectangle
}

// Add inserts one changed rectangle, merging it with any neighbour within
// the page's dirty gap. The list is fixed size, so a live frame allocates
// nothing.
func (d *Damage) Add(r image.Rectangle) {
	if r.Empty() {
		return
	}

	if d.union.Empty() {
		d.union = r
	} else {
		d.union = d.union.Union(r)
	}

	if d.Overflow {
		d.Rects[0] = d.union

		return
	}

	d.Rects[d.N] = r
	d.N++
	d.N = len(dirty.CoalesceRects(d.Rects[:d.N], dirty.Gap))
	if d.N == len(d.Rects) {
		d.Rects[0] = d.union
		d.N = 1
		d.Overflow = true
	}
}

// Bounds is the one box covering every rectangle added, including ones that
// coalescing merged away.
func (d *Damage) Bounds() image.Rectangle {
	return d.union
}

// Empty reports that the step changed nothing.
func (d *Damage) Empty() bool {
	return d.union.Empty()
}

// stampTransformClips records the ancestor overflow box on every operation a
// transform player writes. Geometry-baked clipping moves with the sampled
// matrix, so the replay intersects transformed paint with this rect instead.
// Ops are display-local, so a rebind re-stamps the new list; entries that
// write no matrix are left alone.
func stampTransformClips(d *layout.Display, entries []entry) {
	if d == nil || len(entries) == 0 {
		return
	}

	ppp := d.PointsPerPixel
	if ppp <= 0 {
		ppp = 1
	}

	for i := range entries {
		e := &entries[i]
		if e.mask&fieldMatrix == 0 || len(e.indices) == 0 {
			continue
		}

		for _, index := range e.indices {
			if index < 0 || index >= len(d.Ops) || index >= len(d.OpOwners) {
				continue
			}

			owner := d.OpOwners[index]
			if owner < 0 || owner >= len(d.Boxes) {
				continue
			}

			box := &d.Boxes[owner]
			if !box.HasOverflowClip {
				continue
			}

			c := box.OverflowClip
			d.Ops[index].SetOverflowClip(c[0]*ppp, c[1]*ppp, c[2]*ppp, c[3]*ppp)
		}
	}
}

// Request is one element whose latest sample cannot be written as a 2D matrix:
// the element id, the canonical CSS transform list to stamp when the sample
// carried one, and the perspective values when it carried those. The page turns
// it into a layout.TransformOverride and hands one pass worth of them to
// layout.RestampTransforms.
//
// Transform is empty for a perspective-only sample; the page sets
// HasTransform from that emptiness rather than replacing the element's static
// transform with none. Perspective is in layout points, zero means none.
// PerspectiveOriginX and PerspectiveOriginY are page coordinates in layout
// points, the vanishing point the engine resolves; HasPerspectiveOrigin reports
// whether the keyframe declared an origin, so an undeclared one keeps the
// element's resolved default.
type Request struct {
	ElementID string
	Transform string

	Perspective          float64
	HasPerspective       bool
	PerspectiveOriginX   float64
	PerspectiveOriginY   float64
	HasPerspectiveOrigin bool
}

// TakeRestamp returns the re-projection requests the last Step collected, one
// per element, and clears the runtime's list. The slice is reused by the next
// Step, so the caller consumes it before stepping again. A 2D-only page gets
// nil.
func (r *Runtime) TakeRestamp() []Request {
	if r == nil || len(r.restamp) == 0 {
		return nil
	}

	requests := r.restamp
	r.restamp = r.restamp[:0]

	return requests
}

// addRestamp records one element's request. Every target of one player shares
// the element, so the first value is the list: a box with five operations asks
// for one restamp, not five. A box without an id cannot be named in a
// layout.TransformOverride, so it is skipped; Stats.NeedsRestamp still counts
// the player, because the sample did need the pass.
func (r *Runtime) addRestamp(e *entry, values []RestampValue) {
	if e.key.id == "" || len(values) == 0 {
		return
	}

	value := values[0]
	request := Request{
		ElementID: e.key.id,
		Transform: value.Transform,
	}
	if value.HasPerspective || value.HasPerspectiveOrigin {
		request.HasPerspective = value.HasPerspective
		request.Perspective = value.Perspective
		request.HasPerspectiveOrigin = value.HasPerspectiveOrigin
		request.PerspectiveOriginX = e.pageX + value.PerspectiveOriginX
		request.PerspectiveOriginY = e.pageY + value.PerspectiveOriginY
	}

	r.restamp = append(r.restamp, request)
}

// NeedsRetainedState reports whether the last bind found motion that needs the
// engine's retained layout state: the 3D and perspective re-projection pass or
// the layout-tier override relayout. The page lays out with retained state from
// this answer: without the retained box tree neither pass can run.
func (r *Runtime) NeedsRetainedState() bool {
	return r != nil && r.stats.NeedsRetainedState
}

// NoteRestamp records the outcome of the page's re-projection pass. applied is
// the number of elements the pass stamped; a failure keeps the display the page
// already draws and is reported through Stats.RestampError.
func (r *Runtime) NoteRestamp(applied int, err error) {
	if r == nil {
		return
	}

	r.stats.Restamped = applied
	if err != nil {
		r.stats.RestampError = err.Error()

		return
	}

	r.stats.RestampError = ""
}
