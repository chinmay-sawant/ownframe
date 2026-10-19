package main

import "strings"

func transitionBehavior(name string) string {
	webkit := strings.HasPrefix(name, "-webkit-")
	name = strings.TrimPrefix(name, "-webkit-")
	if webkit {
		return webkitTransitionBehavior(name)
	}
	switch name {
	case "transition":
		return "Slides the coral capsule right across the track."
	case "transition-property":
		return "Expands the lime tile from 45px to 134px wide."
	case "transition-duration":
		return "Fades the blue tile from 25% to fully visible over 1.1 seconds."
	case "transition-delay":
		return "Changes the orange tile to teal after a 400ms pause."
	case "transition-timing-function":
		return "Moves in four visible jumps using a stepped timing curve."
	default:
		return "Changes the violet tile to mint with allow-discrete behavior."
	}
}

func webkitTransitionBehavior(name string) string {
	switch name {
	case "transition":
		return "Slides the purple capsule right with a fast finish."
	case "transition-property":
		return "Expands the gold tile from 46px to 104px in five steps."
	case "transition-duration":
		return "Fades the plum tile from 20% to full opacity over 1.2 seconds."
	case "transition-delay":
		return "Changes the gold tile to purple after a 350ms pause."
	case "transition-timing-function":
		return "Slides 180px with a cubic timing curve."
	default:
		return "Slides with a custom easing curve and a short overshoot."
	}
}
