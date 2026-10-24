package motion

import "github.com/chinmay-sawant/ownframe/internal/animation"

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
func (r *Runtime) addRestamp(e *entry, values []animation.RestampValue) {
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
// engine's re-projection pass. The page lays out with retained state from this
// answer: without the retained box tree the pass cannot run.
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
