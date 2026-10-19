package main

func standardAnimationBehavior(name string) string {
	switch name {
	case "animation":
		return "A teal capsule sweeps diagonally and turns gold at the far end."
	case "animation-delay":
		return "Waits 250ms, traces a wave, then holds the finish for 300ms."
	case "animation-delay-start":
		return "Waits 500ms before a short upward diagonal."
	case "animation-delay-end":
		return "Slides right, then waits 450ms after the last keyframe."
	case "animation-direction":
		return "Starts at the right endpoint and plays the path backward."
	case "animation-duration":
		return "A three-color diagonal sweep takes 1.6 seconds."
	case "animation-fill-mode":
		return "Moves once and keeps its turquoise endpoint after the animation ends."
	case "animation-iteration-count":
		return "Repeats a short zigzag two and a half times, then stops."
	case "animation-name":
		return "The named keyframes trace a violet V across the stage."
	case "animation-play-state":
		return "Starts paused; Start runs a quick coral sweep that fades out."
	case "animation-range":
		return "Runs a short S-curve while the entry and exit range appear in the readout."
	case "animation-range-center":
		return "Fades through three green levels with its center range in the readout."
	case "animation-range-end":
		return "Moves left from the endpoint; its end range appears in the readout."
	case "animation-range-start":
		return "Steps through a short orange path; its start range appears in the readout."
	case "animation-timeline":
		return "Runs a blue vertical sweep on the preview clock."
	case "animation-trigger":
		return "Start launches a dotted yellow sweep and shows the trigger value."
	default:
		return "Uses four distinct color stops while easing across the track."
	}
}
