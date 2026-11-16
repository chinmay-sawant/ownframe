package main

import "time"

// Host is the window-free engine: it owns the nodes, the trigger memory, and
// the interest timers, and turns a viewport plus a pointer position into one
// Frame of numbers. Nothing here touches a window or the page.

// OwnedTrigger is a timeline trigger declared by the node at Owner.
type OwnedTrigger struct {
	Owner int
	TimelineTrigger
}

// OwnedEvent is an event trigger declared by the node at Owner. It fires only
// for events whose target is Owner or one of its descendants.
type OwnedEvent struct {
	Owner int
	EventTrigger
}

type tlKey struct {
	Node int
	Name string
}

// Value is one timeline's reading in a frame.
type Value struct {
	Kind     string
	Progress float64
	Pos      float64
	Space    RangeSpace
	// Ranges holds the progress for cover, entry, exit and contain of a view
	// timeline. It is nil for the other kinds.
	Ranges map[string]float64
	// Inside is only meaningful for a pointer timeline.
	Inside bool
}

// Frame is everything the host decided for one tick.
type Frame struct {
	Timelines  map[tlKey]Value
	TriggerOn  []bool // parallel to Host.Triggers
	EventOn    []bool // parallel to Host.Events
	InterestOn map[int]bool
}

// Host holds the engine state between frames.
type Host struct {
	Nodes    []Node
	Triggers []OwnedTrigger
	Events   []OwnedEvent

	// clicked is the node the latest click landed on. It is only valid while
	// hasClicked is true, and it is what Styles outlines so a click is visible.
	clicked    int
	hasClicked bool

	byID     map[string]int
	tstate   []TriggerState
	events   []*EventTriggers
	interest map[int]*Interest
}

// NewHost builds a host over nodes. delays maps a node index to its
// interest-delay; clock may be nil for the wall clock.
func NewHost(nodes []Node, triggers []OwnedTrigger, events []OwnedEvent, delays map[int]InterestDelay, clock Clock) *Host {
	h := &Host{
		Nodes:    nodes,
		Triggers: triggers,
		Events:   events,
		byID:     map[string]int{},
		tstate:   make([]TriggerState, len(triggers)),
		interest: map[int]*Interest{},
	}

	for i, n := range nodes {
		if n.ID != "" {
			h.byID[n.ID] = i
		}
	}

	for _, e := range events {
		h.events = append(h.events, NewEventTriggers([]EventTrigger{e.EventTrigger}))
	}

	for idx, d := range delays {
		h.interest[idx] = NewInterest(d, clock)
	}

	return h
}

// IndexOf returns the node index for an element id.
func (h *Host) IndexOf(id string) (int, bool) {
	i, ok := h.byID[id]

	return i, ok
}

// SetClicked remembers which node the latest click landed on, or -1 when it
// landed on nothing a pointer can reach. Styles outlines that node.
func (h *Host) SetClicked(i int) { h.clicked, h.hasClicked = i, i >= 0 }

// Hover tells the interest trackers where the pointer is. Inert nodes see no
// pointer, so an inert card never gains interest.
func (h *Host) Hover(px, py float64) {
	hit := HitTest(h.Nodes, px, py)

	for idx, in := range h.interest {
		in.Hover(hit >= 0 && (hit == idx || isAncestor(h.Nodes, idx, hit)))
	}
}

// Dispatch feeds a named event (click, keydown) aimed at the point and
// returns the indexes of the event triggers whose state changed. An inert
// target takes no event.
func (h *Host) Dispatch(event string, px, py float64) []int {
	hit := HitTest(h.Nodes, px, py)
	if hit < 0 {
		return nil
	}

	var changed []int

	for i, e := range h.Events {
		if hit != e.Owner && !isAncestor(h.Nodes, e.Owner, hit) {
			continue
		}

		if len(h.events[i].Dispatch(event)) > 0 {
			changed = append(changed, i)
		}
	}

	return changed
}

