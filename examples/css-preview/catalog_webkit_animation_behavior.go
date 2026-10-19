package main

func webkitAnimationBehavior(name string) string {
	switch name {
	case "animation":
		return "A purple capsule sweeps diagonally and flashes pink at the finish."
	case "animation-delay":
		return "Waits 350ms, then traces a teal zigzag across the stage."
	case "animation-direction":
		return "Starts at the far end and travels backward because direction is reverse."
	case "animation-duration":
		return "A gold diagonal sweep takes 1.2 seconds."
	case "animation-fill-mode":
		return "Moves once to the right and keeps its coral finish."
	case "animation-iteration-count":
		return "Crosses a three-color path three times, then stops."
	case "animation-name":
		return "The WebKit keyframe name makes a green V-shaped trace."
	case "animation-play-state":
		return "Starts paused; Start runs a short violet zigzag."
	case "animation-timing-function":
		return "Jumps through four blue positions at equal steps."
	default:
		return "Moves on a curved path and shifts from teal to pink."
	}
}
