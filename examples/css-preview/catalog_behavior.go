package main

import "strings"

func catalogBehavior(name string) string {
	switch {
	case name == "animation-composition":
		return "A magenta bead sweeps diagonally; the readout shows add composition."
	case strings.Contains(name, "range"), strings.Contains(name, "timeline"), name == "animation-trigger":
		return animationMetadataBehavior(name)
	case name == "transition-behavior":
		return "The tile changes from coral to blue; the row shows the allow-discrete setting."
	case strings.HasPrefix(name, "view-transition") || name == "trigger-scope":
		return otherMetadataBehavior(name)
	case strings.Contains(name, "animation"):
		return animationBehavior(name)
	case strings.Contains(name, "transition"):
		return transitionBehavior(name)
	default:
		return "The preview plays the CSS sample shown on this card."
	}
}