// Compute reads every timeline, then moves every trigger and timer.
func (h *Host) Compute(v Viewport, px, py float64) Frame {
	f := Frame{
		Timelines:  map[tlKey]Value{},
		TriggerOn:  make([]bool, len(h.Triggers)),
		EventOn:    make([]bool, len(h.Events)),
		InterestOn: map[int]bool{},
	}

	for i, n := range h.Nodes {
		for _, tl := range n.Timelines {
			f.Timelines[tlKey{i, tl.Name}] = h.read(i, tl, v, px, py)
		}
	}

	for i, t := range h.Triggers {
		pos, space, ok := h.sourceValue(t, v, f)
		if !ok {
			continue
		}

		f.TriggerOn[i] = t.Evaluate(&h.tstate[i], pos, space)
	}

	for i := range h.Events {
		f.EventOn[i] = h.events[i].Active(h.Events[i].Name)
	}

	for idx, in := range h.interest {
		f.InterestOn[idx] = in.Tick()
	}

	return f
}

func (h *Host) read(i int, tl Timeline, v Viewport, px, py float64) Value {
	switch tl.Kind {
	case KindScroll:
		pos, space := ScrollSpace(v, tl.Axis)

		return Value{Kind: tl.Kind, Progress: ScrollProgress(v, tl.Axis), Pos: pos, Space: space}
	case KindView:
		subject := h.Nodes[i].Rect
		if j, ok := h.byID[tl.SubjectID]; ok && tl.SubjectID != "" {
			subject = h.Nodes[j].Rect
		}

		st := ViewTimeline(v, subject, tl.Axis, tl.InsetStart, tl.InsetEnd)
		ranges := map[string]float64{}

		for name := range st.Space.Named {
			ranges[name] = st.Progress(name)
		}

		return Value{Kind: tl.Kind, Progress: st.Progress("cover"), Pos: st.T, Space: st.Space, Ranges: ranges}
	case KindPointer:
		p, in := PointerProgress(px, py, h.Nodes[i].Rect, tl.Axis)
		span := h.Nodes[i].Rect.size(tl.Axis)

		// Inside on the axis is not enough: a pointer below an x-axis box
		// has the same x as the box, and must not drive it.
		in = in && contains(h.Nodes[i].Rect, px, py)

		if !in {
			p = 0 // the pointer timeline is idle while the pointer is elsewhere
		}

		return Value{Kind: tl.Kind, Progress: p, Pos: p * span,
			Space: RangeSpace{Total: span, Named: map[string][2]float64{"cover": {0, span}}}, Inside: in}
	}

	return Value{Kind: tl.Kind}
}

// sourceValue finds the position and ruler a trigger watches. The first entry
// in its source list that resolves wins. Anonymous sources mean:
//
//   - "scroll()" and "auto": the page scroll timeline on the y axis
//   - "view()": a view timeline on the trigger's own box, y axis, no insets
//   - "none": no timeline, so the trigger never fires
//   - a dashed name: the timeline that name resolves to from the trigger's
//     box (declared there or on an ancestor, or hoisted by timeline-scope)
func (h *Host) sourceValue(t OwnedTrigger, v Viewport, f Frame) (float64, RangeSpace, bool) {
	sources := t.Sources
	if len(sources) == 0 && t.Source != "" {
		sources = []string{t.Source}
	}

	for _, src := range sources {
		switch src {
		case "none":
			continue
		case "scroll()", "auto":
			pos, space := ScrollSpace(v, AxisY)

			return pos, space, true
		case "view()":
			st := ViewTimeline(v, h.Nodes[t.Owner].Rect, AxisY, 0, 0)

			return st.T, st.Space, true
		}

		tl, at, ok := ResolveTimeline(h.Nodes, t.Owner, src)
		if val, found := f.Timelines[tlKey{at, tl.Name}]; ok && found {
			return val.Pos, val.Space, true
		}
	}

	return 0, RangeSpace{}, false
}

// NextDeadline is the earliest time an interest timer fires, so a window can
// wake for it instead of waiting for input.
func (h *Host) NextDeadline() (time.Time, bool) {
	var best time.Time

	found := false

	for _, in := range h.interest {
		if at, ok := in.NextDeadline(); ok && (!found || at.Before(best)) {
			best, found = at, true
		}
	}

	return best, found
}
