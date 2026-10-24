package motion

import (
	"sort"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/animation"
)

// defaultLayoutBudget is the number of element relayouts one frame may apply.
// A relayout is tens to hundreds of microseconds, and a page can declare many
// geometry animations, so the runtime drains one element per frame and queues
// the rest. The engine pass lives on the page, so the budget is enforced here,
// where the queue is owned.
const defaultLayoutBudget = 1

// LayoutOverride is one layout-tier property value for the engine's
// RelayoutElementWithOverrides call. Property is a CSS property name and Value
// is a CSS value the engine accepts.
type LayoutOverride struct {
	Property string
	Value    string
}

// LayoutRequest is one element's sampled layout-tier values for a frame. Every
// layout property the element animates is in one request, so a card that
// animates width and height reaches the engine's relayout path once.
type LayoutRequest struct {
	ElementID string
	Overrides []LayoutOverride
}

// layoutState is the runtime's layout-tier driver state: the element queue, the
// per-frame budget, and the buffers one sample reuses. A page with no layout
// keyframes leaves it empty and pays nothing.
type layoutState struct {
	// queue is the element ids waiting for a relayout, in the order they will
	// run. An id stays in the queue while its animation is live, so a busy
	// element shares the budget with its neighbours instead of starving them.
	queue []string
	// queued mirrors queue for the membership test. A rebind rebuilds the
	// entries but keeps the queue, so a queued element is not lost.
	queued map[string]struct{}
	// budget overrides defaultLayoutBudget when a caller sets it. Zero means
	// the default.
	budget int
	// stamp counts StepLayout calls. served records the call that last sampled
	// an element, so the reported queue depth counts what actually waited.
	stamp  uint64
	served map[string]uint64
	// requests is the slice StepLayout returns, reused by the next call.
	requests []LayoutRequest
	// values is the sample buffer one element uses. It is handed back to
	// animation.LayoutAt, which appends the sampled values to it.
	values []animation.LayoutValue
}

// reset drops every queued element and the sampled buffers. A bind that
// releases the runtime calls it, so a page with animations off pays nothing.
func (l *layoutState) reset() {
	l.queue = l.queue[:0]
	clear(l.queued)
	clear(l.served)
	l.stamp = 0
	l.requests = l.requests[:0]
	l.values = l.values[:0]
}

// enqueue adds an element to the tail of the queue when it is not already
// queued.
func (l *layoutState) enqueue(id string) {
	if id == "" || l.has(id) {
		return
	}
	if l.queued == nil {
		l.queued = make(map[string]struct{})
	}
	l.queue = append(l.queue, id)
	l.queued[id] = struct{}{}
}

// enqueueFront puts an element back at the head of the queue after a frame that
// could not apply its sample, so the element runs first on the next frame. An
// element that was requeued at the tail is moved, not duplicated.
func (l *layoutState) enqueueFront(id string) {
	if id == "" {
		return
	}
	l.remove(id)
	if l.queued == nil {
		l.queued = make(map[string]struct{})
	}
	l.queue = append(l.queue, "")
	copy(l.queue[1:], l.queue)
	l.queue[0] = id
	l.queued[id] = struct{}{}
}

// remove drops an id from the queue when it is present.
func (l *layoutState) remove(id string) {
	for i, candidate := range l.queue {
		if candidate == id {
			l.queue = append(l.queue[:i], l.queue[i+1:]...)

			break
		}
	}
	delete(l.queued, id)
}

// take removes the head of the queue.
func (l *layoutState) take() (string, bool) {
	if len(l.queue) == 0 {
		return "", false
	}
	id := l.queue[0]
	l.queue = l.queue[1:]
	delete(l.queued, id)

	return id, true
}

// has reports whether an element is queued.
func (l *layoutState) has(id string) bool {
	_, ok := l.queued[id]

	return ok
}

// queuedDepth returns the number of queued elements this frame did not sample.
// A requeued element that ran this frame is not waiting; a neighbour that did
// not run is.
func (l *layoutState) queuedDepth() int {
	depth := 0
	for _, id := range l.queue {
		if id == "" {
			continue
		}
		if l.stamp == 0 || l.served[id] != l.stamp {
			depth++
		}
	}

	return depth
}

// syncLayoutQueue adds the layout entries a bind introduced. An entry that was
// sampled before the rebind keeps its place in the queue, so a page that
// republishes mid-animation does not restart the rotation.
func (r *Runtime) syncLayoutQueue() {
	var missing []string
	bound := 0
	for i := range r.entries {
		entry := &r.entries[i]
		if !entry.layout || entry.key.id == "" {
			continue
		}
		bound++
		if entry.sampled || r.layout.has(entry.key.id) {
			continue
		}
		missing = append(missing, entry.key.id)
	}

	sort.Strings(missing)
	for i := 0; i < len(missing); i++ {
		if i > 0 && missing[i] == missing[i-1] {
			continue
		}
		r.layout.enqueue(missing[i])
	}
	r.stats.LayoutBound = bound
	r.stats.LayoutQueued = len(r.layout.queue)
}

// layoutEntry returns the entry that samples one element's layout keyframes.
func (r *Runtime) layoutEntry(id string) *entry {
	if id == "" {
		return nil
	}
	for i := range r.entries {
		entry := &r.entries[i]
		if entry.layout && entry.key.id == id {
			return entry
		}
	}

	return nil
}

