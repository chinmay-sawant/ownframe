package csstrigger

import (
	"github.com/chinmay-sawant/ownframe"
)

// Event is a named host event tied to the element that produced it.
type Event struct {
	Name   string
	Source int
	Active bool
}

// Action identifies one animation action selected by a matching event.
type Action struct {
	BoxIndex, AnimationIndex int
	Name, Value              string
}

// Resolve matches a named event to CSS animation-trigger records.
func Resolve(boxes []ownframe.Box, event Event) []Action {
	if event.Source < 0 || event.Source >= len(boxes) || event.Name == "" {
		return nil
	}
	var out []Action
	for boxIndex, box := range boxes {
		if !inScope(boxes, event.Source, boxIndex, event.Name) {
			continue
		}
		for animationIndex, animation := range box.Animations {
			actions := triggerActions(animation.Trigger, event.Name)
			if len(actions) == 0 || len(actions) > 2 {
				continue
			}
			choice := 0
			if !event.Active {
				if len(actions) < 2 {
					continue
				}
				choice = 1
			}
			out = append(out, Action{
				BoxIndex:       boxIndex,
				AnimationIndex: animationIndex,
				Name:           event.Name,
				Value:          actions[choice],
			})
		}
	}
	return out
}
