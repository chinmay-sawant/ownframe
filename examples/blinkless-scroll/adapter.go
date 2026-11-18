package main

import "github.com/chinmay-sawant/blinkless/layout"

// Model is everything the host needs from one display list.
type Model struct {
	Nodes    []Node
	Triggers []OwnedTrigger
	Events   []OwnedEvent
	Delays   map[int]InterestDelay
}

// Build reads the engine's boxes into host types. Inert comes from
// Display.BoxInert, which already walks the ancestors. An interest delay
// inherits, so only a box whose delay differs from its parent's gets a
// tracker; hovering a descendant still counts as hovering that box.
func Build(d *layout.Display) Model {
	m := Model{Nodes: make([]Node, len(d.Boxes)), Delays: map[int]InterestDelay{}}

	for i, b := range d.Boxes {
		n := Node{
			ID:     b.ID,
			Parent: b.ParentIndex,
			Rect:   Rect{X: b.X, Y: b.Y, W: b.W, H: b.H},
			Inert:  d.BoxInert(i),
			Scope:  scopeFrom(b.TimelineScope),
		}

		for _, tl := range b.Timelines {
			axis := Axis(tl.Axis)
			if axis != AxisX {
				axis = AxisY
			}

			n.Timelines = append(n.Timelines, Timeline{
				Name: tl.Name, Kind: tl.Kind, Axis: axis, SubjectID: tl.SubjectID,
				InsetStart: tl.InsetStart, InsetEnd: tl.InsetEnd,
				ScrollerID: tl.ScrollerID, Anonymous: tl.ScrollerAnonymous,
			})
		}

		m.Nodes[i] = n

		for _, t := range b.TimelineTriggers {
			m.Triggers = append(m.Triggers, OwnedTrigger{Owner: i, TimelineTrigger: TimelineTrigger{
				Name: t.Name, Source: t.Source, Sources: t.Sources,
				ActivationStart: boundFrom(t.ActivationStart), ActivationEnd: boundFrom(t.ActivationEnd),
				ActiveStart: boundFrom(t.ActiveStart), ActiveEnd: boundFrom(t.ActiveEnd),
			}})
		}

		for _, e := range b.EventTriggers {
			m.Events = append(m.Events, OwnedEvent{Owner: i, EventTrigger: EventTrigger{Name: e.Name, Start: e.Start, End: e.End}})
		}

		if b.ParentIndex >= 0 && b.ParentIndex < len(d.Boxes) && b.InterestDelay != d.Boxes[b.ParentIndex].InterestDelay {
			m.Delays[i] = InterestDelay(b.InterestDelay)
		}
	}

	return m
}

func scopeFrom(s layout.TimelineScope) Scope {
	switch {
	case s.All:
		return Scope{Kind: ScopeAll}
	case len(s.Names) > 0:
		return Scope{Kind: ScopeNames, Names: s.Names}
	}

	return Scope{}
}

func boundFrom(b layout.RangeBound) RangeBound {
	return RangeBound{Normal: b.Normal, Auto: b.Auto, Name: b.Name, Percent: b.Percent, Length: b.Length,
		HasPercent: b.HasPercent, HasLength: b.HasLength}
}
