package csspreview

func webkitAnimationBehavior(name string) string {
	switch name {
	case "animation", "animation-name":
		return "A purple tile hops upward, fades and changes color, then drops back down."
	case "animation-delay":
		return "The hop starts without an initial wait."
	case "animation-direction":
		return "Alternates the hop upward and back down."
	case "animation-duration":
		return "Each hop takes 1.2 seconds."
	case "animation-fill-mode":
		return "Keeps the first or last keyframe appearance outside the running interval."
	case "animation-iteration-count":
		return "Repeats the hop continuously."
	case "animation-play-state":
		return "Starts paused. Press Start to repeat the hop."
	default:
		return "Eases the tile through its upward and downward hop."
	}
}
