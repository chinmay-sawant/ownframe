package window

// pressEdge reports the press this frame. A press and release that both land
// between two Update ticks leave only the release level, so a fast click
// during a long frame would otherwise reach the page as a release with no
// press. A release with no press this shell saw is therefore the remainder of
// a real click, and the release point is where to place it. A held button
// whose press was seen keeps wasDown true and never fires twice.
func pressEdge(down, released, wasDown bool) bool {
	if pressedNow(down, wasDown) {
		return true
	}

	return released && !down && !wasDown
}

// pressedNow reports the up-to-down edge of a pointer level.
func pressedNow(down, wasDown bool) bool {
	return down && !wasDown
}
