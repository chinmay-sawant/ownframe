package host

// ViewLock is a screen that fills the window.
// The window scales it to the screen. It does not scroll or pinch-zoom.
type ViewLock interface {
	ViewLocked() bool
}

// ViewportSizer accepts the available layout size independently of native
// window size limits. Responsive documents use it when a dock or DPI change
// leaves less space than the configured desktop minimum.
type ViewportSizer interface {
	SetViewport(width, height int)
}