// StepLayout samples the layout-tier animations and geometry transitions that
// are due this frame and returns at most the per-frame budget as engine
// override requests. Every property one element writes is in one request.
//
// An element that does not fit the budget stays queued, and it samples at the
// frame that runs it rather than at the frame it was queued, so its value
// matches the clock of the frame that applies it. A paused element samples once
// and leaves the queue; a live element goes to the back, so the budget rotates
// through the page instead of starving the later ids. A geometry transition
// joins the queue at bind and every live entry is re-added from the registry
// before sampling, so it never spins the budget loop on a sub-pixel sample.
func (r *Runtime) StepLayout(now time.Time) []LayoutRequest {
	if r == nil {
		return nil
	}

	r.layout.requests = r.layout.requests[:0]
	r.stats.LayoutRelayouts = 0
	r.layout.stamp++
	r.pruneGeometry()
	r.syncGeometryQueue()

	budget := r.layout.budget
	if budget <= 0 {
		budget = defaultLayoutBudget
	}
	for budget > 0 {
		id, ok := r.layout.take()
		if !ok {
			break
		}

		overrides := r.geometryOverrides(id, now, nil)
		values, animationLive := r.sampleLayoutAnimation(id, now)
		overrides = mergeLayoutOverrides(overrides, values)
		if len(overrides) == 0 {
			if animationLive {
				r.layout.enqueue(id)
			}

			continue
		}

		if r.layout.served == nil {
			r.layout.served = make(map[string]uint64)
		}
		r.layout.served[id] = r.layout.stamp
		r.layout.requests = append(r.layout.requests,
			LayoutRequest{ElementID: id, Overrides: overrides})
		budget--
		if animationLive {
			r.layout.enqueue(id)
		}
	}

	r.stats.LayoutQueued = r.layout.queuedDepth()
	r.stats.LayoutBudgetHit = r.stats.LayoutQueued > 0

	return r.layout.requests
}

// sampleLayoutAnimation samples one element's layout keyframe player and
// reports whether it still has work. The transition geometry pass calls it
// after its own sample so one element reaches the engine once, and the
// animation value wins when both write the same property.
func (r *Runtime) sampleLayoutAnimation(id string, now time.Time) ([]animation.LayoutValue, bool) {
	entry := r.layoutEntry(id)
	if entry == nil {
		return nil, false
	}
	if entry.sampled && entry.player.Paused() {
		// The frozen pose was applied on the frame it was sampled.
		return nil, false
	}

	input, _ := r.input(entry, now)
	values, live := entry.player.LayoutAt(input, r.layout.values[:0])
	r.layout.values = values[:0]
	entry.sampled, entry.sampledAt = true, now

	return values, live && !entry.player.Paused()
}

// mergeLayoutOverrides appends the animation values a transition does not
// already own. An animation wins the cascade on the same property, so its
// value replaces the transition's when the two collide.
func mergeLayoutOverrides(overrides []LayoutOverride,
	values []animation.LayoutValue) []LayoutOverride {
	for i := range values {
		property := values[i].Property
		value := values[i].CSS()
		replaced := false
		for j := range overrides {
			if overrides[j].Property == property {
				overrides[j].Value = value
				replaced = true

				break
			}
		}
		if !replaced {
			overrides = append(overrides, LayoutOverride{Property: property, Value: value})
		}
	}

	return overrides
}

// ReturnLayout puts the elements of a sample back at the head of the queue. The
// page calls it when the display it would patch does not carry retained layout
// state yet, so the element samples again on the frame that can apply it.
func (r *Runtime) ReturnLayout(requests []LayoutRequest) {
	if r == nil {
		return
	}
	for i := len(requests) - 1; i >= 0; i-- {
		id := requests[i].ElementID
		if id == "" {
			continue
		}
		if entry := r.layoutEntry(id); entry != nil {
			entry.sampled = false
		}
		r.returnGeometry(id)
		delete(r.layout.served, id)
		r.layout.enqueueFront(id)
	}
	r.stats.LayoutQueued = r.layout.queuedDepth()
}

// RequeueLayout marks every bound layout animation due for another sample. A
// publish rebuilds the display from the static styles, so a paused or finished
// layout animation has to run again to put its pose on the new operations; the
// next StepLayout drains the queue under the same budget.
func (r *Runtime) RequeueLayout() {
	if r == nil {
		return
	}
	for i := range r.entries {
		entry := &r.entries[i]
		if !entry.layout || entry.key.id == "" {
			continue
		}
		entry.sampled = false
		r.layout.enqueue(entry.key.id)
	}
	r.stats.LayoutQueued = len(r.layout.queue)
}

// NoteLayout records the outcome of the page's relayout pass. applied is the
// number of elements the engine patched; a failure keeps the display the page
// already draws and is reported through Stats.LayoutError.
func (r *Runtime) NoteLayout(applied int, err error) {
	if r == nil {
		return
	}

	r.stats.LayoutRelayouts = applied
	if err != nil {
		r.stats.LayoutError = err.Error()

		return
	}

	r.stats.LayoutError = ""
}
