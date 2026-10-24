package motion

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssanimation"
	"github.com/chinmay-sawant/ownframe/internal/transition"
)

// geometryEpsilon is the change in a CSS length at which a geometry transition
// is worth another engine relayout. A sub-pixel sample keeps the previous
// request, so a slow run does not spend the per-frame budget on a value the
// display cannot show.
const geometryEpsilon = 1e-3

// geometryCapture is one layout length at the two commits. The value is the
// element's border-box extent for width and height, and the inset from the
// parent box origin for top and left.
type geometryCapture struct {
	from, to float64
}

// geometryEntry is one live layout-length transition. It shares the relayout
// budget and element queue with the layout keyframe driver, so a run reaches
// the engine's override path at most once per frame.
type geometryEntry struct {
	key    transitionKey
	from   float64
	to     float64
	record layout.Transition
	start  time.Time
	// animated is the layout-tier properties the box's animations declare at
	// the last bind or adoption. A property in this set wins the cascade, so
	// the run is dropped without reparsing the keyframes every frame.
	animated map[string]struct{}
	// reversingStart and reversingFactor are the CSS reversal state, kept for
	// an interrupt that returns to where the run started.
	reversingStart  float64
	reversingFactor float64
	// applied is the value the last emitted request used, so a run that does
	// not move this frame emits nothing.
	applied    float64
	hasApplied bool
	// sampled reports that the run has reached the engine at least once.
	sampled bool
	// finished marks the entry whose final destination has been emitted. It
	// stays until the next layout pass can prune it, so a failed apply can be
	// retried by ReturnLayout.
	finished bool
}

// transitionGeometryProperty reports whether a property is a layout length the
// display exposes through a box. The display carries the border box, not the
// used style, so width and height are exact for a box with no padding or
// border and the inset is measured from the parent border-box origin.
func transitionGeometryProperty(property string) bool {
	switch property {
	case "width", "height", "top", "left":
		return true
	}

	return false
}

// geometryAnimationNames returns the layout-tier keyframe properties that write
// the same used value a geometry transition would. An animation on any of them
// wins the cascade, so the transition never binds.
func geometryAnimationNames(property string) []string {
	switch property {
	case "width":
		return []string{"width", "inline-size"}
	case "height":
		return []string{"height", "block-size"}
	case "top":
		return []string{"top", "inset", "inset-block-start"}
	case "left":
		return []string{"left", "inset", "inset-inline-start"}
	}

	return nil
}

// layoutAnimatedProperties collects the layout-tier properties the box's
// animations declare. The keyframe parser is the same one the layout binder
// uses, so a property it accepts is the property the player will sample.
func layoutAnimatedProperties(d *Display, box layout.Box) map[string]struct{} {
	if d == nil || len(box.Animations) == 0 {
		return nil
	}

	var out map[string]struct{}
	for i := range box.Animations {
		animation := box.Animations[i]
		if animation.Name == "" || animation.Name == "none" {
			continue
		}
		frames, err := cssanimation.LayoutKeyframes(d.Keyframes, animation.Name)
		if err != nil {
			continue
		}
		for f := range frames {
			for v := range frames[f].Layout {
				if out == nil {
					out = make(map[string]struct{})
				}
				out[frames[f].Layout[v].Property] = struct{}{}
			}
		}
	}

	return out
}

// layoutAnimationOwns reports whether any of the layout properties that write
// the same used value is animated on the box.
func layoutAnimationOwns(animated map[string]struct{}, property string) bool {
	if len(animated) == 0 {
		return false
	}
	for _, name := range geometryAnimationNames(property) {
		if _, ok := animated[name]; ok {
			return true
		}
	}

	return false
}

// animationOwnsReason names the precedence skip for one geometry property.
func animationOwnsReason(property string) string {
	return fmt.Sprintf("%s: an animation owns the property, so the transition is skipped", property)
}

// duplicateIDReason names a geometry change the engine cannot address because
// two boxes share the element id.
func duplicateIDReason(property string) string {
	return fmt.Sprintf("%s: a repeated element id cannot be relaid out", property)
}

// captureGeometryChange builds one layout-length capture from the two boxes.
// changed is false when the used length is the same on both displays, which the
// caller still uses to cancel a run heading elsewhere.
//
// Width and height use the border box directly. Top and left are expressed
// against the new box's parent origin, so a parent that moves in the same
// commit still interpolates the element's visual position from the old box to
// the new one.
func captureGeometryChange(old boxRef, next *Display, current boxRef,
	property string) (TransitionCapture, bool) {
	from, to := geometryEndpoints(old, next, current, property)
	capture := TransitionCapture{
		kind:     transitionGeometry,
		geometry: geometryCapture{from: from, to: to},
	}

	return capture, math.Abs(to-from) > transitionPaintEpsilon
}

// geometryEndpoints returns the two lengths one property interpolates.
func geometryEndpoints(old boxRef, next *Display, current boxRef,
	property string) (from, to float64) {
	switch property {
	case "width":
		return old.box.W, current.box.W
	case "height":
		return old.box.H, current.box.H
	case "top":
		origin := geometryOrigin(next, current.box, true)

		return old.box.Y - origin, current.box.Y - origin
	case "left":
		origin := geometryOrigin(next, current.box, false)

		return old.box.X - origin, current.box.X - origin
	}

	return 0, 0
}

