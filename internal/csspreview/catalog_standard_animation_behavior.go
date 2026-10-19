package csspreview

func standardAnimationBehavior(name string) string {
	switch name {
	case "animation", "animation-name":
		return "A blue tile glides right, fades and turns orange midway, then alternates back."
	case "animation-delay":
		return "The glide waits 100ms before starting and 150ms after its final keyframe."
	case "animation-delay-start":
		return "Waits 100ms before the first glide begins."
	case "animation-delay-end":
		return "Holds for 150ms after the final keyframe before ending."
	case "animation-direction":
		return "Alternates the glide to the right with a return to the left."
	case "animation-duration":
		return "Each 180px glide takes 1.6 seconds."
	case "animation-fill-mode":
		return "Keeps the first or last keyframe appearance outside the running interval."
	case "animation-iteration-count":
		return "Runs two and a half alternating glides, then stops."
	case "animation-play-state":
		return "Starts paused. Press Start to run the glide."
	default:
		return "Eases the tile across the track instead of moving at a constant speed."
	}
}
