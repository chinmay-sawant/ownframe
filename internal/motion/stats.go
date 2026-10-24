package motion

// Stats is a snapshot of the runtime counters. The page copies it into the
// probe block; motion logs nothing on the hot path.
type Stats struct {
	// Bound is the number of players the last bind installed.
	Bound int
	// Active is the number of players that reported work at the last Step.
	Active int
	// Sampled is the number of players the last Step sampled.
	Sampled int
	// NeedsRestamp is the number of sampled players whose transform or
	// perspective could not be written as a 2D matrix at the last Step,
	// because the element or the keyframe is 3D, or because it sampled a
	// perspective. Step collects one Request per element for them and the
	// page applies them through layout.RestampTransforms.
	NeedsRestamp int
	// NeedsRetainedState reports whether the last bind found motion that needs
	// the re-projection pass. Without retained layout state the engine pass
	// cannot run, so the page turns it on and lays the page out once more
	// before the first restamp.
	NeedsRetainedState bool
	// Restamped is the number of elements the page re-projected at the last
	// tick.
	Restamped int
	// RestampError is the last re-projection failure. The page keeps drawing
	// the display it already had, so a bad override leaves the last good frame
	// on screen and the next sample tries again.
	RestampError string
	// Rebinds counts the binds that replaced at least one bound player.
	Rebinds int
	// LastError is the last bind failure. It is empty when the last bind
	// succeeded; a failed box does not stop the other boxes from binding.
	LastError string
	// Gaps lists the fidelity gaps the last bind accepted, one entry per
	// element and reason. A gap is a property or approximation the runtime
	// knows it cannot reproduce exactly yet: the declaration stays at the
	// element's base paint, or the approximation is written and reported, such
	// as per-operation opacity on overlapping content until phase 8. The list
	// is rebuilt at every bind and empty after Bind(nil).
	Gaps []string
}