// geometryOrigin returns the parent border-box origin a top or left override is
// expressed against. A top-level box uses the page origin. The engine's used
// style is not in the display, so a parent padding or border is not part of the
// reference; the approximation is reported in the phase notes.
func geometryOrigin(d *Display, box layout.Box, vertical bool) float64 {
	if d == nil || box.ParentIndex < 0 || box.ParentIndex >= len(d.Boxes) {
		return 0
	}
	parent := d.Boxes[box.ParentIndex]
	if vertical {
		return parent.Y
	}

	return parent.X
}

// liveGeometry returns the running layout-length entry for one element and
// property, or nil.
func (r *Runtime) liveGeometry(key transitionKey) *geometryEntry {
	for i := range r.geometry {
		if r.geometry[i].key == key {
			return &r.geometry[i]
		}
	}

	return nil
}

// geometryTargetUnchanged reports that a live run already heads for the
// captured destination, so an unrelated publish must keep its clock.
func (r *Runtime) geometryTargetUnchanged(key transitionKey, to float64) bool {
	entry := r.liveGeometry(key)

	return entry != nil && math.Abs(entry.to-to) <= transitionPaintEpsilon
}

// syncGeometryQueue adds every live geometry element that is not already in
// the shared layout queue. The queue is rebuilt from the registry instead of
// requeued inside the budget loop, so one element cannot be sampled twice in
// one frame.
func (r *Runtime) syncGeometryQueue() {
	if len(r.geometry) == 0 {
		return
	}

	ids := make([]string, 0, len(r.geometry))
	for i := range r.geometry {
		entry := &r.geometry[i]
		if entry.finished || entry.key.element.id == "" {
			continue
		}
		ids = append(ids, entry.key.element.id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		r.layout.enqueue(id)
	}
}

// pruneGeometry drops the entries whose final destination already reached the
// engine. A ReturnLayout on the same frame clears the finished mark first, so a
// refused final patch is not lost.
func (r *Runtime) pruneGeometry() {
	if len(r.geometry) == 0 {
		return
	}

	kept := r.geometry[:0]
	for i := range r.geometry {
		if r.geometry[i].finished {
			continue
		}
		kept = append(kept, r.geometry[i])
	}
	r.geometry = kept
}

// returnGeometry clears the finished mark for one element after a request the
// page could not apply, so the next layout pass emits its final value again.
func (r *Runtime) returnGeometry(id string) {
	for i := range r.geometry {
		entry := &r.geometry[i]
		if entry.key.element.id == id && entry.finished {
			entry.finished = false
		}
	}
}

// geometryOverrides samples every live layout-length run on one element and
// appends the override list the engine needs.
func (r *Runtime) geometryOverrides(id string, now time.Time,
	buf []LayoutOverride) []LayoutOverride {
	for i := range r.geometry {
		entry := &r.geometry[i]
		if entry.key.element.id != id || entry.finished {
			continue
		}
		if layoutAnimationOwns(entry.animated, entry.key.property) {
			// The animation writes this used value later in the frame. Drop the
			// run rather than sampling a value that is overwritten.
			entry.finished = true
			r.stats.TransitionsCancelled++
			continue
		}

		progress, active := transition.ProgressFor(entry.record, entry.start, now)
		value := entry.to
		if active {
			value = entry.from + (entry.to-entry.from)*progress
		}
		entry.sampled = true
		if !active {
			entry.finished = true
		}
		if entry.hasApplied && math.Abs(value-entry.applied) <= geometryEpsilon {
			continue
		}
		entry.applied, entry.hasApplied = value, true
		buf = append(buf, LayoutOverride{
			Property: entry.key.property,
			Value:    cssPixels(value),
		})
	}

	return buf
}

// geometryActive reports whether a layout-length run still needs a frame.
func (r *Runtime) geometryActive() bool {
	for i := range r.geometry {
		if !r.geometry[i].finished {
			return true
		}
	}

	return false
}

// nextGeometryFrame is the next time a layout-length run needs the engine. A
// run inside its delay parks until the delay ends; a moving run asks for the
// next frame.
func (r *Runtime) nextGeometryFrame(now time.Time) (time.Time, bool) {
	var wake time.Time
	found := false
	for i := range r.geometry {
		entry := &r.geometry[i]
		if entry.finished {
			continue
		}
		at := now
		if end := entry.start.Add(milliseconds(entry.record.DelayMS)); entry.sampled && now.Before(end) {
			at = end
		}
		if !found || at.Before(wake) {
			wake, found = at, true
		}
	}

	return wake, found
}

// cssPixels formats one length in the form the engine's override relayout
// accepts. A value that rounds to zero prints as 0px rather than a negative
// zero or a long tail.
func cssPixels(value float64) string {
	if math.Abs(value) < geometryEpsilon {
		value = 0
	}

	return strconv.FormatFloat(value, 'f', -1, 64) + "px"
}
