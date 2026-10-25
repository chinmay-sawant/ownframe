package motion

import (
	"strings"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/animation"
)

// View is the page scroller state a scroll or view timeline resolves against.
// Every value is a CSS pixel. The page owns scrolling, so it reports the
// offset, the viewport, and the painted content before each sample; a runtime
// that never receives a view leaves every timeline inactive.
type View struct {
	ScrollX, ScrollY float64
	Width, Height    float64
	ContentWidth     float64
	ContentHeight    float64
}

// SetView records the page scroller state Step resolves timelines against. It
// is read on the page loop, like Bind and Step.
func (r *Runtime) SetView(v View) {
	if r == nil {
		return
	}
	r.view, r.hasView = v, true
}

// Timeline kinds the engine exports on a box.
const (
	timelineScroll = "scroll"
	timelineView   = "view"
	// timelineRangeCount is the number of named animation-range edges a view
	// timeline resolves: entry, exit, cover, and contain.
	timelineRangeCount = 4
)

// timelineRef is one animation-timeline name resolved against the boxes. A
// scroll reference reads the page scroller on axis; a view reference reads
// the subject's border box against the scrollport adjusted by the declared
// insets.
type timelineRef struct {
	name       string
	kind       string
	axis       string
	insetStart float64
	insetEnd   float64
	subject    timelineBox
}

// timelineBox is the subject geometry a view timeline samples. It is the
// declaring element's border box in CSS pixels, page space.
type timelineBox struct {
	x, y, w, h float64
}

// timelineState is one bound player's timeline input: the names it references,
// the concrete ranges its geometry produces, and the maps a sample reuses.
// The maps are allocated at bind and cleared in place, so a live sample never
// allocates.
type timelineState struct {
	refs   []timelineRef
	values map[string]float64
	ranges map[string]animation.Range

	// viewRanges reports that the named range map comes from view geometry.
	// A scroll timeline leaves every named range at the full progress.
	viewRanges bool
	// axis, insets, and subject describe the view geometry the range map
	// uses. They are the first view timeline the element references.
	axis       string
	insetStart float64
	insetEnd   float64
	subject    timelineBox
	hasSubject bool
}

