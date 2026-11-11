package host

// Scroll is one page scroll request. Absolute reports that X and Y are a
// target offset; otherwise they are a delta.
type Scroll struct {
	X, Y     int
	Absolute bool
}

// ScrollRequester is a screen that wants the window to move its scroll
// offset.
type ScrollRequester interface {
	// TakeScroll returns the pending request and clears it. Reported is
	// false when the page has none.
	TakeScroll() (Scroll, bool)
}

// ScrollViewportObserver consumes logical viewport geometry after host zoom.
// It supplies the scroll space used by engine sticky paint and hit placement.
type ScrollViewportObserver interface {
	SetScrollViewport(x, y, width, height float64)
}
