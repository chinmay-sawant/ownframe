package csspreview

import "strings"

func transitionBehavior(name string) string {
	webkit := strings.HasPrefix(name, "-webkit-")
	name = strings.TrimPrefix(name, "-webkit-")
	if webkit {
		return webkitTransitionBehavior(name)
	}
	switch name {
	case "transition", "transition-timing-function":
		return "Moves the tile horizontally from left to right across the track."
	case "transition-property":
		return "Expands the tile from 50px to 126px wide."
	case "transition-duration":
		return "Fades the tile from 55% opacity to fully visible over 800ms."
	case "transition-delay":
		return "Changes the tile from coral to blue after a 150ms wait."
	default:
		return "Moves the tile horizontally from left to right across the track."
	}
}

func webkitTransitionBehavior(name string) string {
	switch name {
	case "transition", "transition-timing-function":
		return "Moves the tile horizontally to the right with an ease-out finish."
	case "transition-property":
		return "Expands the tile from 50px to 116px wide."
	case "transition-duration":
		return "Fades the tile from 50% to fully visible in four steps."
	case "transition-delay":
		return "Changes the tile from gold to purple, starting partway through the change."
	default:
		return "Moves the tile horizontally to the right with an ease-out finish."
	}
}