// bindTimelines resolves the named timelines and range geometry one element
// references. It returns the state and the number of dashed-ident references
// that did not resolve: a missing name, a nested element scroller this engine
// does not implement, or a view subject that is not in the display. An
// unresolved reference makes its animation inactive; there is no time
// fallback.
func bindTimelines(d *Display, box layout.Box, index int) (timelineState, int) {
	var state timelineState
	if d == nil || index < 0 || index >= len(d.Boxes) {
		return state, 0
	}

	timelineUsed := false
	inlineView := false
	var seen map[string]bool
	unresolved := 0
	for i := range box.Animations {
		name := strings.TrimSpace(box.Animations[i].Timeline)
		switch {
		case name == "" || name == "auto":
			continue
		case name == "none":
			timelineUsed = true

			continue
		case strings.HasPrefix(name, "scroll("):
			timelineUsed = true

			continue
		case strings.HasPrefix(name, "view("):
			timelineUsed, inlineView = true, true

			continue
		}
		timelineUsed = true
		if seen == nil {
			seen = make(map[string]bool, len(box.Animations))
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		ref, ok := resolveTimeline(d, index, name)
		if !ok {
			unresolved++

			continue
		}
		state.refs = append(state.refs, ref)
		if ref.kind == timelineView && !state.viewRanges {
			state.viewRanges = true
			state.axis = ref.axis
			state.insetStart, state.insetEnd = ref.insetStart, ref.insetEnd
			state.subject, state.hasSubject = ref.subject, true
		}
	}

	if !timelineUsed {
		return timelineState{}, 0
	}

	if len(state.refs) > 0 {
		state.values = make(map[string]float64, len(state.refs))
	}
	state.ranges = make(map[string]animation.Range, timelineRangeCount)
	state.axis = timelineAxis(state.axis)
	if inlineView && !state.viewRanges {
		state.viewRanges = true
		state.axis = "y"
		state.subject = timelineBox{box.X, box.Y, box.W, box.H}
		state.hasSubject = true
	}

	return state, unresolved
}

// resolveTimeline finds the nearest declaration of name on the element or one
// of its ancestors. A nearer declaration wins even when it cannot resolve, so
// a shadowing name never falls through to a farther box.
func resolveTimeline(d *Display, from int, name string) (timelineRef, bool) {
	index := from
	for depth := 0; index >= 0 && index < len(d.Boxes) && depth <= len(d.Boxes); depth++ {
		box := &d.Boxes[index]
		for i := range box.Timelines {
			def := &box.Timelines[i]
			if def.Name != name {
				continue
			}
			// A named element scroll container is not implemented; the page
			// scroller is the only one this engine models. Do not pretend
			// otherwise.
			if def.ScrollerID != "" {
				return timelineRef{}, false
			}
			switch def.Kind {
			case timelineScroll:
				return timelineRef{name: name, kind: timelineScroll, axis: timelineAxis(def.Axis)}, true
			case timelineView:
				subject := index
				if def.SubjectID != "" {
					found := boxIndexByID(d.Boxes, def.SubjectID)
					if found < 0 {
						return timelineRef{}, false
					}
					subject = found
				}
				b := d.Boxes[subject]

				return timelineRef{
					name:       name,
					kind:       timelineView,
					axis:       timelineAxis(def.Axis),
					insetStart: def.InsetStart,
					insetEnd:   def.InsetEnd,
					subject:    timelineBox{b.X, b.Y, b.W, b.H},
				}, true
			default:
				return timelineRef{}, false
			}
		}
		index = box.ParentIndex
	}

	return timelineRef{}, false
}

// boxIndexByID returns the first box index with id, or -1.
func boxIndexByID(boxes []layout.Box, id string) int {
	for i := range boxes {
		if boxes[i].ID == id {
			return i
		}
	}

	return -1
}

// timelineAxis normalizes a declared axis to the page's physical axes. The
// engine already resolves block and inline through the writing mode; anything
// else reads the block axis.
func timelineAxis(axis string) string {
	if axis == "x" {
		return "x"
	}

	return "y"
}

// input builds the single FrameInput a sample uses. It fills the scroll and
// view progress, the element's named timelines, the range values, and the
// subject progress. The maps it points at belong to the entry and are reused,
// so a live sample allocates nothing. The returned count is the number of
// resolved named timelines that had no range at this geometry.
func (r *Runtime) input(e *entry, now time.Time) (animation.FrameInput, int) {
	in := animation.FrameInput{Now: now}
	if !r.hasView {
		return in, 0
	}

	box := timelineBox{e.boxX, e.boxY, e.boxW, e.boxH}
	in.ScrollProgress, _ = r.scrollProgress("y")
	in.ViewProgress, _ = r.viewProgress("y", box, 0, 0)

	return in, e.timelines.fill(r, box, &in)
}

// scrollProgress is the page scroller's progress on axis: 0 at the top, 1 at
// the end. A scroller with no range has no progress and is not a timeline.
func (r *Runtime) scrollProgress(axis string) (float64, bool) {
	if !r.hasView {
		return 0, false
	}
	if axis == "x" {
		return progressOrNone(r.view.ScrollX, r.view.ContentWidth-r.view.Width)
	}

	return progressOrNone(r.view.ScrollY, r.view.ContentHeight-r.view.Height)
}

func progressOrNone(offset, span float64) (float64, bool) {
	if span <= 0 {
		return 0, false
	}

	return clamp01(offset / span), true
}

// viewProgress maps a subject border box to its visibility progress in the
// scrollport: 0 when the leading edge reaches the scrollport's trailing edge,
// 1 when the trailing edge passes its leading edge. Insets shrink the
// scrollport on each edge, the way view-timeline-inset does.
func (r *Runtime) viewProgress(axis string, box timelineBox, insetStart, insetEnd float64) (float64, bool) {
	start, end, ok := r.viewportBounds(axis, insetStart, insetEnd)
	if !ok {
		return 0, false
	}
	pos, size := box.y, box.h
	if axis == "x" {
		pos, size = box.x, box.w
	}
	length := end - start
	span := length + size
	if length <= 0 || span <= 0 {
		return 0, false
	}

	return clamp01((end - pos) / span), true
}

// viewportBounds is the scrollport in page space on axis, adjusted by the view
// timeline insets in CSS pixels.
func (r *Runtime) viewportBounds(axis string, insetStart, insetEnd float64) (float64, float64, bool) {
	if !r.hasView {
		return 0, 0, false
	}
	if axis == "x" {
		if r.view.Width <= 0 {
			return 0, 0, false
		}

		return r.view.ScrollX + insetStart, r.view.ScrollX + r.view.Width - insetEnd, true
	}
	if r.view.Height <= 0 {
		return 0, 0, false
	}

	return r.view.ScrollY + insetStart, r.view.ScrollY + r.view.Height - insetEnd, true
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}

	return v
}

