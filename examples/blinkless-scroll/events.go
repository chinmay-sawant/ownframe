package main

import "strings"

// EventTrigger is one event-trigger: named start events turn it on and named
// end events turn it off. An event listed in both lists toggles it.
type EventTrigger struct {
	Name  string
	Start []string
	End   []string
}

// EventTriggers holds the on/off state of a set of event triggers.
type EventTriggers struct {
	defs   []EventTrigger
	active map[string]bool
}

// NewEventTriggers builds a set with every trigger off.
func NewEventTriggers(defs []EventTrigger) *EventTriggers {
	return &EventTriggers{defs: defs, active: map[string]bool{}}
}

// Active reports one trigger's state.
func (e *EventTriggers) Active(name string) bool { return e.active[name] }

// Dispatch feeds one event and returns the names whose state changed.
func (e *EventTriggers) Dispatch(event string) []string {
	var changed []string

	for _, d := range e.defs {
		start, end := has(d.Start, event), has(d.End, event)
		next := e.active[d.Name]

		switch {
		case start && end:
			next = !next
		case start:
			next = true
		case end:
			next = false
		default:
			continue
		}

		if next != e.active[d.Name] {
			e.active[d.Name] = next
			changed = append(changed, d.Name)
		}
	}

	return changed
}

func has(list []string, s string) bool {
	for _, v := range list {
		if v == s || v == eventBase(s) {
			return true
		}
	}

	return false
}

// eventBase strips the detail from an event name: keydown(Enter) is keydown.
// A trigger that lists plain keydown fires for any key; one that lists
// keydown(Enter) fires only for that key.
func eventBase(event string) string {
	if i := strings.IndexByte(event, '('); i > 0 {
		return event[:i]
	}

	return event
}