// fill writes this element's timeline progress and range values into in. The
// maps are cleared in place, so a name that stops resolving between samples
// leaves the animation inactive instead of holding a stale progress value.
func (t *timelineState) fill(r *Runtime, box timelineBox, in *animation.FrameInput) int {
	inactive := 0
	if t.values != nil {
		clear(t.values)
		for i := range t.refs {
			ref := &t.refs[i]
			progress, ok := r.referenceProgress(ref)
			if !ok {
				inactive++

				continue
			}
			t.values[ref.name] = progress
		}
		in.Timelines = t.values
	}

	if t.ranges != nil {
		clear(t.ranges)
		in.Ranges = t.ranges
		if t.viewRanges {
			t.fillViewRanges(r)
		} else {
			t.ranges["entry"] = animation.Range{Start: 0, End: 1}
			t.ranges["exit"] = animation.Range{Start: 0, End: 1}
			t.ranges["cover"] = animation.Range{Start: 0, End: 1}
			t.ranges["contain"] = animation.Range{Start: 0, End: 1}
		}
	}

	if t.hasSubject {
		in.SourceProgress, _ = r.viewProgress(t.axis, t.subject, t.insetStart, t.insetEnd)
		in.TargetProgress, _ = r.viewProgress(t.axis, box, 0, 0)
	}

	return inactive
}

// fillViewRanges resolves the four named view ranges from the subject box and
// the adjusted scrollport. A contain range is reversed when the subject is
// taller than the scrollport, and the animation engine treats a reversed range
// as inactive, which is the spec's empty range.
func (t *timelineState) fillViewRanges(r *Runtime) {
	start, end, ok := r.viewportBounds(t.axis, t.insetStart, t.insetEnd)
	if !ok {
		return
	}
	size := t.subject.h
	if t.axis == "x" {
		size = t.subject.w
	}
	length := end - start
	span := length + size
	if span <= 0 {
		return
	}
	entryEnd := clamp01(size / span)
	exitStart := clamp01(length / span)
	t.ranges["entry"] = animation.Range{Start: 0, End: entryEnd}
	t.ranges["exit"] = animation.Range{Start: exitStart, End: 1}
	t.ranges["cover"] = animation.Range{Start: 0, End: 1}
	t.ranges["contain"] = animation.Range{Start: entryEnd, End: exitStart}
}

// referenceProgress reads one resolved named timeline at the current page
// state.
func (r *Runtime) referenceProgress(ref *timelineRef) (float64, bool) {
	if ref.kind == timelineView {
		return r.viewProgress(ref.axis, ref.subject, ref.insetStart, ref.insetEnd)
	}

	return r.scrollProgress(ref.axis)
}

// hasTimeEffect reports whether the element has an animation that follows the
// wall clock. A player made only of scroll, view, or named timeline effects
// never asks for a frame on its own: a still scroll position is a static
// sample. A triggered time effect still counts, because a click starts it.
func hasTimeEffect(items []layout.Animation) bool {
	for i := range items {
		item := &items[i]
		if item.Name == "" || item.Name == "none" {
			continue
		}
		if item.Timeline == "" || item.Timeline == "auto" {
			return true
		}
	}

	return false
}
